package main

import (
	"io"
	"mime"
	"net"
	"net/http"
	"path/filepath"
	"sync"
	"time"

	"crypto/rand"
	"encoding/base64"
	"encoding/json"

	"github.com/gorilla/mux"
	"github.com/hashicorp/yamux"
	log "github.com/sirupsen/logrus"
)

// a proxy-server that accepts:
//   - TCP connection
//        - creates a session ID for each new TCP connection
//        - creates a mux and is ready to proxy all HTTP connectios of type:  host/<session id>/<path> over this TCP connection
//   - HTTP connections
//        - GET /s/<ID>/   - forwards this GET connection to the TCP connection
//        - GET /ws/<DI>/  - forwards this WS connection to the TCP connection

type HelloClient struct {
	Version string
	Data    string
}

type HelloServer struct {
	Version   string
	SessionID string
	PublicURL string
	Data      string
}

type sessionWrapper struct {
	ySession  *yamux.Session
	sessionID string
}

type serverConfig struct {
	publicURL          string
	backListenAddress  string
	frontListenAddress string
}
type server struct {
	config               serverConfig
	httpServer           *http.Server
	activeSessions       map[string]*sessionWrapper
	activeSessionsRWLock sync.RWMutex
	backListener         net.Listener
}

func errToString(err error) string {
	if err != nil {
		return err.Error()
	}
	return "nil"
}

func generateNewSessionID() string {
	binID := make([]byte, 50)
	_, err := rand.Read(binID)

	if err != nil {
		log.Fatalf(err.Error())
	}

	return base64.RawURLEncoding.EncodeToString([]byte(binID))
}

func pipeConnectionsAndWait(backConn, frontConn net.Conn) error {
	errChan := make(chan error, 2)

	backConnAddr := backConn.RemoteAddr().String()
	frontConnAddr := frontConn.RemoteAddr().String()

	log.Debugf("Piping the two conn %s <-> %s ..", backConnAddr, frontConnAddr)

	copyAndNotify := func(dst, src net.Conn, info string) {
		n, err := io.Copy(dst, src)
		log.Debugf("%s: piping done with %d bytes, and err %s", info, n, errToString(err))
		errChan <- err

		// Close both connections when done with copying. Yeah, both will beclosed two
		// times, but it doesn't matter. By closing them both, we unblock the other copy
		// call which would block indefinitely otherwise
		dst.Close()
		src.Close()
	}

	go copyAndNotify(backConn, frontConn, "front->back")
	go copyAndNotify(frontConn, backConn, "back->front")
	err1 := <-errChan
	err2 := <-errChan

	log.Debugf("Piping finished for %s <-> %s .", backConnAddr, frontConnAddr)

	// Return one of the two error that is not nil
	if err1 != nil {
		return err1
	}
	return err2
}

func mainHandler(w http.ResponseWriter, r *http.Request, backConn net.Conn) {
	// write this request further to the connection
	// hijack the connection comming with the request
	// hook the two one to the other
	defer backConn.Close()

	hj, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Can't proxy this connection", http.StatusInternalServerError)
		return
	}

	frontConn, _, err := hj.Hijack()
	// TODO: what about the buffer - might have data inside?
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer frontConn.Close()

	err = r.Write(backConn)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pipeConnectionsAndWait(backConn, frontConn)
	log.Debugf("Done with proxying.")
}

func performHandshake(conn net.Conn, publicBaseURL string, sessionID string) error {
	var helloC HelloClient
	dec := json.NewDecoder(conn)
	err := dec.Decode(&helloC)
	if err != nil {
		return err
	}

	helloS := HelloServer{
		Version:   "1",
		SessionID: sessionID,
		PublicURL: publicBaseURL + "/s/" + sessionID + "/",
		Data:      "",
	}
	enc := json.NewEncoder(conn)
	err = enc.Encode(helloS)

	log.Debugf("Handshake: client version=%s. Sent session %s", helloC.Version, helloS.SessionID)

	return err
}

func newServer(c serverConfig) *server {
	s := &server{
		config: c,
	}
	s.activeSessions = make(map[string]*sessionWrapper)
	return s
}

func addNewSession(s *server, session *sessionWrapper) {
	s.activeSessionsRWLock.Lock()
	s.activeSessions[session.sessionID] = session
	s.activeSessionsRWLock.Unlock()
}

func removeSession(s *server, session *sessionWrapper) {
	s.activeSessionsRWLock.Lock()
	delete(s.activeSessions, session.sessionID)
	s.activeSessionsRWLock.Unlock()
}

func getSession(s *server, sessionID string) (wrapper *sessionWrapper) {
	s.activeSessionsRWLock.RLock()
	wrapper = s.activeSessions[sessionID]
	s.activeSessionsRWLock.RUnlock()
	return
}

func (s *server) serveContent(w http.ResponseWriter, r *http.Request, name string) {
	file, err := Asset(name)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	ctype := mime.TypeByExtension(filepath.Ext(name))
	if ctype == "" {
		ctype = http.DetectContentType(file)
	}
	w.Header().Set("Content-Type", ctype)
	w.Write(file)
}

func (s *server) handleFrontConnections() error {
	router := mux.NewRouter()
	router.PathPrefix("/s/{sessionID}/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		log.Infof("New front client connection: %s, from %s", r.URL.Path, r.RemoteAddr)
		vars := mux.Vars(r)
		sessionID := vars["sessionID"]
		wrapper := getSession(s, sessionID)
		if wrapper == nil {
			log.Warnf("Invalid session: %s, from %s", sessionID, r.RemoteAddr)
			s.serveContent(w, r, "invalid-session.html")
			return
		}

		backConn, err := wrapper.ySession.Open()
		defer backConn.Close()
		if err != nil {
			log.Warnf("Cannot serve session %s for %s, back request error: ", sessionID, r.RemoteAddr, err.Error())
			s.serveContent(w, r, "invalid-session.html")
			return
		}

		mainHandler(w, r, backConn)
		duration := time.Now().Sub(startTime)
		log.Infof("Front client request %s from %s proxied for %.2f sec", r.URL.Path, r.RemoteAddr, duration.Seconds())
	})

	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			s.serveContent(w, r, r.URL.Path)
		})))

	router.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.serveContent(w, r, "404.html")
	})

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, s.config.publicURL, http.StatusMovedPermanently)
	})

	s.httpServer = &http.Server{
		Addr:    s.config.frontListenAddress,
		Handler: router,
	}
	log.Debugf("Listening for HTTP on: http://%s", s.config.frontListenAddress)
	return s.httpServer.ListenAndServe()
}

func (s *server) handleBackConnections() (err error) {
	s.backListener, err = net.Listen("tcp", s.config.backListenAddress)
	if err != nil {
		log.Fatalf(err.Error())
	}

	log.Debugf("Listening for proxy-client connections on %s", s.config.backListenAddress)

	for {
		conn, err := s.backListener.Accept()
		if err != nil {
			log.Error(err.Error())
			break
		}
		log.Debugf("New back connection: %s", conn.RemoteAddr().String())

		sessionID := generateNewSessionID()

		err = performHandshake(conn, s.config.publicURL, sessionID)

		if err != nil {
			log.Warn("Cannot perform handshake on the back connection: %s", err.Error())
			conn.Close()
			continue

		}

		ymuxServerSession, err := yamux.Client(conn, nil)

		if err != nil {
			log.Warn("Cannot create back tunnel: %s", err.Error())
			conn.Close()
			continue
		}

		wrapper := &sessionWrapper{
			sessionID: sessionID,
			ySession:  ymuxServerSession,
		}

		addNewSession(s, wrapper)
		log.Infof("New tty-share session %s from for %s", sessionID, conn.RemoteAddr().String())

		// Open a "monitoring" connection to detect when the underlying TCP connection
		// dies. Note that this "monitoring" connection is a virtual one, over the
		// multiplexer, so it should have minimal resources impact.
		go func() {
			_, err := ymuxServerSession.Accept()
			if err != nil {
				log.Infof("tty-share session %s closed: %s", sessionID, err.Error())
				removeSession(s, wrapper)
				conn.Close()
				return
			}
		}()

	}
	return
}

func (s *server) Run() (err error) {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		err = s.handleBackConnections()
		wg.Done()
	}()
	err2 := s.handleFrontConnections()
	wg.Wait()

	if err2 != nil {
		return err2
	}

	return
}

func (s *server) Stop() {
	s.httpServer.Close()
	s.backListener.Close()
}

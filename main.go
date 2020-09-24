package main

import (
	"flag"
	"os"
	"os/signal"

	log "github.com/sirupsen/logrus"
)

func main() {
	frontListenAddress := flag.String("front-address", "localhost:8080", "The front address to listen for the http requests")
	backListenAddress := flag.String("back-address", "localhost:3456", "The back address to listen for reverse proxy connections")
	url := flag.String("url", "http://localhost:8080", "The public URL this can be accessed over, to be sent back to tty-share")
	verbose := flag.Bool("verbose", false, "Verbose logging")
	flag.Parse()

	// Log setup
	log.SetLevel(log.InfoLevel)
	if *verbose {
		log.SetLevel(log.DebugLevel)
	}

	config := serverConfig{
		backListenAddress:  *backListenAddress,
		frontListenAddress: *frontListenAddress,
		publicURL:          *url,
	}

	server := newServer(config)

	// Install a signal and wait until we get Ctrl-C
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		s := <-c
		log.Debug("Received signal <", s, ">. Stopping the server")
		server.Stop()
	}()

	log.Info("Listening on address: http://", *frontListenAddress, ", and TCP://", *backListenAddress)
	err := server.Run()
	log.Debug("Exiting. Error: ", err)
}

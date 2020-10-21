# tty-proxy

## Building the gobindata.go file

All files under ~assets/*~ are packed to the gobindata.go file which will be statically compiled
within the final binary.

```bash
	go get github.com/go-bindata/go-bindata/...
	go-bindata --prefix static -o gobindata.go static/*
```

## Docker

The `tty-proxy` can be built into a docker image as follows:

    docker build -t tty-proxy .

To run the container, type:p

    docker run \
      -p 3456:3456 -p 8080:8080 \
      -e URL=http://localhost:8080 \
      --cap-drop=all --rm \
      tty-proxy

where you can replace `URL` by whatever will be the publicly visible URL of the server.

After this, clients can be connected as follows:

    tty-share --tty-proxy localhost:3456 --no-tls --public

In the above command, `:3456` is the default port where `tty-proxy` listens for incoming back
connections (i.e. `tty-share` clients), and 5000 is the port of the web interface through which
remote users can connect. You can override the defaults by specifying a different port mapping on
the command line, e.g.  `-p 4567:3456 -p 80:8080` to listen on `4567` and serve on `80`.

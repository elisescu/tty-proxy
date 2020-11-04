# tty-proxy

[![Build Status](https://travis-ci.com/elisescu/tty-proxy.svg?token=N6UZN7xxNqNmweAn6y5D&branch=master)](https://travis-ci.com/elisescu/tty-proxy)

This is the public facing service that allows `tty-share` command to create public sessions, in
addition to local ones.

`tty-proxy` will listen to the address passed by the `--back-address` flag, and any connections to
this address from `tty-share` will create a new session (`<session-id>` that will be used to proxy
any requests from any url path of the form `/s/<session-id>/` back over the corresponding
`tty-share` connection. See more documentation on the
[tty-share](https://github.com/elisescu/tty-share) project.

**Note** `tty-proxy` replaces a part of the old `tty-server` which has moved inside the actual
`tty-share` command itself. Read more
[here](https://github.com/elisescu/tty-share/wiki/tty-share-V2)

## Building

### Build the gobindata.go file

All files under `assets/*` are packed to the `gobindata.go` file which will be statically compiled
within the final binary.

```bash
	go get github.com/go-bindata/go-bindata/...
	go-bindata --prefix static -o gobindata.go static/*
```

### Build the final `tty-proxy` binary

```bash
go build
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

## nginx

Take a look at [this snippet](doc/nginx.conf) to see how I configured my nginx installation for TLS termination.

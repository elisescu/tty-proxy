FROM alpine:3.12

ENV URL=http://localhost:8080

ARG build_deps="go git"
ARG runtime_deps="dumb-init"
ARG user_id=1000

COPY . /go/src/github.com/elisescu/tty-proxy

RUN apk update && \
    apk add -u $build_deps $runtime_deps && \
    adduser -D -H -h / -u $user_id tty-proxy


RUN cd /go/src/github.com/elisescu/tty-proxy && \
    GOPATH=/go go get github.com/go-bindata/go-bindata/... && \
    GOPATH=/go /go/bin/go-bindata --prefix static -o gobindata.go static/* && \
    GOPATH=/go go build && \
    cp tty-proxy /usr/bin/ && \
    rm -r /go && \
    apk del $build_deps

EXPOSE 8080
EXPOSE 3456
USER tty-proxy

ENTRYPOINT ["/usr/bin/dumb-init", "--"]
CMD ["/bin/sh", "-c", "/usr/bin/tty-proxy --front-address :8080 --back-address :3456 -url $URL"]

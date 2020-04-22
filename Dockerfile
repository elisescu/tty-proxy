FROM alpine:3.11

ENV URL=http://localhost:5000

ARG build_deps="go make dep"
ARG runtime_deps="dumb-init"
ARG user_id=1000

COPY . /go/src/github.com/elisescu/tty-server

RUN apk update && \
    apk add -u $build_deps $runtime_deps && \
    adduser -D -H -h / -u $user_id tty-server && \
    cd /go/src/github.com/elisescu/tty-server && \
    GOPATH=/go dep ensure && \
    GOPATH=/go make all && \
    cp tty-server /usr/bin/ && \
    rm -r /go && \
    apk del $build_deps

EXPOSE 5000
EXPOSE 6543
USER tty-server

ENTRYPOINT ["/usr/bin/dumb-init", "--"]
CMD ["/bin/sh", "-c", "/usr/bin/tty-server -web_address :5000 --sender_address :6543 -url $URL"]

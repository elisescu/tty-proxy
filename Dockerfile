# builder image:
FROM alpine:3.12 as build
ARG build_deps="go git"
ARG runtime_deps="dumb-init"

COPY . /go/src/github.com/elisescu/tty-proxy
RUN apk update && \
    apk add -u $build_deps $runtime_deps && \
    cd /go/src/github.com/elisescu/tty-proxy && \
    GOPATH=/go go get github.com/go-bindata/go-bindata/... && \
    GOPATH=/go /go/bin/go-bindata --prefix static -o gobindata.go static/* && \
    GOPATH=/go go build && \
    cp tty-proxy /usr/bin/ && \
    rm -r /go && \
    apk del $build_deps

# runtime image:
FROM alpine:3.12
ARG user_id=1000
RUN adduser -D -H -h /home/tty-proxy/ -u $user_id tty-proxy

EXPOSE 8080
EXPOSE 3456
USER tty-proxy
ENV URL=http://localhost:8080
ENV FRONT_ADDRESS=:9000
ENV BACK_ADDRESS=:3456

ENTRYPOINT ["/usr/bin/dumb-init", "--"]
CMD ["/bin/sh", "-c", "/usr/bin/tty-proxy --front-address $FRONT_ADDRESS --back-address $BACK_ADDRESS -url $URL"]

COPY --from=build /usr/bin/dumb-init /usr/bin/tty-proxy /usr/bin/

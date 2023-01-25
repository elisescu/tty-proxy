FROM golang:1.19 AS build
ENV GO111MODULE=on
WORKDIR /go/src/app
COPY . .
RUN mkdir -p /build
ENV CGO_ENABLED=0
RUN go build -ldflags="-s -w" -o=/tty-proxy

FROM alpine:latest
# Timezone = Tokyo
RUN apk --no-cache add tzdata && \
    cp /usr/share/zoneinfo/Asia/Tokyo /etc/localtime

COPY --from=build /tty-proxy /tty-proxy
RUN chmod u+x /tty-proxy

ARG user_id=1000
RUN adduser -D -H -h / -u $user_id tty-proxy
USER tty-proxy

ENV URL=http://localhost:8080
EXPOSE 8080
EXPOSE 3456
CMD ["/bin/sh", "-c", "/tty-proxy --front-address :8080 --back-address :3456 -url $URL"]

FROM golang:1.19 AS build-env
ENV GO111MODULE=on
WORKDIR /go/src/app
COPY . .
RUN mkdir -p /build
ENV CGO_ENABLED=0
RUN go build -ldflags="-s -w" -o=/build/app .

FROM alpine:latest
# Timezone = Tokyo
RUN apk --no-cache add tzdata && \
    cp /usr/share/zoneinfo/Asia/Tokyo /etc/localtime

ARG user_id=1000
RUN adduser -D -H -h / -u $user_id tty-proxy

USER tty-proxy
COPY --from=build-env /build/app /build/app
RUN chmod u+x /build/app

EXPOSE 8080
EXPOSE 3456
CMD /build/app

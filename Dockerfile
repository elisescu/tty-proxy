FROM golang:1.19 AS build-env
ENV GO111MODULE=on
WORKDIR /go/src/app
COPY . .
RUN mkdir -p /build
RUN go build -ldflags="-s -w" -o=/build/app .

FROM alpine:latest
# Timezone = Tokyo
RUN apk --no-cache add tzdata && \
    cp /usr/share/zoneinfo/Asia/Tokyo /etc/localtime

COPY --from=build-env /build/app /build/app
RUN chmod u+x /build/app

CMD /build/app

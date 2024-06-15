FROM golang:1-alpine as builder

RUN apk update && apk add git build-base

WORKDIR /go/src/github.com/nevkontakte/pat
COPY . .
RUN go mod verify
RUN go build -v .

FROM alpine:latest
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
COPY --from=builder /go/src/github.com/nevkontakte/pat /srv
WORKDIR /srv/
EXPOSE 8080
ENTRYPOINT ["/srv/pat", "-bind=:8080"]
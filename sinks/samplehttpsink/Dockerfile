FROM golang:1.9-alpine

RUN apk add --no-cache git
COPY server.go .
RUN go get -v -d ./...

RUN go build -o httpsink
ENTRYPOINT ./httpsink

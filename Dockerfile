FROM golang:alpine

RUN apk update && apk add --no-cache git

WORKDIR /root
RUN git clone https://github.com/twentyfourbytes/api

WORKDIR  /root/api

RUN GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /go/bin/api


ENTRYPOINT ["/go/bin/api"]

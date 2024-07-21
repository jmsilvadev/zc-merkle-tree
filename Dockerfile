FROM golang:1.22-alpine

RUN mkdir /zc
RUN mkdir /zc/pkg
RUN mkdir /zc/cmd
RUN apk update && apk add --upgrade git openssh

WORKDIR /zc

COPY go.mod .
COPY go.sum .

COPY cmd ./cmd
COPY pkg ./pkg
COPY vendor ./vendor

RUN GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -mod vendor -a -installsuffix nocgo -o /bin/zc cmd/server/main.go

FROM alpine:latest  
COPY --from=0 /bin/zc /bin

WORKDIR /
ENTRYPOINT ["/bin/zc"]



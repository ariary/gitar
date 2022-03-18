# syntax=docker/dockerfile:1

FROM golang:1.17-alpine

RUN apk add tree bind-tools
RUN apk add openssl && rm -rf /var/cache/apk/*
RUN mkdir /gitar
WORKDIR /gitar
RUN mkdir ./certs && mkdir ./exchange


COPY go.mod ./
COPY *.go ./
COPY pkg ./pkg
COPY ./entrypoint.sh /gitar/
RUN go mod tidy
RUN go mod download

RUN go build gitar.go

RUN chmod 777 /gitar -R
RUN chmod 777 /tmp  -R

ENTRYPOINT [ "/gitar/entrypoint.sh" ]

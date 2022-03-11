# syntax=docker/dockerfile:1

FROM golang:1.17-alpine

RUN apk add tree bind-tools
RUN apk add openssl && rm -rf /var/cache/apk/*
RUN mkdir /gitar
WORKDIR /gitar
RUN mkdir ./certs && mkdir ./exchange

# Write  access for non-root docker user
# RUN addgroup --gid 1024 nonroot
# RUN adduser --disabled-password --gecos "" --ingroup nonroot nonroot 
# RUN chown :nonroot /app
# RUN chmod g+s /app
# RUN chown :nonroot /certs
# RUN chmod 444 /certs

COPY go.mod ./
COPY *.go ./
COPY pkg ./pkg
COPY ./entrypoint.sh /gitar/
RUN go mod tidy
RUN go mod download


# RUN addgroup --system nonroot
# RUN adduser --system nonroot --ingroup nonroot
# RUN chown -R nonroot:nonroot /gitar
# USER nonroot:nonroot


RUN go build gitar.go

RUN chmod 777 /gitar -R


ENTRYPOINT [ "/gitar/entrypoint.sh" ]

# syntax=docker/dockerfile:1

FROM golang:1.16-alpine

RUN mkdir /certs
WORKDIR /app

# Write  access for non-root docker user
# RUN addgroup --gid 1024 nonroot
# RUN adduser --disabled-password --gecos "" --ingroup nonroot nonroot 
# RUN chown :nonroot /app
# RUN chmod g+s /app
# RUN chown :nonroot /certs
# RUN chmod 444 /certs

COPY go.mod ./
RUN go mod download

COPY *.go ./
COPY pkg ./pkg
RUN go build -o /gitar gitar.go

#USER nonroot:nonroot
ENTRYPOINT [ "/gitar" ]

FROM golang:1.9.2
WORKDIR ./src/gopherDigest/
COPY . .
RUN go get ./...
RUN CGO_ENABLED=0 GOOS=linux make build

FROM debian:stretch
RUN apt-get update \
    && apt-get install -y wget \
    && wget https://www.percona.com/downloads/percona-toolkit/3.0.5/binary/debian/stretch/x86_64/percona-toolkit_3.0.5-1.stretch_amd64.deb \
    && apt install -y ./percona-toolkit_3.0.5-1.stretch_amd64.deb 
COPY --from=0 /go/src/gopherDigest/main .
ENTRYPOINT ["./main"]
FROM golang:1.9.2
RUN wget https://www.percona.com/downloads/percona-toolkit/3.0.5/binary/debian/stretch/x86_64/percona-toolkit_3.0.5-1.stretch_amd64.deb \
    sudo apt install percona-toolkit_3.0.5-1.stretch_amd64.deb 
WORKDIR ./src/gopherDigest/
COPY . .
RUN go get ./...
RUN CGO_ENABLED=0 GOOS=linux make build

FROM scratch
COPY --from=0 /go/src/gopherDigest/main .
ENTRYPOINT ["./main"]
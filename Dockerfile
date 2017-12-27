FROM golang:1.9.2 AS build_stage
WORKDIR /go/src/gopherDigest
RUN apt-get update \
    && apt-get install --no-install-recommends -y ca-certificates wget \
    && wget https://www.percona.com/downloads/percona-toolkit/3.0.5/binary/debian/stretch/x86_64/percona-toolkit_3.0.5-1.stretch_amd64.deb \
    && apt install --no-install-recommends -y ./percona-toolkit_3.0.5-1.stretch_amd64.deb \
    && apt-get purge -y wget \
    && rm -rf /var/lib/apt/lists/* 
COPY . .
RUN export GOBIN="/go/bin" \
    && go get ./... \
    && CGO_ENABLED=0 GOOS=linux go install ./src/main.go \
    && rm -rf /go/src/gopherDigest \
    && apt-get purge -y ca-certificates

FROM alpine:3.7
COPY --from=build_stage /go/bin/main .
COPY --from=build_stage /usr/bin/pt-query-digest /usr/bin/vendor_perl/pt-query-digest
ENTRYPOINT ["./main"]
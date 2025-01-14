FROM golang:1.23-bookworm AS build-stage

LABEL org.opencontainers.image.source="https://github.com/thz/dimplex-modbus-exporter"
LABEL org.opencontainers.image.description="thz/dimplex-modbus-exporter expose sensor readings a la prometheus"
LABEL org.opencontainers.image.licenses="Apache-2.0"

ADD . /go/src/github.com/thz/dimplex-modbus-exporter
WORKDIR /go/src/github.com/thz/dimplex-modbus-exporter

ENV GOCACHE=/root/.cache/go-build
RUN --mount=type=cache,target="/root/.cache/go-build" env CGO_ENABLED=0 go build -o dimplex-modbus-exporter ./cmd/exporter

FROM debian:bookworm-slim AS run-stage


# make the container slightly more useful for diagostics
RUN apt-get update && apt-get install -qq -y \
	inetutils-telnet \
	iputils-ping \
	openssl \
	socat

COPY --from=build-stage /go/src/github.com/thz/dimplex-modbus-exporter/dimplex-modbus-exporter /usr/bin/dimplex-modbus-exporter

ENTRYPOINT [ "/usr/bin/dimplex-modbus-exporter" ]

FROM golang:1.23 AS build-stage

WORKDIR /app

COPY ../../src/pinger/go.mod ../../src/pinger/go.sum ./

RUN go mod download

COPY ../../src/pinger ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /pinger

FROM debian:11-slim AS build-release-stage

WORKDIR /

RUN apt-get update && apt-get install -y libcap2-bin && rm -rf /var/lib/apt/lists/*

COPY --from=build-stage /pinger /pinger

COPY ../../config/pinger_config.json /cfg.json

RUN setcap cap_sys_admin+ep /pinger

ENTRYPOINT ["/pinger"]

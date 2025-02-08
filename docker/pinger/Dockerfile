FROM golang:1.23 AS build-stage

WORKDIR /app

COPY ../../src/pinger/go.mod ../../src/pinger/go.sum ./

RUN go mod download

COPY ../../src/pinger ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /pinger

FROM gcr.io/distroless/base-debian11 AS build-release-stage

WORKDIR /

COPY --from=build-stage /pinger /pinger

COPY ../../config/pinger_config.json /cfg.json

ENTRYPOINT ["/pinger"]

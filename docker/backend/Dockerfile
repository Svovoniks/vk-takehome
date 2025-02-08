FROM golang:1.23 AS build-stage

WORKDIR /app

COPY ../../src/backend/go.mod ../../src/backend/go.sum ./

RUN go mod download

COPY ../../src/backend ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /backend

FROM gcr.io/distroless/base-debian11 AS build-release-stage

WORKDIR /

COPY --from=build-stage /backend /backend

ENTRYPOINT ["/backend"]

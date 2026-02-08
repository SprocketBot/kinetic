# syntax=docker/dockerfile:1

FROM golang:1.24.6-alpine AS builder
WORKDIR /src

ARG TARGETOS
ARG TARGETARCH

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} go build -o /out/api ./cmd/api

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app

COPY --from=builder /out/api /app/api
COPY migrations /app/migrations

EXPOSE 8080
ENTRYPOINT ["/app/api"]

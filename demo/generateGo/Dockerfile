# syntax=docker/dockerfile:1.7

##############################
# Builder
##############################
FROM golang:1.25-alpine AS builder

WORKDIR /build

ENV CGO_ENABLED=0 \
    GO111MODULE=on \
    GOPROXY=https://goproxy.cn,direct

RUN apk add --no-cache git ca-certificates tzdata

COPY go.mod go.sum ./

RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .

RUN --mount=type=cache,target=/root/.cache/go-build \
    go build \
    -trimpath \
    -ldflags="-s -w" \
    -o server \
    ./cmd/server

##############################
# Runtime
##############################
FROM alpine:3.22

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S app \
    && adduser -S app -G app

ENV TZ=Asia/Shanghai

COPY --from=builder /build/server .

USER app

EXPOSE 8080

ENTRYPOINT ["./server"]

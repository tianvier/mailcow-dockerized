FROM golang:alpine AS builder

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    GOPROXY=https://goproxy.cn,direct

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -ldflags="-s -w" -o api ./cmd/


FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache tzdata

COPY --from=builder /build/api .

ENTRYPOINT ["./api"]

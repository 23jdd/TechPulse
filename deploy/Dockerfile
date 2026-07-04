# syntax=docker/dockerfile:1.7

FROM golang:1.25-alpine AS build

ARG GOPROXY=https://goproxy.cn,direct
ARG ALPINE_MIRROR=https://mirrors.aliyun.com/alpine
ENV GOPROXY=${GOPROXY} \
    GOSUMDB=sum.golang.google.cn

RUN if [ -n "$ALPINE_MIRROR" ]; then \
      sed -i "s#https://dl-cdn.alpinelinux.org/alpine#$ALPINE_MIRROR#g" /etc/apk/repositories; \
    fi

WORKDIR /src
COPY go.mod go.sum* ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    mkdir -p /out \
 && go build -o /out/ ./cmd/gateway ./cmd/scheduler ./cmd/fetcher ./cmd/parser ./cmd/ai-pipeline ./cmd/search ./cmd/rag ./cmd/worker

FROM alpine:3.20

ARG ALPINE_MIRROR=https://mirrors.aliyun.com/alpine
RUN if [ -n "$ALPINE_MIRROR" ]; then \
      sed -i "s#https://dl-cdn.alpinelinux.org/alpine#$ALPINE_MIRROR#g" /etc/apk/repositories; \
    fi

RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=build /out/ /usr/local/bin/
COPY web ./web
CMD ["gateway"]

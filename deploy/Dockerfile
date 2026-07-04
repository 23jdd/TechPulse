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
RUN go mod download
COPY . .
RUN go build -o /out/gateway ./cmd/gateway \
 && go build -o /out/scheduler ./cmd/scheduler \
 && go build -o /out/fetcher ./cmd/fetcher \
 && go build -o /out/parser ./cmd/parser \
 && go build -o /out/ai-pipeline ./cmd/ai-pipeline \
 && go build -o /out/search ./cmd/search \
 && go build -o /out/rag ./cmd/rag \
 && go build -o /out/worker ./cmd/worker

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

FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY . .
RUN apk add --no-cache make git ca-certificates && \
    make deps && make build-all

FROM node:20-alpine
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app
COPY --from=builder /app/build/meme-server /usr/local/bin/meme-server
RUN chmod +x /usr/local/bin/meme-server

EXPOSE 8080

# Railway 会自动注入 $PORT 和你的环境变量
CMD sh -c 'npx -y supergateway \
  --stdio "/usr/local/bin/meme-server" \
  --port ${PORT:-8080} \
  --listen 0.0.0.0 \
  --ssePath /sse'

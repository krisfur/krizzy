FROM golang:1.25-bookworm AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go run github.com/a-h/templ/cmd/templ@latest generate
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o /out/krizzy ./cmd/server

FROM debian:bookworm-slim AS runtime

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates curl && rm -rf /var/lib/apt/lists/*

WORKDIR /app

RUN useradd --system --create-home --home-dir /app --shell /usr/sbin/nologin krizzy && mkdir -p /data && chown -R krizzy:krizzy /app /data

COPY --from=builder /out/krizzy /app/krizzy
COPY --from=builder /app/static /app/static

USER krizzy

ENV SERVER_ADDRESS=:8080
ENV DATABASE_PATH=/data/krizzy.db

EXPOSE 8080
VOLUME ["/data"]

HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 CMD curl --fail --silent http://127.0.0.1:8080/healthz > /dev/null || exit 1

CMD ["/app/krizzy"]

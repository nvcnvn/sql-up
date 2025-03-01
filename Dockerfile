FROM golang:1.24-bookworm AS builder

WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build -o /app/sql-up ./cmd/

FROM debian:bookworm-slim
COPY --from=builder /app/sql-up /usr/local/bin/sql-up

ENTRYPOINT ["sql-up"]

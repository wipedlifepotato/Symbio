FROM docker.io/golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o mFrelance main.go

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/mFrelance .
COPY config.yaml .
COPY mods/ ./mods/
RUN mkdir -p ./db/migrations
COPY db/migrations/ ./db/migrations

EXPOSE 9999
#ENV ELECTRUM_PORT 7777
#ENV POSTGRES_PORT 5432
#ENV REDIS_PORT 6379

#ENTRYPOINT ["./mFrelance"]
CMD sh -c "./mFrelance \
  --electrum.port=${ELECTRUM_PORT} \
  --listen_addr=0.0.0.0 \
  --postgres.host=postgres \
  --postgres.port=${POSTGRES_PORT} \
  --postgres.user=${POSTGRES_USER} \
  --postgres.password=${POSTGRES_PASSWORD} \
  --postgres.db=${POSTGRES_DB} \
  --redis.host=redis \
  --redis.port=${REDIS_PORT} \
  --electrum.host=electrum-server"

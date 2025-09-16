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

EXPOSE 9999

ENTRYPOINT ["./mFrelance"]
CMD ["--electrum.port=7777","--listen_addr=0.0.0.0","--postgres.host=postgres","--postgres.port=5432","--postgres.user=postgres","--postgres.password=mysecretpassword","--postgres.db=mydb","--redis.host=redis","--redis.port=6379","--electrum.host=electrum-server"]


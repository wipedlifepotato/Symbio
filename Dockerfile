FROM docker.io/golang:1.24 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod tidy
RUN go mod download
COPY . .
RUN go build -o mFrelance main.go

FROM debian:bookworm-slim

#RUN curl https://get.wasmer.io -sSfL | sh
RUN apt-get update && apt-get install -y --no-install-recommends wget ca-certificates \ 
 && wget https://github.com/wasmerio/wasmer/releases/download/v6.0.1/wasmer-linux-amd64.tar.gz \
 && tar -xzf wasmer-linux-amd64.tar.gz -C /usr/local \
 && rm wasmer-linux-amd64.tar.gz

ENV PATH="/usr/local/bin:$PATH"
ENV LD_LIBRARY_PATH="/usr/local/lib:$LD_LIBRARY_PATH"
RUN ldconfig
WORKDIR /app
COPY --from=builder /app/mFrelance .
COPY config.yaml .
COPY mods/ ./mods/
RUN mkdir -p ./modsWasm
COPY modsWasm/*.wasm ./modsWasm
RUN mkdir -p ./db/migrations
COPY db/migrations/ ./db/migrations

EXPOSE 9999

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

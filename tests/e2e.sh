#!/bin/bash
export BASE_URL=http://127.0.0.1:9999
export REDIS_ADDR=127.0.0.1:6379
go test -tags=e2e ../tests -v

#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
export PATH="${HOME}/sdk/go/bin:${HOME}/go/bin:${HOME}/sdk/protoc/bin:${PATH}"

cd "$ROOT"
mkdir -p gen/demo/v1
protoc \
  -I api \
  --go_out=gen --go_opt=paths=source_relative \
  --go-grpc_out=gen --go-grpc_opt=paths=source_relative \
  api/demo/v1/demo.proto

find gen -type f | sort

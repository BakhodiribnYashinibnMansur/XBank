#!/usr/bin/env bash
set -euo pipefail

# Protobuf code generation for XBank Kafka messages.
#
# Prerequisites:
#   brew install protobuf
#   go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
#
# Usage:
#   ./scripts/protogen.sh

PROTO_DIR="proto"
OUT_DIR="internal/kernel/infrastructure/kafka/pb"

if ! command -v protoc &> /dev/null; then
    echo "ERROR: protoc not found. Install Protocol Buffers compiler."
    echo "  brew install protobuf   # macOS"
    echo "  apt install protobuf-compiler  # Ubuntu"
    exit 1
fi

if ! command -v protoc-gen-go &> /dev/null; then
    echo "ERROR: protoc-gen-go not found. Install it:"
    echo "  go install google.golang.org/protobuf/cmd/protoc-gen-go@latest"
    exit 1
fi

mkdir -p "$OUT_DIR"

echo "Generating Go code from .proto files..."

protoc \
    --proto_path="$PROTO_DIR" \
    --go_out="$OUT_DIR" \
    --go_opt=paths=source_relative \
    "$PROTO_DIR"/**/*.proto

echo "Done. Generated files in $OUT_DIR/"
ls -la "$OUT_DIR"/*.go 2>/dev/null || echo "  (no .go files generated — check proto definitions)"

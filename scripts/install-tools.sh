#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
source tools/versions.env
mkdir -p .tools/bin
export GOBIN="$PWD/.tools/bin"
export PATH="$GOBIN:$PATH"
go install "github.com/zeromicro/go-zero/tools/goctl@v$GOCTL_VERSION"
go install "google.golang.org/protobuf/cmd/protoc-gen-go@v$PROTOC_GEN_GO_VERSION"
go install "google.golang.org/grpc/cmd/protoc-gen-go-grpc@v$PROTOC_GEN_GO_GRPC_VERSION"
if command -v protoc >/dev/null && [[ "$(protoc --version)" == "libprotoc $PROTOC_VERSION" ]]; then
 echo "protoc $PROTOC_VERSION already available"
else
 case "$(uname -s)-$(uname -m)" in
  Darwin-arm64|Darwin-x86_64) platform=osx-universal_binary;;
  Linux-x86_64) platform=linux-x86_64;;
  Linux-aarch64|Linux-arm64) platform=linux-aarch_64;;
  *) echo "Install protoc $PROTOC_VERSION for your platform manually" >&2; exit 1;;
 esac
 archive=$(mktemp)
 trap 'rm -f "$archive"' EXIT
 curl --fail --location --retry 3 "https://github.com/protocolbuffers/protobuf/releases/download/v$PROTOC_VERSION/protoc-$PROTOC_VERSION-$platform.zip" -o "$archive"
 unzip -oq "$archive" -d .tools
fi

#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
source tools/versions.env
export PATH="$PWD/.tools/bin:$PATH"
check() {
 [[ "$2" == "$3" ]] || { echo "$1: expected $3, got $2" >&2; exit 1; }
 echo "$1: $2"
}
check Go "$(go env GOVERSION)" "go$GO_VERSION"
check goctl "$(goctl --version | awk '{print $3}')" "$GOCTL_VERSION"
check protoc "$(protoc --version | awk '{print $2}')" "$PROTOC_VERSION"
check protoc-gen-go "$(protoc-gen-go --version | awk '{print $2}')" "v$PROTOC_GEN_GO_VERSION"
check protoc-gen-go-grpc "$(protoc-gen-go-grpc --version | awk '{print $2}')" "$PROTOC_GEN_GO_GRPC_VERSION"
check Node "$(node --version)" "v$NODE_VERSION"
check pnpm "$(pnpm --version)" "$PNPM_VERSION"
python3 --version
docker compose version
docker info --format '{{.ServerVersion}}'
[[ -f .ai-context/go-zero/00-instructions.md && -f .agents/skills/zero-skills/SKILL.md ]] || { echo 'Run git submodule update --init --recursive'; exit 1; }

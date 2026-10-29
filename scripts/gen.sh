#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
export PATH="$PWD/.tools/bin:$PATH"
source tools/versions.env
[[ "$(goctl --version)" == "goctl version ${GOCTL_VERSION} "* ]] || { echo "Run make tools: goctl version mismatch" >&2; exit 1; }
mode=${1:-all}
if [[ "$mode" == api || "$mode" == all ]]; then
 for app in ${APP:-admin app}; do
  [[ "$app" == admin || "$app" == app ]] || exit 1
  goctl api validate --api "apps/$app/api/desc/$app.api"
  goctl api go --api "apps/$app/api/desc/$app.api" --dir "apps/$app/api" --style go_zero --home templates/goctl
 done
fi
if [[ "$mode" == rpc || "$mode" == all ]]; then
 [[ "${DOMAIN:-system}" == system ]] || { echo "Unknown domain" >&2; exit 1; }
 [[ "$(protoc --version)" == "libprotoc $PROTOC_VERSION" ]] || exit 1
 [[ "$(protoc-gen-go --version)" == "protoc-gen-go v$PROTOC_GEN_GO_VERSION" ]] || exit 1
 [[ "$(protoc-gen-go-grpc --version)" == "protoc-gen-go-grpc $PROTOC_GEN_GO_GRPC_VERSION" ]] || exit 1
 goctl rpc protoc services/system/rpc/system.proto --proto_path=services/system/rpc   --go_out=. --go_opt=module=github.com/kar1hsu/go-seed   --go-grpc_out=. --go-grpc_opt=module=github.com/kar1hsu/go-seed   --zrpc_out=services/system/rpc --style go_zero --name-from-filename --home templates/goctl
fi
if [[ "$mode" == web || "$mode" == all ]]; then
 for app in admin app; do
  goctl api swagger --api "apps/$app/api/desc/$app.api" --dir docs/openapi --filename "$app"
 done
 node web/admin/scripts/resources.mjs --write
 node web/admin/scripts/generate-api.mjs
fi
case "$mode" in all|api|rpc|web) ;; *) echo "Usage: $0 [all|api|rpc|web]" >&2; exit 1;; esac

FROM golang:1.25.11-bookworm AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY apps ./apps
COPY services ./services
COPY pkg ./pkg
COPY cmd ./cmd
ARG SERVICE_PATH
RUN CGO_ENABLED=0 go build -trimpath -o /out/server ./${SERVICE_PATH} && CGO_ENABLED=0 go build -trimpath -o /out/probe ./cmd/probe && CGO_ENABLED=0 go build -trimpath -o /out/systemctl ./services/system/cmd/system && CGO_ENABLED=0 go build -trimpath -o /out/memberctl ./services/member/cmd/member

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=build /out/ /app/
COPY apps/admin/api/etc /app/apps/admin/api/etc
COPY apps/app/api/etc /app/apps/app/api/etc
COPY services/system/rpc/etc /app/services/system/rpc/etc
COPY services/member/rpc/etc /app/services/member/rpc/etc
COPY services/system/worker/etc /app/services/system/worker/etc
COPY services/system/scheduler/etc /app/services/system/scheduler/etc
RUN mkdir -p /app/data/files && chown -R 65532:65532 /app/data
USER 65532:65532
ENTRYPOINT ["/app/server"]

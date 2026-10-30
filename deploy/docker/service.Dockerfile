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

RUN mkdir -p /out/data/files

FROM gcr.io/distroless/static-debian13:nonroot@sha256:18f4cc26c9ffa7f20a359298be7deadac5c607c8e03a02b8119f1a5fc2f82c99
ENV TZ=Asia/Shanghai
WORKDIR /app
COPY --from=build /out/server /out/probe /out/systemctl /out/memberctl /app/
COPY --from=build --chown=65532:65532 /out/data/ /app/data/
COPY apps/admin/api/etc /app/apps/admin/api/etc
COPY apps/app/api/etc /app/apps/app/api/etc
COPY services/system/rpc/etc /app/services/system/rpc/etc
COPY services/member/rpc/etc /app/services/member/rpc/etc
COPY services/system/worker/etc /app/services/system/worker/etc
COPY services/system/scheduler/etc /app/services/system/scheduler/etc
USER 65532:65532
ENTRYPOINT ["/app/server"]

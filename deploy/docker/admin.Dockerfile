FROM node:24.19.0-bookworm-slim AS build
WORKDIR /src
RUN npm install --global pnpm@11.15.1
COPY web/admin/package.json web/admin/pnpm-lock.yaml web/admin/pnpm-workspace.yaml ./
RUN pnpm install --frozen-lockfile
COPY web/admin/ ./
RUN pnpm build

FROM nginx:1.28.0-alpine
ENV TZ=Asia/Shanghai
RUN apk add --no-cache tzdata
COPY deploy/nginx/default.conf /etc/nginx/conf.d/default.conf
COPY --from=build /src/dist /usr/share/nginx/html

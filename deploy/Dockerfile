# =============================================================================
# TokenRouter Multi-Stage Dockerfile
# =============================================================================
# Stage 1: Build frontend
# Stage 2: Build Go backend with embedded frontend
# Stage 3: Final minimal image
# =============================================================================

ARG NODE_IMAGE=node:24-alpine
ARG GOLANG_IMAGE=golang:1.27.0-alpine
ARG ALPINE_IMAGE=alpine:3.20
ARG GOPROXY=https://goproxy.cn,direct
ARG GOSUMDB=sum.golang.google.cn

# -----------------------------------------------------------------------------
# Stage 1: Frontend Builder
# -----------------------------------------------------------------------------
FROM ${NODE_IMAGE} AS frontend-builder

WORKDIR /app/frontend
ENV NODE_OPTIONS=--max-old-space-size=1536

# 安装 pnpm，并固定到与 lockfile 兼容的精确版本，避免配置校验规则变化破坏构建。
RUN corepack enable && corepack prepare pnpm@9.15.9 --activate

# Install dependencies first (better caching)
COPY frontend/package.json frontend/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

# 复制前端源码并构建。
# LegalDocumentView.vue 构建时会通过 ../../../../docs/legal/*.md?raw
# 读取法律文档，因此 docs/legal/ 需要与 frontend/ 同级放在 /app/docs/legal/。
COPY frontend/ ./
COPY docs/legal/ /app/docs/legal/
RUN pnpm run build

# -----------------------------------------------------------------------------
# Stage 2: Backend Builder
# -----------------------------------------------------------------------------
FROM ${GOLANG_IMAGE} AS backend-builder

# 构建版本信息参数，由 CI 或手工构建传入
ARG VERSION=
ARG COMMIT=docker
ARG DATE
ARG GOPROXY
ARG GOSUMDB

ENV GOPROXY=${GOPROXY}
ENV GOSUMDB=${GOSUMDB}

# 安装构建依赖
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app/backend

# 先复制 go mod 文件以提高缓存命中率
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# 复制后端源码
COPY backend/ ./

# 再复制前端构建产物，避免被后端源码覆盖
COPY --from=frontend-builder /app/backend/internal/web/dist ./internal/web/dist

# 构建 release 二进制，并嵌入前端资源
# 版本优先级：构建参数 VERSION > 精确 Git tag > cmd/server/VERSION
RUN VERSION_VALUE="${VERSION}" && \
    if [ -z "${VERSION_VALUE}" ]; then VERSION_VALUE="$(./scripts/resolve-version.sh)"; fi && \
    DATE_VALUE="${DATE:-$(date -u +%Y-%m-%dT%H:%M:%SZ)}" && \
    CGO_ENABLED=0 GOOS=linux go build \
    -tags embed \
    -ldflags="-s -w -X main.Version=${VERSION_VALUE} -X main.Commit=${COMMIT} -X main.Date=${DATE_VALUE} -X main.BuildType=release" \
    -o /app/tokenrouter \
    ./cmd/server

# -----------------------------------------------------------------------------
# Stage 3: Final Runtime Image
# -----------------------------------------------------------------------------
FROM ${ALPINE_IMAGE}

# Labels
LABEL maintainer="TokenFlux <github.com/TokenFlux>"
LABEL description="TokenRouter - AI API Gateway Platform"
LABEL org.opencontainers.image.source="https://github.com/TokenFlux/TokenRouter"

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    curl \
    su-exec \
    && rm -rf /var/cache/apk/*

# Create non-root user
RUN addgroup -g 1000 tokenrouter && \
    adduser -u 1000 -G tokenrouter -s /bin/sh -D tokenrouter

# Set working directory
WORKDIR /app

# 从构建阶段复制已内嵌默认价格补充的二进制。
COPY --from=backend-builder /app/tokenrouter /app/tokenrouter

# Create data directory
RUN mkdir -p /app/data && chown -R tokenrouter:tokenrouter /app

# Copy entrypoint script (fixes volume permissions then drops to tokenrouter)
COPY deploy/docker-entrypoint.sh /app/docker-entrypoint.sh
RUN chmod +x /app/docker-entrypoint.sh

# Expose port (can be overridden by SERVER_PORT env var)
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3 \
    CMD wget -q -T 5 -O /dev/null http://localhost:${SERVER_PORT:-8080}/health || exit 1

# Run the application (entrypoint fixes /app/data ownership then execs as tokenrouter)
ENTRYPOINT ["/app/docker-entrypoint.sh"]
# 旧 Compose 可继续调用原可执行路径。
RUN ln -s /app/tokenrouter /app/sub2api
CMD ["/app/tokenrouter"]

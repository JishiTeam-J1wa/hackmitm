# 多阶段构建的 Dockerfile
# Stage 1: 构建阶段
FROM golang:1.21-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装必要的工具
RUN apk add --no-cache git ca-certificates tzdata

# 复制 go.mod 和 go.sum
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建可执行文件
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-X main.Version=1.0.0 -X main.BuildTime=$(date +%Y-%m-%d_%H:%M:%S) -s -w" \
    -o hackmitm ./cmd/hackmitm

# Stage 2: 运行阶段
FROM alpine:latest

# 安装 CA 证书和时区数据
RUN apk --no-cache add ca-certificates tzdata

# 创建非 root 用户
RUN addgroup -g 1001 hackmitm && \
    adduser -D -s /bin/sh -u 1001 -G hackmitm hackmitm

# 设置工作目录
WORKDIR /app

# 从构建阶段复制可执行文件
COPY --from=builder /app/hackmitm /usr/local/bin/hackmitm

# 复制配置文件和文档
COPY --chown=hackmitm:hackmitm configs ./configs
COPY --chown=hackmitm:hackmitm docs ./docs

# 创建证书目录
RUN mkdir -p /app/certs && chown hackmitm:hackmitm /app/certs

# 切换到非 root 用户
USER hackmitm

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider --proxy=on --proxy-user="" --proxy-password="" \
    --proxy-on http://localhost:8080 http://httpbin.org/status/200 || exit 1

# 设置环境变量
ENV HACKMITM_CONFIG=/app/configs/config.json
ENV HACKMITM_CERT_DIR=/app/certs

# 默认命令
CMD ["hackmitm", "-config", "/app/configs/config.json"] 
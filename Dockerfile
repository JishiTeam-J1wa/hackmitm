# 多阶段构建的 Dockerfile - 优化版本
# Stage 1: 构建阶段
FROM golang:1.21-alpine AS builder

# 设置构建参数
ARG VERSION=dev
ARG BUILD_TIME
ARG GIT_COMMIT

# 安装构建依赖
RUN apk add --no-cache \
    git \
    ca-certificates \
    tzdata \
    upx

# 设置工作目录
WORKDIR /app

# 优化Go构建环境
ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    GO111MODULE=on \
    GOPROXY=https://goproxy.cn,direct

# 复制依赖文件并下载（利用Docker缓存）
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# 复制源代码
COPY . .

# 构建插件
RUN cd plugins && make examples

# 构建主程序（优化编译参数）
RUN go build \
    -ldflags "-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT} -s -w -extldflags '-static'" \
    -tags netgo \
    -installsuffix netgo \
    -trimpath \
    -o hackmitm ./cmd/hackmitm

# 使用UPX压缩二进制文件（可选）
RUN upx --best --lzma hackmitm || true

# Stage 2: 运行阶段 - 使用distroless镜像提高安全性
FROM gcr.io/distroless/static-debian11:nonroot

# 添加标签信息
LABEL maintainer="JishiTeam-J1wa" \
      description="HackMITM - 高性能HTTP/HTTPS代理服务器" \
      version="${VERSION}" \
      org.opencontainers.image.title="HackMITM" \
      org.opencontainers.image.description="高性能、可扩展的HTTP/HTTPS代理服务器" \
      org.opencontainers.image.authors="JishiTeam-J1wa" \
      org.opencontainers.image.vendor="JishiTeam-J1wa" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.url="https://github.com/JishiTeam-J1wa/hackmitm" \
      org.opencontainers.image.source="https://github.com/JishiTeam-J1wa/hackmitm" \
      org.opencontainers.image.documentation="https://github.com/JishiTeam-J1wa/hackmitm/blob/main/README.md"

# 创建必要的目录结构
USER nonroot:nonroot
WORKDIR /app

# 从构建阶段复制文件
COPY --from=builder --chown=nonroot:nonroot /app/hackmitm /app/
COPY --from=builder --chown=nonroot:nonroot /app/configs /app/configs/
COPY --from=builder --chown=nonroot:nonroot /app/plugins/examples/*.so /app/plugins/examples/
COPY --from=builder --chown=nonroot:nonroot /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# 创建数据目录
USER root
RUN mkdir -p /app/logs /app/certs && \
    chown -R nonroot:nonroot /app/logs /app/certs
USER nonroot:nonroot

# 暴露端口
EXPOSE 8081 9090

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD ["/app/hackmitm", "--health-check"]

# 设置默认命令
ENTRYPOINT ["/app/hackmitm"]
CMD ["-config", "/app/configs/config.json"] 
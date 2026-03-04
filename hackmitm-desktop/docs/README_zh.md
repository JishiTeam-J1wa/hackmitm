# HackMITM - 高性能被动代理处理包

[![Go版本](https://img.shields.io/badge/go-%3E%3D1.21-blue.svg)](https://golang.org/)
[![许可证](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

HackMITM 是一个使用纯原生 Golang 开发的高性能被动代理处理包，专门用于处理 HTTP 和 HTTPS 流量，支持完整的 MITM (Man-in-the-Middle) 代理功能。它具有高效、安全、灵活且易于扩展的特点，旨在成为 Golang 安全研发领域的跨时代工具。

## 主要特性

### 🚀 高性能设计
- **纯原生 Golang 实现**：无依赖外部库，最大化性能表现
- **高效并发模型**：基于 goroutine 的并发处理，支持高并发场景
- **连接池优化**：智能连接复用，减少资源消耗
- **缓存机制**：证书缓存、连接缓存等多级缓存优化

### 🔐 完整的 MITM 功能
- **动态证书生成**：根据目标域名自动生成 TLS 证书
- **CA 证书管理**：完整的 CA 证书生成、存储和管理
- **透明代理**：对客户端完全透明的代理服务
- **TLS 流量解密**：支持 HTTPS 流量的完整解密和重加密

### 🧩 模块化架构
- **证书管理模块**：负责 CA 和服务器证书的生成与管理
- **代理服务器模块**：核心代理服务器实现
- **流量处理模块**：HTTP/HTTPS 流量解析、处理和重新封装
- **配置管理模块**：灵活的配置系统，支持热加载

### ⚙️ 丰富的配置选项
- **灵活配置**：支持 JSON 配置文件和命令行参数
- **热加载配置**：无需重启即可应用新配置
- **性能调优**：可配置的性能参数和资源限制
- **日志管理**：分级日志记录，支持多种输出格式

## 架构设计

### 整体架构图

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   客户端应用     │    │   HackMITM      │    │   目标服务器     │
│   Client App    │◄──►│   Proxy Server  │◄──►│  Target Server  │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                              │
                              ▼
                    ┌─────────────────────┐
                    │    模块化组件        │
                    │  ┌───────────────┐  │
                    │  │  证书管理模块  │  │
                    │  │ Certificate   │  │
                    │  │   Manager     │  │
                    │  └───────────────┘  │
                    │  ┌───────────────┐  │
                    │  │  流量处理模块  │  │
                    │  │   Traffic     │  │
                    │  │  Processor    │  │
                    │  └───────────────┘  │
                    │  ┌───────────────┐  │
                    │  │  配置管理模块  │  │
                    │  │ Configuration │  │
                    │  │   Manager     │  │
                    │  └───────────────┘  │
                    └─────────────────────┘
```

### 模块详细说明

#### 1. 证书管理模块 (Certificate Manager)
- **CA 证书生成**：使用 ECDSA P-256 算法生成高安全性的 CA 证书
- **服务器证书生成**：根据目标域名动态生成有效的服务器证书
- **证书缓存**：智能缓存机制，避免重复生成相同域名的证书
- **证书导出**：支持将 CA 证书导出供客户端安装

#### 2. 代理服务器模块 (Proxy Server)
- **HTTP 代理**：完整的 HTTP 1.1 代理支持
- **HTTPS 代理**：通过 CONNECT 方法实现 HTTPS 流量劫持
- **连接管理**：高效的连接池和连接复用机制
- **并发处理**：每个连接独立的 goroutine 处理

#### 3. 流量处理模块 (Traffic Processor)
- **请求处理链**：可扩展的请求处理器链模式
- **响应处理链**：可配置的响应处理器链
- **内容压缩**：支持 Gzip 和 Brotli 压缩算法
- **流量分析**：详细的流量日志和统计信息

#### 4. 配置管理模块 (Configuration Manager)
- **JSON 配置**：结构化的 JSON 配置文件
- **热加载**：配置文件变更自动检测和应用
- **参数验证**：配置参数的完整性和有效性验证
- **默认配置**：合理的默认配置，开箱即用

## 快速开始

### 环境要求
- Go 1.21 或更高版本
- 支持的操作系统：Linux、macOS、Windows

### 安装与构建

```bash
# 克隆项目
git clone https://github.com/your-org/hackmitm.git
cd hackmitm

# 安装依赖
go mod tidy

# 构建项目
go build -o hackmitm ./cmd/hackmitm

# 或者使用 Makefile（如果提供）
make build
```

### 基本使用

#### 1. 启动代理服务器

```bash
# 使用默认配置启动
./hackmitm

# 指定配置文件启动
./hackmitm -config ./configs/my-config.json

# 指定监听端口
./hackmitm -port 8888

# 启用详细日志
./hackmitm -verbose
```

#### 2. 配置客户端代理

```bash
# 设置 HTTP 代理
export http_proxy=http://localhost:8080
export https_proxy=http://localhost:8080

# 或者使用 curl 测试
curl --proxy http://localhost:8080 http://example.com
curl --proxy http://localhost:8080 https://example.com
```

#### 3. 安装 CA 证书

```bash
# 导出 CA 证书
./hackmitm -export-ca ./ca-cert.pem

# 在客户端安装 CA 证书以信任 HTTPS 连接
# macOS:
sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain ca-cert.pem

# Linux:
sudo cp ca-cert.pem /usr/local/share/ca-certificates/hackmitm.crt
sudo update-ca-certificates

# Windows:
# 双击 ca-cert.pem 文件，选择"安装证书"到"受信任的根证书颁发机构"
```

## 配置说明

### 配置文件结构

```json
{
  "server": {
    "listen_port": 8080,
    "listen_addr": "0.0.0.0",
    "read_timeout": "30s",
    "write_timeout": "30s"
  },
  "tls": {
    "cert_dir": "./certs",
    "ca_key_file": "./certs/ca-key.pem",
    "ca_cert_file": "./certs/ca-cert.pem",
    "enable_cert_cache": true,
    "cert_cache_ttl": "24h"
  },
  "proxy": {
    "enable_http": true,
    "enable_https": true,
    "upstream_timeout": "30s",
    "max_idle_conns": 100,
    "enable_compression": true
  },
  "logging": {
    "level": "info",
    "output": "stdout",
    "format": "text",
    "enable_file_rotation": false
  },
  "performance": {
    "max_goroutines": 10000,
    "buffer_size": 4096,
    "enable_pprof": false,
    "pprof_port": 6060
  }
}
```

### 配置项详解

#### 服务器配置 (server)
- `listen_port`: 代理服务器监听端口（默认：8080）
- `listen_addr`: 监听地址（默认："0.0.0.0"）
- `read_timeout`: 读取超时时间
- `write_timeout`: 写入超时时间

#### TLS 配置 (tls)
- `cert_dir`: 证书存储目录
- `ca_key_file`: CA 私钥文件路径
- `ca_cert_file`: CA 证书文件路径
- `enable_cert_cache`: 是否启用证书缓存
- `cert_cache_ttl`: 证书缓存有效期

#### 代理配置 (proxy)
- `enable_http`: 启用 HTTP 代理
- `enable_https`: 启用 HTTPS 代理
- `upstream_timeout`: 上游服务器超时时间
- `max_idle_conns`: 最大空闲连接数
- `enable_compression`: 启用响应压缩

#### 日志配置 (logging)
- `level`: 日志级别（debug、info、warn、error）
- `output`: 输出目标（stdout、stderr 或文件路径）
- `format`: 日志格式（text、json）
- `enable_file_rotation`: 启用文件轮转

#### 性能配置 (performance)
- `max_goroutines`: 最大 goroutine 数量
- `buffer_size`: 缓冲区大小
- `enable_pprof`: 启用性能分析
- `pprof_port`: 性能分析端口

## 高级使用

### 自定义处理器

HackMITM 支持自定义请求和响应处理器，可以实现各种高级功能：

```go
// 自定义请求处理器
type CustomRequestHandler struct{}

func (h *CustomRequestHandler) HandleRequest(req *http.Request) error {
    // 在这里添加自定义请求处理逻辑
    // 例如：修改请求头、记录请求信息、实现访问控制等
    return nil
}

// 自定义响应处理器
type CustomResponseHandler struct{}

func (h *CustomResponseHandler) HandleResponse(resp *http.Response, req *http.Request) error {
    // 在这里添加自定义响应处理逻辑
    // 例如：修改响应内容、添加响应头、实现内容过滤等
    return nil
}
```

### 上游代理链

支持设置上游代理，实现代理链功能：

```bash
# 设置上游代理
./hackmitm -upstream http://upstream-proxy:8080
```

### 性能监控

启用 pprof 性能分析：

```bash
# 启动时启用 pprof
./hackmitm -config config.json

# 访问性能分析页面
go tool pprof http://localhost:6060/debug/pprof/profile
```

## 安全考虑

### 证书安全
- CA 私钥使用 0600 权限存储，仅文件所有者可读写
- 证书使用 ECDSA P-256 算法，提供高级别安全性
- 支持证书定期轮换和缓存清理

### 网络安全
- 支持 TLS 1.2 和 TLS 1.3 协议
- 完整的证书链验证
- 防止证书伪造和中间人攻击

### 访问控制
- 可配置的访问日志记录
- 支持基于 IP 的访问控制（通过自定义处理器实现）
- 完整的审计日志

## 故障排除

### 常见问题

#### 1. 证书相关问题
```
错误：x509: certificate signed by unknown authority
解决：确保已正确安装 CA 证书到客户端的受信任根证书存储
```

#### 2. 连接问题
```
错误：dial tcp: connection refused
解决：检查代理服务器是否正常启动，监听端口是否正确
```

#### 3. 性能问题
```
问题：代理速度较慢
解决：调整配置中的 max_idle_conns 和 buffer_size 参数
```

### 调试模式

启用详细调试日志：

```bash
./hackmitm -verbose -log-level debug
```

### 日志分析

HackMITM 提供详细的日志信息，包括：
- 请求/响应详情
- TLS 握手信息
- 证书生成和缓存状态
- 性能统计信息

## 贡献指南

我们欢迎社区贡献！请遵循以下步骤：

1. Fork 项目仓库
2. 创建特性分支：`git checkout -b feature/amazing-feature`
3. 提交更改：`git commit -m 'Add amazing feature'`
4. 推送分支：`git push origin feature/amazing-feature`
5. 提交 Pull Request

### 开发规范

- 遵循 Go 语言编码规范
- 添加适当的单元测试
- 更新相关文档
- 确保所有测试通过

## 许可证

本项目采用 MIT 许可证 - 详情请参阅 [LICENSE](LICENSE) 文件。

## 支持与联系

- 📚 文档：[完整文档](./docs/user_manual_zh.md)
- 🐛 问题反馈：[GitHub Issues](https://github.com/your-org/hackmitm/issues)
- 💬 讨论：[GitHub Discussions](https://github.com/your-org/hackmitm/discussions)
- 📧 邮件：hackmitm@example.com

---

**HackMITM** - 让 MITM 代理变得简单而强大 🚀 
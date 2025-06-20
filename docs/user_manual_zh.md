# HackMITM 用户手册

## 目录

1. [简介](#简介)
2. [安装部署](#安装部署)
3. [基础配置](#基础配置)
4. [使用指南](#使用指南)
5. [高级功能](#高级功能)
6. [API 参考](#api-参考)
7. [故障排除](#故障排除)
8. [最佳实践](#最佳实践)

## 简介

HackMITM 是一个高性能的 MITM（中间人）代理工具，使用 Go 语言开发。它能够透明地拦截、分析和修改 HTTP/HTTPS 流量，是安全研究、渗透测试和流量分析的理想工具。

### 核心功能

- **透明代理**: 对客户端应用完全透明的代理服务
- **HTTPS 支持**: 自动生成证书，完整支持 HTTPS 流量解密
- **流量处理**: 可扩展的请求/响应处理器链
- **高性能**: 基于 goroutine 的高并发处理
- **易于扩展**: 模块化设计，支持自定义插件

## 安装部署

### 系统要求

- **操作系统**: Linux、macOS、Windows
- **Go 版本**: 1.21 或更高版本
- **内存**: 最低 512MB，推荐 2GB 或以上
- **磁盘空间**: 最低 100MB

### 从源码安装

```bash
# 1. 克隆项目
git clone https://github.com/your-org/hackmitm.git
cd hackmitm

# 2. 安装依赖
go mod tidy

# 3. 构建可执行文件
go build -ldflags "-X main.Version=1.0.0 -X main.BuildTime=$(date +%Y-%m-%d_%H:%M:%S)" -o hackmitm ./cmd/hackmitm

# 4. 验证安装
./hackmitm -version
```

### 预编译二进制文件

```bash
# 下载预编译文件
wget https://github.com/your-org/hackmitm/releases/latest/download/hackmitm-linux-amd64.tar.gz

# 解压并安装
tar -xzf hackmitm-linux-amd64.tar.gz
sudo mv hackmitm /usr/local/bin/
sudo chmod +x /usr/local/bin/hackmitm
```

### Docker 部署

```bash
# 构建 Docker 镜像
docker build -t hackmitm:latest .

# 运行容器
docker run -d \
  --name hackmitm \
  -p 8080:8080 \
  -v $(pwd)/certs:/app/certs \
  -v $(pwd)/configs:/app/configs \
  hackmitm:latest
```

## 基础配置

### 配置文件

HackMITM 使用 JSON 格式的配置文件。默认配置文件位置：`./configs/config.json`

#### 最小配置示例

```json
{
  "server": {
    "listen_port": 8080,
    "listen_addr": "0.0.0.0"
  },
  "tls": {
    "cert_dir": "./certs"
  },
  "proxy": {
    "enable_http": true,
    "enable_https": true
  }
}
```

#### 完整配置示例

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

### 环境变量

支持通过环境变量覆盖配置：

```bash
export HACKMITM_LISTEN_PORT=9090
export HACKMITM_LOG_LEVEL=debug
export HACKMITM_CERT_DIR=/custom/cert/path
```

## 使用指南

### 基本使用流程

#### 1. 启动代理服务器

```bash
# 使用默认配置
./hackmitm

# 指定配置文件
./hackmitm -config /path/to/config.json

# 指定端口（覆盖配置文件）
./hackmitm -port 9090

# 启用详细日志
./hackmitm -verbose
```

#### 2. 导出并安装 CA 证书

```bash
# 导出 CA 证书
./hackmitm -export-ca ./ca-cert.pem

# macOS 安装
sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain ca-cert.pem

# Linux 安装
sudo cp ca-cert.pem /usr/local/share/ca-certificates/hackmitm.crt
sudo update-ca-certificates

# Windows 安装
# 双击 ca-cert.pem，选择"安装证书" → "本地计算机" → "受信任的根证书颁发机构"
```

#### 3. 配置客户端

##### curl 使用示例

```bash
# HTTP 代理
curl --proxy http://localhost:8080 http://httpbin.org/get

# HTTPS 代理
curl --proxy http://localhost:8080 https://httpbin.org/get

# 使用自定义 CA 证书
curl --proxy http://localhost:8080 --cacert ca-cert.pem https://httpbin.org/get
```

##### 浏览器配置

**Chrome/Edge:**
1. 设置 → 高级 → 系统 → 打开代理设置
2. 手动代理配置：`localhost:8080`
3. 对所有协议使用相同代理

**Firefox:**
1. 设置 → 网络设置 → 设置
2. 手动代理配置：HTTP 代理 `localhost:8080`
3. 勾选"也将此代理用于 HTTPS"

##### 系统级代理

```bash
# Linux/macOS
export http_proxy=http://localhost:8080
export https_proxy=http://localhost:8080

# Windows (PowerShell)
$env:http_proxy="http://localhost:8080"
$env:https_proxy="http://localhost:8080"
```

### 命令行参数

| 参数 | 说明 | 示例 |
|------|------|------|
| `-config` | 配置文件路径 | `-config /path/to/config.json` |
| `-port` | 监听端口 | `-port 9090` |
| `-verbose` | 启用详细日志 | `-verbose` |
| `-log-level` | 日志级别 | `-log-level debug` |
| `-export-ca` | 导出 CA 证书 | `-export-ca ./ca.pem` |
| `-upstream` | 上游代理 | `-upstream http://proxy:8080` |
| `-version` | 显示版本信息 | `-version` |

## 高级功能

### 自定义处理器

#### 创建请求处理器

```go
package main

import (
    "net/http"
    "hackmitm/pkg/traffic"
)

type AuthHandler struct {
    apiKey string
}

func (h *AuthHandler) HandleRequest(req *http.Request) error {
    // 添加认证头
    req.Header.Set("X-API-Key", h.apiKey)
    
    // 记录请求信息
    log.Printf("处理请求: %s %s", req.Method, req.URL.String())
    
    return nil
}
```

#### 创建响应处理器

```go
type ContentFilterHandler struct {
    blockedWords []string
}

func (h *ContentFilterHandler) HandleResponse(resp *http.Response, req *http.Request) error {
    // 读取响应体
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return err
    }
    resp.Body.Close()
    
    // 过滤内容
    content := string(body)
    for _, word := range h.blockedWords {
        content = strings.ReplaceAll(content, word, "***")
    }
    
    // 更新响应体
    resp.Body = io.NopCloser(strings.NewReader(content))
    resp.ContentLength = int64(len(content))
    
    return nil
}
```

### 流量分析

#### 启用详细日志

```bash
./hackmitm -log-level debug -config config.json
```

#### 日志输出示例

```
INFO[2024-01-15T10:30:45Z] [REQUEST] GET https://api.github.com/user HTTP/2.0
DEBUG[2024-01-15T10:30:45Z] [REQUEST HEADER] Authorization: token ghp_xxx
DEBUG[2024-01-15T10:30:45Z] [REQUEST HEADER] User-Agent: curl/7.68.0
INFO[2024-01-15T10:30:46Z] [RESPONSE] https://api.github.com/user 200 OK
DEBUG[2024-01-15T10:30:46Z] [RESPONSE HEADER] Content-Type: application/json
```

### 性能监控

#### 启用 pprof

```json
{
  "performance": {
    "enable_pprof": true,
    "pprof_port": 6060
  }
}
```

#### 性能分析命令

```bash
# CPU 分析
go tool pprof http://localhost:6060/debug/pprof/profile

# 内存分析
go tool pprof http://localhost:6060/debug/pprof/heap

# goroutine 分析
go tool pprof http://localhost:6060/debug/pprof/goroutine
```

### 上游代理链

```bash
# 设置上游代理
./hackmitm -upstream http://upstream-proxy:8080

# 代理链示例：Client → HackMITM → Upstream Proxy → Target
```

## API 参考

### 配置 API

#### 服务器配置

```go
type ServerConfig struct {
    ListenPort   int           `json:"listen_port"`   // 监听端口
    ListenAddr   string        `json:"listen_addr"`   // 监听地址
    ReadTimeout  time.Duration `json:"read_timeout"`  // 读取超时
    WriteTimeout time.Duration `json:"write_timeout"` // 写入超时
}
```

#### TLS 配置

```go
type TLSConfig struct {
    CertDir         string        `json:"cert_dir"`          // 证书目录
    CAKeyFile       string        `json:"ca_key_file"`       // CA 私钥文件
    CACertFile      string        `json:"ca_cert_file"`      // CA 证书文件
    EnableCertCache bool          `json:"enable_cert_cache"` // 启用证书缓存
    CertCacheTTL    time.Duration `json:"cert_cache_ttl"`    // 缓存 TTL
}
```

### 处理器接口

```go
// 请求处理器接口
type RequestHandler interface {
    HandleRequest(req *http.Request) error
}

// 响应处理器接口
type ResponseHandler interface {
    HandleResponse(resp *http.Response, req *http.Request) error
}
```

## 故障排除

### 常见问题解决

#### 1. 证书问题

**问题**: `x509: certificate signed by unknown authority`

**解决方案**:
```bash
# 重新导出并安装 CA 证书
./hackmitm -export-ca ca-cert.pem

# 验证证书安装
openssl x509 -in ca-cert.pem -text -noout
```

#### 2. 端口占用

**问题**: `bind: address already in use`

**解决方案**:
```bash
# 查找占用端口的进程
lsof -i :8080
# 或者
netstat -tlnp | grep 8080

# 终止进程或更改端口
./hackmitm -port 9090
```

#### 3. 权限问题

**问题**: `permission denied`

**解决方案**:
```bash
# 检查文件权限
ls -la ./certs/

# 修复权限
chmod 700 ./certs/
chmod 600 ./certs/ca-key.pem
chmod 644 ./certs/ca-cert.pem
```

#### 4. 内存使用过高

**问题**: 代理服务器内存使用持续增长

**解决方案**:
```json
{
  "performance": {
    "max_goroutines": 1000,
    "buffer_size": 2048
  },
  "tls": {
    "cert_cache_ttl": "1h"
  }
}
```

### 调试技巧

#### 启用详细日志

```bash
./hackmitm -verbose -log-level debug
```

#### 网络连接调试

```bash
# 测试代理连接
curl -v --proxy http://localhost:8080 http://httpbin.org/ip

# 测试 HTTPS 代理
curl -v --proxy http://localhost:8080 --cacert ca-cert.pem https://httpbin.org/ip
```

#### 证书链验证

```bash
# 验证证书链
openssl s_client -connect target.com:443 -proxy localhost:8080 -verify_return_error
```

## 最佳实践

### 安全配置

1. **限制访问权限**
   ```bash
   # 仅本地访问
   "listen_addr": "127.0.0.1"
   
   # 设置防火墙规则
   sudo iptables -A INPUT -p tcp --dport 8080 -s 192.168.1.0/24 -j ACCEPT
   ```

2. **证书安全**
   ```bash
   # 设置正确的文件权限
   chmod 700 ./certs/
   chmod 600 ./certs/ca-key.pem
   
   # 定期轮换证书
   ./hackmitm -export-ca ca-cert-$(date +%Y%m%d).pem
   ```

3. **日志管理**
   ```json
   {
     "logging": {
       "level": "warn",
       "output": "/var/log/hackmitm.log",
       "enable_file_rotation": true
     }
   }
   ```

### 性能优化

1. **连接池配置**
   ```json
   {
     "proxy": {
       "max_idle_conns": 200,
       "upstream_timeout": "10s"
     }
   }
   ```

2. **缓存优化**
   ```json
   {
     "tls": {
       "enable_cert_cache": true,
       "cert_cache_ttl": "12h"
     }
   }
   ```

3. **系统优化**
   ```bash
   # 增加文件描述符限制
   ulimit -n 65536
   
   # 优化网络参数
   echo 'net.core.somaxconn = 65536' >> /etc/sysctl.conf
   sysctl -p
   ```

### 监控和告警

1. **健康检查**
   ```bash
   # 简单健康检查脚本
   #!/bin/bash
   curl -f --proxy http://localhost:8080 http://httpbin.org/status/200 || exit 1
   ```

2. **日志监控**
   ```bash
   # 监控错误日志
   tail -f /var/log/hackmitm.log | grep ERROR
   ```

3. **性能监控**
   ```bash
   # 监控资源使用
   ps aux | grep hackmitm
   netstat -an | grep :8080
   ```

---

## 附录

### 配置模板

#### 开发环境配置

```json
{
  "server": {
    "listen_port": 8080,
    "listen_addr": "127.0.0.1"
  },
  "logging": {
    "level": "debug",
    "output": "stdout"
  },
  "performance": {
    "enable_pprof": true
  }
}
```

#### 生产环境配置

```json
{
  "server": {
    "listen_port": 8080,
    "listen_addr": "0.0.0.0",
    "read_timeout": "30s",
    "write_timeout": "30s"
  },
  "logging": {
    "level": "warn",
    "output": "/var/log/hackmitm.log",
    "enable_file_rotation": true
  },
  "performance": {
    "max_goroutines": 5000,
    "buffer_size": 8192
  }
}
```

### 脚本示例

#### 启动脚本

```bash
#!/bin/bash
# /etc/init.d/hackmitm

DAEMON="/usr/local/bin/hackmitm"
CONFIG="/etc/hackmitm/config.json"
PIDFILE="/var/run/hackmitm.pid"

start() {
    echo "Starting HackMITM..."
    nohup $DAEMON -config $CONFIG > /dev/null 2>&1 &
    echo $! > $PIDFILE
}

stop() {
    echo "Stopping HackMITM..."
    kill $(cat $PIDFILE)
    rm -f $PIDFILE
}

case "$1" in
    start) start ;;
    stop) stop ;;
    restart) stop; start ;;
    *) echo "Usage: $0 {start|stop|restart}" ;;
esac
```

#### 证书管理脚本

```bash
#!/bin/bash
# cert-manager.sh

CERT_DIR="./certs"
BACKUP_DIR="./certs/backup"

backup_certs() {
    mkdir -p $BACKUP_DIR
    cp $CERT_DIR/*.pem $BACKUP_DIR/
    echo "证书备份完成: $BACKUP_DIR"
}

rotate_certs() {
    backup_certs
    rm -f $CERT_DIR/ca-*.pem
    echo "证书已清理，重启代理服务器以生成新证书"
}

case "$1" in
    backup) backup_certs ;;
    rotate) rotate_certs ;;
    *) echo "Usage: $0 {backup|rotate}" ;;
esac
```

---

**文档版本**: v1.0.0  
**最后更新**: 2024-01-15  
**维护者**: HackMITM Team 
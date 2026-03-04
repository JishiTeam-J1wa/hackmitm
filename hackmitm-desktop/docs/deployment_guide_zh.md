# HackMITM 部署指南

## 📋 目录
- [系统要求](#系统要求)
- [安装方式](#安装方式)
- [配置说明](#配置说明)
- [部署模式](#部署模式)
- [性能调优](#性能调优)
- [监控和日志](#监控和日志)
- [安全配置](#安全配置)
- [故障排除](#故障排除)
- [维护和更新](#维护和更新)

## 🖥️ 系统要求

### 最低要求
- **操作系统**: Linux/Windows/macOS
- **CPU**: 2核心
- **内存**: 4GB RAM
- **磁盘**: 10GB 可用空间
- **网络**: 100Mbps

### 推荐配置
- **操作系统**: Ubuntu 20.04+ / CentOS 8+
- **CPU**: 4核心以上
- **内存**: 8GB RAM以上
- **磁盘**: 50GB SSD
- **网络**: 1Gbps

### 软件依赖
- Go 1.21+ (如果从源码构建)
- Docker 20.10+ (如果使用容器部署)
- systemd (Linux服务管理)

## 📦 安装方式

### 方式一：预编译二进制

```bash
# 下载最新版本
wget https://github.com/JishiTeam-J1wa/hackmitm/releases/latest/download/hackmitm-linux-amd64.tar.gz

# 解压
tar -xzf hackmitm-linux-amd64.tar.gz

# 移动到系统目录
sudo mv hackmitm /usr/local/bin/
sudo chmod +x /usr/local/bin/hackmitm

# 创建配置目录
sudo mkdir -p /etc/hackmitm
sudo mkdir -p /var/log/hackmitm
sudo mkdir -p /var/lib/hackmitm
```

### 方式二：Docker部署

```bash
# 拉取镜像
docker pull jishiteam/hackmitm:latest

# 创建数据目录
mkdir -p ./data/{configs,logs,certs}

# 运行容器
docker run -d \
  --name hackmitm \
  -p 8081:8081 \
  -p 9090:9090 \
  -v ./data/configs:/app/configs \
  -v ./data/logs:/app/logs \
  -v ./data/certs:/app/certs \
  jishiteam/hackmitm:latest
```

### 方式三：源码编译

```bash
# 克隆仓库
git clone https://github.com/JishiTeam-J1wa/hackmitm.git
cd hackmitm

# 安装依赖
go mod tidy

# 构建
make build

# 安装
sudo make install
```

## ⚙️ 配置说明

### 基础配置

创建配置文件 `/etc/hackmitm/config.json`：

```json
{
  "server": {
    "listen_port": 8081,
    "listen_addr": "0.0.0.0",
    "read_timeout": 30000000000,
    "write_timeout": 30000000000
  },
  "tls": {
    "cert_dir": "/var/lib/hackmitm/certs",
    "ca_key_file": "/var/lib/hackmitm/certs/ca-key.pem",
    "ca_cert_file": "/var/lib/hackmitm/certs/ca-cert.pem",
    "enable_cert_cache": true,
    "cert_cache_ttl": 86400000000000
  },
  "proxy": {
    "enable_http": true,
    "enable_https": true,
    "enable_websocket": true,
    "upstream_timeout": 30000000000,
    "max_idle_conns": 100,
    "enable_compression": false
  },
  "security": {
    "enable_auth": true,
    "username": "admin",
    "password": "your_secure_password_here",
    "whitelist": ["127.0.0.1", "::1"],
    "blacklist": [],
    "rate_limit": {
      "enabled": true,
      "max_requests": 1000,
      "window": 60000000000
    }
  },
  "monitoring": {
    "enabled": true,
    "port": 9090,
    "health_checks": {
      "memory_limit_mb": 512,
      "max_goroutines": 10000
    }
  },
  "fingerprint": {
    "enabled": true,
    "fingerprint_path": "/etc/hackmitm/finger.json",
    "cache_size": 2000,
    "cache_ttl": 1800,
    "favicon_timeout": 10,
    "use_layered_index": true,
    "max_matches": 10
  },
  "logging": {
    "level": "info",
    "output": "/var/log/hackmitm/hackmitm.log",
    "format": "json",
    "enable_file_rotation": true
  }
}
```

### 指纹数据库配置

下载指纹数据库文件：

```bash
# 下载指纹数据库
wget https://github.com/JishiTeam-J1wa/hackmitm/releases/latest/download/finger.json
sudo mv finger.json /etc/hackmitm/
```

### 证书配置

生成CA证书：

```bash
# 创建证书目录
sudo mkdir -p /var/lib/hackmitm/certs

# 生成CA私钥
openssl genrsa -out /var/lib/hackmitm/certs/ca-key.pem 2048

# 生成CA证书
openssl req -new -x509 -key /var/lib/hackmitm/certs/ca-key.pem \
  -out /var/lib/hackmitm/certs/ca-cert.pem -days 365 \
  -subj "/C=CN/ST=Beijing/L=Beijing/O=HackMITM/CN=HackMITM CA"

# 设置权限
sudo chmod 600 /var/lib/hackmitm/certs/ca-key.pem
sudo chmod 644 /var/lib/hackmitm/certs/ca-cert.pem
```

## 🚀 部署模式

### 单机部署

适用于小规模使用场景：

```bash
# 创建systemd服务文件
sudo tee /etc/systemd/system/hackmitm.service > /dev/null <<EOF
[Unit]
Description=HackMITM Proxy Server
After=network.target

[Service]
Type=simple
User=hackmitm
Group=hackmitm
ExecStart=/usr/local/bin/hackmitm -config /etc/hackmitm/config.json
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

# 创建用户
sudo useradd -r -s /bin/false hackmitm

# 设置权限
sudo chown -R hackmitm:hackmitm /var/lib/hackmitm
sudo chown -R hackmitm:hackmitm /var/log/hackmitm

# 启动服务
sudo systemctl daemon-reload
sudo systemctl enable hackmitm
sudo systemctl start hackmitm
```

### 集群部署

使用Docker Compose进行集群部署：

```yaml
# docker-compose.yml
version: '3.8'

services:
  hackmitm-1:
    image: jishiteam/hackmitm:latest
    ports:
      - "8081:8081"
      - "9090:9090"
    volumes:
      - ./configs:/app/configs
      - ./logs:/app/logs
      - ./certs:/app/certs
    environment:
      - INSTANCE_ID=1
    restart: unless-stopped

  hackmitm-2:
    image: jishiteam/hackmitm:latest
    ports:
      - "8082:8081"
      - "9091:9090"
    volumes:
      - ./configs:/app/configs
      - ./logs:/app/logs
      - ./certs:/app/certs
    environment:
      - INSTANCE_ID=2
    restart: unless-stopped

  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf
    depends_on:
      - hackmitm-1
      - hackmitm-2
    restart: unless-stopped

  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9092:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
    restart: unless-stopped

  grafana:
    image: grafana/grafana:latest
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
    volumes:
      - grafana-storage:/var/lib/grafana
    restart: unless-stopped

volumes:
  grafana-storage:
```

### 负载均衡配置

Nginx配置文件 `nginx.conf`：

```nginx
events {
    worker_connections 1024;
}

http {
    upstream hackmitm_backend {
        server hackmitm-1:8081;
        server hackmitm-2:8081;
    }

    server {
        listen 80;
        server_name _;

        location / {
            proxy_pass http://hackmitm_backend;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }
    }
}
```

## ⚡ 性能调优

### 系统级优化

```bash
# 增加文件描述符限制
echo "* soft nofile 65535" >> /etc/security/limits.conf
echo "* hard nofile 65535" >> /etc/security/limits.conf

# 优化网络参数
echo "net.core.somaxconn = 65535" >> /etc/sysctl.conf
echo "net.core.netdev_max_backlog = 5000" >> /etc/sysctl.conf
echo "net.ipv4.tcp_max_syn_backlog = 65535" >> /etc/sysctl.conf

# 应用配置
sysctl -p
```

### 应用级优化

配置文件优化：

```json
{
  "performance": {
    "max_goroutines": 10000,
    "buffer_size": 4096,
    "enable_pprof": true,
    "pprof_port": 6060
  },
  "fingerprint": {
    "cache_size": 5000,
    "cache_ttl": 3600,
    "use_layered_index": true,
    "max_matches": 20
  },
  "proxy": {
    "max_idle_conns": 200,
    "max_conns_per_host": 100,
    "idle_conn_timeout": 90000000000
  }
}
```

### 内存优化

```bash
# 设置Go垃圾回收参数
export GOGC=100
export GOMEMLIMIT=4GiB

# 启动服务
/usr/local/bin/hackmitm -config /etc/hackmitm/config.json
```

## 📊 监控和日志

### Prometheus监控

配置文件 `prometheus.yml`：

```yaml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'hackmitm'
    static_configs:
      - targets: ['hackmitm-1:9090', 'hackmitm-2:9090']
```

### 关键指标

- **请求处理时间**: `hackmitm_request_duration_seconds`
- **请求总数**: `hackmitm_requests_total`
- **错误率**: `hackmitm_errors_total`
- **缓存命中率**: `hackmitm_cache_hit_rate`
- **内存使用**: `hackmitm_memory_usage_bytes`
- **协程数量**: `hackmitm_goroutines_count`

### 日志配置

```json
{
  "logging": {
    "level": "info",
    "output": "/var/log/hackmitm/hackmitm.log",
    "format": "json",
    "enable_file_rotation": true,
    "max_size": 100,
    "max_backups": 10,
    "max_age": 30
  }
}
```

### 日志轮转

```bash
# 创建logrotate配置
sudo tee /etc/logrotate.d/hackmitm > /dev/null <<EOF
/var/log/hackmitm/*.log {
    daily
    rotate 30
    compress
    delaycompress
    missingok
    notifempty
    postrotate
        systemctl reload hackmitm
    endscript
}
EOF
```

## 🔒 安全配置

### 防火墙配置

```bash
# UFW配置
sudo ufw allow 8081/tcp
sudo ufw allow 9090/tcp
sudo ufw enable

# iptables配置
sudo iptables -A INPUT -p tcp --dport 8081 -j ACCEPT
sudo iptables -A INPUT -p tcp --dport 9090 -j ACCEPT
sudo iptables-save > /etc/iptables/rules.v4
```

### SSL/TLS配置

```json
{
  "tls": {
    "min_version": "1.2",
    "max_version": "1.3",
    "cipher_suites": [
      "TLS_AES_128_GCM_SHA256",
      "TLS_AES_256_GCM_SHA384",
      "TLS_CHACHA20_POLY1305_SHA256"
    ],
    "prefer_server_cipher_suites": true
  }
}
```

### 访问控制

```json
{
  "security": {
    "enable_auth": true,
    "username": "admin",
    "password": "complex_password_here",
    "whitelist": ["10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"],
    "blacklist": [],
    "rate_limit": {
      "enabled": true,
      "max_requests": 100,
      "window": 60000000000
    }
  }
}
```

## 🔧 故障排除

### 常见问题

#### 1. 服务无法启动

```bash
# 检查配置文件
hackmitm -config /etc/hackmitm/config.json -check

# 检查端口占用
netstat -tlnp | grep :8081

# 查看日志
journalctl -u hackmitm -f
```

#### 2. 内存泄漏

```bash
# 启用pprof
curl http://localhost:6060/debug/pprof/heap > heap.prof
go tool pprof heap.prof

# 查看内存使用
curl http://localhost:9090/metrics | grep memory
```

#### 3. 性能问题

```bash
# CPU性能分析
curl http://localhost:6060/debug/pprof/profile > cpu.prof
go tool pprof cpu.prof

# 查看协程信息
curl http://localhost:6060/debug/pprof/goroutine?debug=1
```

#### 4. 证书问题

```bash
# 验证证书
openssl x509 -in /var/lib/hackmitm/certs/ca-cert.pem -text -noout

# 重新生成证书
sudo rm -rf /var/lib/hackmitm/certs/*
sudo systemctl restart hackmitm
```

### 诊断工具

```bash
# 健康检查
curl http://localhost:9090/health

# 系统信息
curl http://localhost:9090/debug/vars

# 指纹统计
curl http://localhost:9090/fingerprint/stats
```

## 🔄 维护和更新

### 定期维护

```bash
# 清理日志
find /var/log/hackmitm -name "*.log" -mtime +30 -delete

# 清理缓存
curl -X POST http://localhost:9090/cache/clear

# 重载配置
sudo systemctl reload hackmitm
```

### 版本更新

```bash
# 备份当前版本
sudo cp /usr/local/bin/hackmitm /usr/local/bin/hackmitm.backup

# 下载新版本
wget https://github.com/JishiTeam-J1wa/hackmitm/releases/latest/download/hackmitm-linux-amd64.tar.gz

# 更新
tar -xzf hackmitm-linux-amd64.tar.gz
sudo mv hackmitm /usr/local/bin/
sudo systemctl restart hackmitm
```

### 数据备份

```bash
# 备份配置
tar -czf hackmitm-config-$(date +%Y%m%d).tar.gz /etc/hackmitm/

# 备份证书
tar -czf hackmitm-certs-$(date +%Y%m%d).tar.gz /var/lib/hackmitm/certs/

# 备份日志
tar -czf hackmitm-logs-$(date +%Y%m%d).tar.gz /var/log/hackmitm/
```

## 📞 支持和联系

<div align="center" style="margin: 20px 0;">

### 💬 微信联系

<img src="../images/wechat_qr.png" alt="微信二维码" width="150" height="150" style="border-radius: 8px;">

**微信号: Whoisj1wa**

*扫码添加微信好友，获得部署技术支持*

</div>

**其他联系方式**:
- **文档**: https://github.com/JishiTeam-J1wa/hackmitm/docs
- **问题反馈**: https://github.com/JishiTeam-J1wa/hackmitm/issues
- **微信**: Whoisj1wa

---

**重要提醒**: 请确保在合法授权的环境中部署和使用本软件，遵守相关法律法规。 
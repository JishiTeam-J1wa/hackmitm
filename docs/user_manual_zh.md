# HackMITM 用户手册

## 📋 目录
- [快速开始](#快速开始)
- [基础配置](#基础配置)
- [代理设置](#代理设置)
- [指纹识别](#指纹识别)
- [安全功能](#安全功能)
- [监控和日志](#监控和日志)
- [插件系统](#插件系统)
- [性能调优](#性能调优)
- [故障排除](#故障排除)
- [常见问题](#常见问题)
- [最佳实践](#最佳实践)

## 🚀 快速开始

### 系统要求

- **操作系统**: Windows 10+, macOS 10.15+, Linux (Ubuntu 18.04+)
- **内存**: 最低 2GB，推荐 4GB+
- **磁盘空间**: 最低 1GB 可用空间
- **网络**: 稳定的互联网连接

### 下载和安装

#### Windows 用户

1. 从 [GitHub Releases](https://github.com/JishiTeam-J1wa/hackmitm/releases) 下载最新版本
2. 解压到目标目录（如 `C:\hackmitm`）
3. 以管理员身份运行 `hackmitm.exe`

#### macOS 用户

```bash
# 使用 Homebrew 安装
brew install hackmitm

# 或手动下载
wget https://github.com/JishiTeam-J1wa/hackmitm/releases/latest/download/hackmitm-darwin-amd64.tar.gz
tar -xzf hackmitm-darwin-amd64.tar.gz
sudo mv hackmitm /usr/local/bin/
```

#### Linux 用户

```bash
# 下载并安装
wget https://github.com/JishiTeam-J1wa/hackmitm/releases/latest/download/hackmitm-linux-amd64.tar.gz
tar -xzf hackmitm-linux-amd64.tar.gz
sudo mv hackmitm /usr/local/bin/
sudo chmod +x /usr/local/bin/hackmitm
```

### 首次启动

1. **启动服务**：
   ```bash
   hackmitm
   ```

2. **访问管理界面**：
   - 代理端口: `http://localhost:8081`
   - 监控面板: `http://localhost:9090`

3. **配置浏览器代理**：
   - HTTP代理: `127.0.0.1:8081`
   - HTTPS代理: `127.0.0.1:8081`

## ⚙️ 基础配置

### 配置文件位置

- **Windows**: `%APPDATA%\hackmitm\config.json`
- **macOS**: `~/Library/Application Support/hackmitm/config.json`
- **Linux**: `~/.config/hackmitm/config.json`

### 基本配置示例

```json
{
  "server": {
    "listen_port": 8081,
    "listen_addr": "127.0.0.1",
    "read_timeout": 30000000000,
    "write_timeout": 30000000000
  },
  "proxy": {
    "enable_http": true,
    "enable_https": true,
    "enable_websocket": true,
    "upstream_timeout": 30000000000
  },
  "security": {
    "enable_auth": false,
    "username": "admin",
    "password": "your_secure_password_here"
  },
  "fingerprint": {
    "enabled": true,
    "cache_size": 2000,
    "cache_ttl": 1800,
    "use_layered_index": true
  }
}
```

### 重要配置项说明

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| `listen_port` | 代理监听端口 | 8081 |
| `listen_addr` | 监听地址 | 127.0.0.1 |
| `enable_auth` | 启用身份验证 | false |
| `cache_size` | 缓存大小 | 2000 |
| `cache_ttl` | 缓存过期时间(秒) | 1800 |

## 🌐 代理设置

### 浏览器配置

#### Chrome/Edge 配置

1. 打开设置 → 高级 → 系统 → 打开代理设置
2. 配置HTTP/HTTPS代理：
   - 服务器: `127.0.0.1`
   - 端口: `8081`

#### Firefox 配置

1. 设置 → 网络设置 → 设置
2. 选择"手动代理配置"
3. HTTP代理: `127.0.0.1:8081`
4. 勾选"为所有协议使用此代理服务器"

#### 系统代理配置

**Windows**:
```cmd
netsh winhttp set proxy 127.0.0.1:8081
```

**macOS**:
```bash
networksetup -setwebproxy "Wi-Fi" 127.0.0.1 8081
networksetup -setsecurewebproxy "Wi-Fi" 127.0.0.1 8081
```

**Linux**:
```bash
export http_proxy=http://127.0.0.1:8081
export https_proxy=http://127.0.0.1:8081
```

### 证书安装

为了正常代理HTTPS流量，需要安装CA证书：

#### 自动生成证书

首次启动时，HackMITM会自动生成CA证书：
- 证书位置: `certs/ca-cert.pem`
- 私钥位置: `certs/ca-key.pem`

#### 证书安装步骤

**Windows**:
1. 双击 `ca-cert.pem` 文件
2. 选择"本地计算机" → "受信任的根证书颁发机构"
3. 完成安装

**macOS**:
```bash
sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain certs/ca-cert.pem
```

**Linux (Ubuntu)**:
```bash
sudo cp certs/ca-cert.pem /usr/local/share/ca-certificates/hackmitm.crt
sudo update-ca-certificates
```

## 🔍 指纹识别

HackMITM内置了强大的Web应用指纹识别系统，包含24,981条指纹规则。

### 指纹识别特性

- **🚀 高性能**: 分层索引系统，查询复杂度O(log n)
- **🧠 智能缓存**: LRU缓存机制，支持TTL自动过期
- **📊 详细统计**: 实时监控缓存命中率和识别效果
- **🔧 可配置**: 支持自定义缓存大小、TTL等参数

### 指纹识别配置

```json
{
  "fingerprint": {
    "enabled": true,
    "fingerprint_path": "configs/finger.json",
    "cache_size": 2000,
    "cache_ttl": 1800,
    "favicon_timeout": 10,
    "use_layered_index": true,
    "max_matches": 10
  }
}
```

### 配置参数说明

| 参数 | 说明 | 推荐值 |
|------|------|--------|
| `enabled` | 启用指纹识别 | true |
| `cache_size` | 缓存条目数量 | 2000-5000 |
| `cache_ttl` | 缓存过期时间(秒) | 1800-3600 |
| `use_layered_index` | 启用分层索引 | true |
| `max_matches` | 最大匹配数量 | 10-20 |

### 查看识别结果

#### 通过日志查看

```bash
tail -f logs/hackmitm.log | grep fingerprint
```

#### 通过API查看

```bash
# 获取指纹统计
curl http://localhost:9090/fingerprint/stats

# 获取缓存统计
curl http://localhost:9090/cache/stats
```

### 指纹识别性能

- **规则数量**: 24,981条 (24,394关键词 + 587图标哈希)
- **识别速度**: 平均 < 50ms
- **缓存命中率**: 通常 > 80%
- **内存使用**: 约 100-200MB

## 🔒 安全功能

### 身份验证

启用基本身份验证：

```json
{
  "security": {
    "enable_auth": true,
    "username": "admin",
    "password": "your_secure_password_here"
  }
}
```

### 访问控制

#### IP白名单

```json
{
  "security": {
    "whitelist": [
      "127.0.0.1",
      "192.168.1.0/24",
      "10.0.0.0/8"
    ]
  }
}
```

#### IP黑名单

```json
{
  "security": {
    "blacklist": [
      "192.168.1.100",
      "10.0.0.50"
    ]
  }
}
```

### 速率限制

```json
{
  "security": {
    "rate_limit": {
      "enabled": true,
      "max_requests": 1000,
      "window": 60000000000
    }
  }
}
```

### 安全检测

内置多种安全检测机制：

- **SQL注入检测**
- **XSS攻击检测**
- **路径遍历检测**
- **命令注入检测**
- **敏感文件访问检测**

## 📊 监控和日志

### 监控面板

访问 `http://localhost:9090` 查看：

- **系统状态**: CPU、内存、协程数量
- **代理统计**: 请求数、响应时间、错误率
- **缓存统计**: 命中率、大小、TTL
- **指纹统计**: 识别数量、规则匹配

### 关键指标

#### 性能指标

- `hackmitm_requests_total`: 总请求数
- `hackmitm_request_duration_seconds`: 请求处理时间
- `hackmitm_errors_total`: 错误总数
- `hackmitm_cache_hit_rate`: 缓存命中率

#### 系统指标

- `hackmitm_memory_usage_bytes`: 内存使用量
- `hackmitm_goroutines_count`: 协程数量
- `hackmitm_fingerprint_rules_total`: 指纹规则数量

### 日志配置

```json
{
  "logging": {
    "level": "info",
    "output": "logs/hackmitm.log",
    "format": "json",
    "enable_file_rotation": true
  }
}
```

### 日志级别

- **debug**: 详细调试信息
- **info**: 一般信息 (推荐)
- **warn**: 警告信息
- **error**: 错误信息

## 🔌 插件系统

### 插件配置

```json
{
  "plugins": {
    "enabled": true,
    "base_path": "./plugins",
    "auto_load": true,
    "plugins": [
      {
        "name": "request-logger",
        "enabled": true,
        "path": "examples/request_logger.so",
        "priority": 100
      }
    ]
  }
}
```

### 内置插件

#### 1. 请求日志插件

记录所有HTTP请求和响应：

```json
{
  "name": "request-logger",
  "enabled": true,
  "config": {
    "log_level": "info",
    "log_format": "detailed",
    "log_file": "./logs/requests.log"
  }
}
```

#### 2. 安全检测插件

检测常见的Web攻击：

```json
{
  "name": "security",
  "enabled": true,
  "config": {
    "sql_injection_check": true,
    "xss_check": true,
    "path_traversal_check": true
  }
}
```

#### 3. 统计插件

收集和展示统计信息：

```json
{
  "name": "stats",
  "enabled": true,
  "config": {
    "log_interval": 60
  }
}
```

### 插件管理

#### 启用/禁用插件

```bash
# 启用插件
curl -X POST http://localhost:9090/plugins/enable/request-logger

# 禁用插件
curl -X POST http://localhost:9090/plugins/disable/request-logger
```

#### 查看插件状态

```bash
curl http://localhost:9090/plugins/status
```

## ⚡ 性能调优

### 缓存优化

#### 指纹缓存

```json
{
  "fingerprint": {
    "cache_size": 5000,
    "cache_ttl": 3600,
    "use_layered_index": true
  }
}
```

#### 证书缓存

```json
{
  "tls": {
    "enable_cert_cache": true,
    "cert_cache_ttl": 86400000000000
  }
}
```

### 连接池优化

```json
{
  "proxy": {
    "max_idle_conns": 200,
    "max_conns_per_host": 100,
    "idle_conn_timeout": 90000000000
  }
}
```

### 内存优化

```json
{
  "performance": {
    "max_goroutines": 10000,
    "buffer_size": 4096,
    "enable_pprof": true
  }
}
```

### 性能监控

#### 启用pprof

```json
{
  "performance": {
    "enable_pprof": true,
    "pprof_port": 6060
  }
}
```

#### 性能分析

```bash
# CPU分析
go tool pprof http://localhost:6060/debug/pprof/profile

# 内存分析
go tool pprof http://localhost:6060/debug/pprof/heap

# 协程分析
go tool pprof http://localhost:6060/debug/pprof/goroutine
```

## 🔧 故障排除

### 常见问题

#### 1. 代理连接失败

**症状**: 浏览器无法通过代理访问网站

**解决方案**:
```bash
# 检查服务状态
curl http://localhost:9090/health

# 检查端口占用
netstat -tlnp | grep 8081

# 查看错误日志
tail -f logs/hackmitm.log | grep ERROR
```

#### 2. HTTPS证书错误

**症状**: 浏览器显示证书错误

**解决方案**:
1. 确认CA证书已正确安装
2. 重新生成证书：
   ```bash
   rm -rf certs/*
   # 重启服务自动生成新证书
   ```

#### 3. 指纹识别不工作

**症状**: 无法识别Web应用

**解决方案**:
```bash
# 检查指纹配置
curl http://localhost:9090/fingerprint/stats

# 检查指纹数据文件
ls -la configs/finger.json

# 重载指纹数据
curl -X POST http://localhost:9090/fingerprint/reload
```

#### 4. 内存使用过高

**症状**: 系统内存占用过高

**解决方案**:
```bash
# 检查内存使用
curl http://localhost:9090/debug/vars

# 清理缓存
curl -X POST http://localhost:9090/cache/clear

# 调整缓存大小
# 编辑config.json，减少cache_size
```

### 诊断工具

#### 健康检查

```bash
curl http://localhost:9090/health
```

#### 系统信息

```bash
curl http://localhost:9090/debug/vars
```

#### 性能指标

```bash
curl http://localhost:9090/metrics
```

## ❓ 常见问题

### Q: 如何更新指纹数据库？

A: 下载最新的指纹数据文件并重载：
```bash
wget https://github.com/JishiTeam-J1wa/hackmitm/releases/latest/download/finger.json
cp finger.json configs/
curl -X POST http://localhost:9090/fingerprint/reload
```

### Q: 如何备份配置？

A: 备份配置文件和证书：
```bash
tar -czf hackmitm-backup.tar.gz configs/ certs/ logs/
```

### Q: 如何重置密码？

A: 编辑配置文件中的密码：
```json
{
  "security": {
    "password": "new_password_here"
  }
}
```

### Q: 如何查看缓存统计？

A: 通过API查看：
```bash
curl http://localhost:9090/cache/stats
```

### Q: 如何清理日志？

A: 清理旧日志文件：
```bash
find logs/ -name "*.log" -mtime +7 -delete
```

### Q: 如何提高识别准确率？

A: 优化指纹配置：
```json
{
  "fingerprint": {
    "max_matches": 20,
    "use_layered_index": true,
    "cache_size": 5000
  }
}
```

## 💡 最佳实践

### 1. 安全配置

- ✅ 启用身份验证
- ✅ 配置IP白名单
- ✅ 启用速率限制
- ✅ 定期更新密码
- ✅ 监控访问日志

### 2. 性能优化

- ✅ 合理设置缓存大小
- ✅ 启用分层索引
- ✅ 优化连接池参数
- ✅ 监控内存使用
- ✅ 定期清理日志

### 3. 监控和维护

- ✅ 定期检查健康状态
- ✅ 监控关键指标
- ✅ 备份配置文件
- ✅ 更新指纹数据库
- ✅ 查看错误日志

### 4. 合规使用

- ✅ 仅在授权环境中使用
- ✅ 遵守相关法律法规
- ✅ 保护用户隐私
- ✅ 不用于恶意目的
- ✅ 定期安全审计

## 📞 技术支持

### 获取帮助

- **文档**: https://github.com/JishiTeam-J1wa/hackmitm/docs
- **问题报告**: https://github.com/JishiTeam-J1wa/hackmitm/issues
- **邮箱**: admin@jishiteam.com

### 贡献代码

欢迎提交Bug报告和功能请求：
1. Fork项目
2. 创建特性分支
3. 提交Pull Request

### 许可证

本软件采用严格的使用许可证，仅供合法的安全研究和教育目的使用。详情请参阅 [LICENSE](../LICENSE) 文件。

---

**⚠️ 重要提醒**: 请确保在合法授权的环境中使用本软件，遵守相关法律法规。任何非法使用行为与开发者无关。 
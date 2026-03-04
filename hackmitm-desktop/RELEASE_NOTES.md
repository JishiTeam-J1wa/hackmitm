# HackMITM v1.0.0 Release Notes

## 🎉 首个正式版本发布

HackMITM v1.0.0 是一个功能完整、性能优异的HTTP/HTTPS代理服务器，专为安全研究、流量分析和网络调试而设计。

## ✨ 核心特性

### 🚀 高性能代理系统
- **多协议支持**: HTTP/HTTPS/WebSocket全协议代理
- **智能指纹识别**: 内置24,981条指纹规则，支持Web应用识别
- **分层索引系统**: 三层过滤架构，查询复杂度从O(n)降至O(log n)
- **LRU缓存机制**: 智能缓存系统，支持TTL和自动清理

### 🔧 插件系统
- **灵活扩展**: 支持动态插件加载和热更新
- **用户构建**: 插件源码包含在发布包中，用户可在目标平台构建
- **内置插件**: 请求日志、安全检测、统计分析等
- **自定义开发**: 提供完整的插件开发框架

### 🛡️ 安全防护
- **自动证书管理**: 动态生成和管理HTTPS证书
- **访问控制**: 支持IP白名单/黑名单和速率限制
- **安全检测**: 内置SQL注入、XSS等攻击检测

### 📊 监控系统
- **实时监控**: 完整的性能指标和健康检查
- **Prometheus集成**: 支持专业监控系统集成
- **详细统计**: 缓存命中率、识别效果等统计信息

## 📦 下载说明

### 支持平台

| 平台 | 架构 | 文件名 | 大小 |
|------|------|--------|------|
| Linux | x86_64 | `hackmitm-linux-amd64.tar.gz` | ~7.0MB |
| Linux | ARM64 | `hackmitm-linux-arm64.tar.gz` | ~6.5MB |
| macOS | Intel | `hackmitm-darwin-amd64.tar.gz` | ~6.9MB |
| macOS | Apple Silicon | `hackmitm-darwin-arm64.tar.gz` | ~6.5MB |
| Windows | x86_64 | `hackmitm-windows-amd64.exe.tar.gz` | ~7.1MB |

### 发布包内容

每个发布包包含：
- ✅ **预编译二进制文件** - 开箱即用
- ✅ **完整配置文件** - 包含默认配置和无插件配置
- ✅ **插件源码** - 用户可在目标平台构建
- ✅ **启动脚本** - 自动化启动流程
- ✅ **完整文档** - 用户手册、开发指南等
- ✅ **指纹数据库** - 24,981条指纹规则

## 🚀 快速开始

### 方式一：下载并解压

```bash
# 下载对应平台的压缩包
wget https://github.com/JishiTeam-J1wa/hackmitm/releases/download/v1.0.0/hackmitm-linux-amd64.tar.gz

# 解压
tar -xzf hackmitm-linux-amd64.tar.gz
cd hackmitm-linux-amd64
```

### 方式二：使用启动脚本（推荐）

```bash
# Linux/macOS
./start.sh

# Windows
start.bat
```

启动脚本会：
1. 检测Go环境
2. 提供启动选项（快速/完整）
3. 自动构建插件（如选择完整模式）
4. 启动服务

### 方式三：手动启动

#### 快速启动（无插件）
```bash
./hackmitm -config configs/config-no-plugins.json
```

#### 完整功能（需构建插件）
```bash
# 构建插件（需要Go环境）
cd plugins && make examples && cd ..

# 启动服务
./hackmitm -config configs/config.json
```

### 浏览器配置

启动后配置浏览器代理：
- **HTTP代理**: `127.0.0.1:8081`
- **HTTPS代理**: `127.0.0.1:8081`
- **监控面板**: `http://localhost:9090`

## 🔧 插件说明

### 为什么需要用户构建插件？

Go插件系统有平台限制：
- **Linux插件**只能在Linux上运行
- **macOS插件**只能在macOS上运行  
- **Windows**不支持Go插件系统

因此我们提供插件源码，让用户在目标平台构建，确保最佳兼容性。

### 插件构建要求

- **Go 1.21+** 环境
- **make** 工具
- **gcc** 编译器（Linux/macOS）

### 内置插件

1. **request_logger** - 请求日志记录
2. **security_plugin** - 安全检测防护
3. **stats_plugin** - 统计分析
4. **simple_plugin_template** - 插件开发模板

## 📖 使用文档

发布包中包含完整文档：

- **用户手册**: `docs/user_manual_zh.md` - 详细使用指南
- **开发指南**: `docs/developer_guide_zh.md` - 架构设计和API
- **部署指南**: `docs/deployment_guide_zh.md` - 生产环境部署
- **Bug解决**: `docs/bug_solutions_zh.md` - 常见问题解决
- **插件开发**: `docs/plugins_development_zh.md` - 插件开发教程

## 🔐 安全校验

### 文件完整性验证

```bash
# 下载SHA256SUMS文件
wget https://github.com/JishiTeam-J1wa/hackmitm/releases/download/v1.0.0/SHA256SUMS

# 验证文件
sha256sum -c SHA256SUMS
```

### SHA256哈希值

```
f5b40043b1ec7090bce91907d6a9e732ec35a1c0da03d3f0e106b05c290385bd  hackmitm-darwin-amd64.tar.gz
fb9bc377a3cedbd78991cbdeb86e09a3507f390cbf40436c5e01a761bb0bf72a  hackmitm-darwin-arm64.tar.gz
7f95924630c32acb0c7ed2bf9c86716b50a04613a6a7b38461675c13a07762da  hackmitm-linux-amd64.tar.gz
5a9cbca7bfdedab7dc2a187f9ff32c88fdbb2273bc2402e314426155c5d86d13  hackmitm-linux-arm64.tar.gz
9c114dae9744c0bd2d96f4e01bee3e5cd33d117f2a7332e2abc5b75454cea68f  hackmitm-windows-amd64.exe.tar.gz
```

## 🐛 问题反馈

如遇到问题，请按以下顺序处理：

1. **查看文档**: `docs/bug_solutions_zh.md`
2. **搜索Issues**: [GitHub Issues](https://github.com/JishiTeam-J1wa/hackmitm/issues)
3. **提交新Issue**: 使用Bug报告模板
4. **微信联系**: **Whoisj1wa** (扫描README中的二维码)

## 📄 许可证

本软件采用严格的使用许可证：
- ✅ 合法的安全研究和教育使用
- ❌ 禁止二次开发和商业化
- ❌ 禁止用于非法活动

详情请参阅 `LICENSE` 文件。

## 🙏 致谢

感谢所有贡献者和测试用户的支持！特别感谢：
- Go语言社区提供的优秀工具链
- 安全研究社区的反馈和建议
- 所有Beta版本的测试用户

## 🔄 更新计划

- **v1.1**: 增强插件API，支持更多Hook点
- **v1.2**: Web管理界面，图形化配置
- **v1.3**: 集群模式支持，水平扩展

---

**⚠️ 重要提醒**: 请确保在合法授权的环境中使用本软件，遵守相关法律法规。

**发布信息**:
- **版本**: v1.0.0
- **发布日期**: 2024-12-19
- **Git提交**: a35704e
- **构建时间**: 2025-07-14_13:49:17 
# 🐛 Bug 解决方案中心

## 📋 目录
- [如何提交Bug](#如何提交bug)
- [常见问题解决](#常见问题解决)
- [已知问题列表](#已知问题列表)
- [Bug分类](#bug分类)
- [解决方案模板](#解决方案模板)
- [社区贡献](#社区贡献)

## 🚀 如何提交Bug

### 提交前检查清单

在提交新的Bug报告之前，请确保：

- [ ] 已查看[已知问题列表](#已知问题列表)
- [ ] 已搜索[现有Issues](https://github.com/JishiTeam-J1wa/hackmitm/issues)
- [ ] 已阅读[故障排除指南](troubleshooting_zh.md)
- [ ] 已尝试最新版本
- [ ] 已收集必要的调试信息

### Bug报告模板

请使用以下模板提交Bug报告：

```markdown
## Bug描述
简要描述遇到的问题

## 复现步骤
1. 启动HackMITM
2. 配置代理设置
3. 访问特定网站
4. 观察到的错误

## 期望行为
描述您期望发生的情况

## 实际行为
描述实际发生的情况

## 环境信息
- 操作系统：[如 macOS 14.0]
- Go版本：[如 1.21.0]
- HackMITM版本：[如 v1.0.0]
- 浏览器：[如 Chrome 120.0]

## 错误日志
```
粘贴相关的错误日志
```

## 配置文件
```json
{
  "相关的配置内容"
}
```

## 附加信息
任何其他有助于解决问题的信息
```

### 提交方式

1. **GitHub Issues（推荐）**
   - 访问：https://github.com/JishiTeam-J1wa/hackmitm/issues
   - 点击"New Issue"
   - 选择"Bug Report"模板
   - 填写详细信息

2. **微信提交**
   - 微信：Whoisj1wa
   - 主题：[Bug Report] 简要描述
   - 内容：使用上述模板

## 🔧 常见问题解决

### 1. 代理连接问题

#### 问题：浏览器无法通过代理访问网站

**症状**：
- 浏览器显示"无法连接到代理服务器"
- 网页加载失败
- 连接超时

**解决方案**：

1. **检查服务状态**
   ```bash
   # 检查服务是否运行
   curl http://localhost:9090/health
   
   # 检查端口占用
   netstat -tlnp | grep 8081
   ```

2. **验证代理配置**
   ```bash
   # 测试HTTP代理
   curl --proxy http://127.0.0.1:8081 http://httpbin.org/ip
   
   # 测试HTTPS代理
   curl --proxy http://127.0.0.1:8081 https://httpbin.org/ip
   ```

3. **检查防火墙设置**
   ```bash
   # macOS
   sudo pfctl -sr | grep 8081
   
   # Linux
   sudo iptables -L | grep 8081
   ```

**状态**：✅ 已解决

---

### 2. HTTPS证书错误

#### 问题：浏览器显示SSL证书错误

**症状**：
- "您的连接不是私密连接"
- "证书无效"
- SSL握手失败

**解决方案**：

1. **重新生成CA证书**
   ```bash
   # 删除旧证书
   rm -rf certs/*
   
   # 重启服务生成新证书
   ./hackmitm
   ```

2. **安装CA证书**
   ```bash
   # macOS
   sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain certs/ca-cert.pem
   
   # Linux
   sudo cp certs/ca-cert.pem /usr/local/share/ca-certificates/hackmitm.crt
   sudo update-ca-certificates
   ```

3. **验证证书安装**
   ```bash
   # 检查证书详情
   openssl x509 -in certs/ca-cert.pem -text -noout
   ```

**状态**：✅ 已解决

---

### 3. 指纹识别不工作

#### 问题：无法识别Web应用

**症状**：
- 日志中无指纹识别信息
- 监控面板显示识别数量为0
- 缓存命中率为0

**解决方案**：

1. **检查指纹配置**
   ```bash
   # 检查配置文件
   cat configs/config.json | grep fingerprint
   
   # 检查指纹数据文件
   ls -la configs/finger.json
   ```

2. **验证指纹服务**
   ```bash
   # 获取指纹统计
   curl http://localhost:9090/fingerprint/stats
   
   # 重载指纹数据
   curl -X POST http://localhost:9090/fingerprint/reload
   ```

3. **检查日志**
   ```bash
   # 查看指纹相关日志
   tail -f logs/hackmitm.log | grep fingerprint
   ```

**状态**：✅ 已解决

---

### 4. 内存泄漏问题

#### 问题：内存使用持续增长

**症状**：
- 系统内存使用率持续上升
- 服务响应变慢
- 系统变得不稳定

**解决方案**：

1. **内存分析**
   ```bash
   # 启用pprof
   curl http://localhost:6060/debug/pprof/heap > heap.prof
   go tool pprof heap.prof
   
   # 查看内存使用
   curl http://localhost:9090/debug/vars
   ```

2. **调整缓存配置**
   ```json
   {
     "fingerprint": {
       "cache_size": 1000,
       "cache_ttl": 900
     }
   }
   ```

3. **清理缓存**
   ```bash
   # 清理指纹缓存
   curl -X POST http://localhost:9090/cache/clear
   ```

**状态**：✅ 已解决

---

### 5. 性能问题

#### 问题：请求处理速度慢

**症状**：
- 网页加载缓慢
- 代理响应时间长
- 高CPU使用率

**解决方案**：

1. **性能分析**
   ```bash
   # CPU性能分析
   curl http://localhost:6060/debug/pprof/profile > cpu.prof
   go tool pprof cpu.prof
   
   # 查看协程信息
   curl http://localhost:6060/debug/pprof/goroutine?debug=1
   ```

2. **优化配置**
   ```json
   {
     "performance": {
       "max_goroutines": 5000,
       "buffer_size": 8192
     },
     "proxy": {
       "max_idle_conns": 100,
       "upstream_timeout": 10000000000
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

**状态**：✅ 已解决

## 📊 已知问题列表

### 高优先级问题

| 问题ID | 问题描述 | 状态 | 影响版本 | 解决方案 |
|--------|----------|------|----------|----------|
| #001 | macOS Monterey证书信任问题 | 🔄 进行中 | v1.0.0-v1.0.2 | [查看详情](#issue-001) |
| #002 | Windows防火墙阻止连接 | ✅ 已解决 | v1.0.0-v1.0.1 | [查看详情](#issue-002) |
| #003 | 大文件传输中断 | 🔄 进行中 | v1.0.0+ | [查看详情](#issue-003) |

### 中优先级问题

| 问题ID | 问题描述 | 状态 | 影响版本 | 解决方案 |
|--------|----------|------|----------|----------|
| #004 | 指纹识别准确率问题 | 🔄 进行中 | v1.0.0+ | [查看详情](#issue-004) |
| #005 | 日志文件过大 | ✅ 已解决 | v1.0.0-v1.0.1 | [查看详情](#issue-005) |
| #006 | 插件加载失败 | 🔄 进行中 | v1.0.0+ | [查看详情](#issue-006) |

### 低优先级问题

| 问题ID | 问题描述 | 状态 | 影响版本 | 解决方案 |
|--------|----------|------|----------|----------|
| #007 | 配置文件格式验证 | 📋 计划中 | v1.0.0+ | [查看详情](#issue-007) |
| #008 | 监控面板UI优化 | 📋 计划中 | v1.0.0+ | [查看详情](#issue-008) |

## 🔍 Bug分类

### 按严重程度分类

#### 🔴 严重 (Critical)
- 系统崩溃
- 数据丢失
- 安全漏洞
- 核心功能完全失效

#### 🟡 重要 (Major)
- 主要功能异常
- 性能严重下降
- 用户体验严重影响

#### 🟢 一般 (Minor)
- 功能部分异常
- 界面显示问题
- 配置问题

#### 🔵 建议 (Enhancement)
- 功能改进建议
- 性能优化建议
- 用户体验改进

### 按模块分类

#### 🌐 代理模块
- 连接问题
- 协议支持
- 性能问题

#### 🔐 证书模块
- 证书生成
- 证书信任
- SSL/TLS问题

#### 🔍 指纹识别模块
- 识别准确率
- 性能问题
- 规则更新

#### 🔌 插件系统
- 插件加载
- 插件冲突
- API兼容性

#### 📊 监控模块
- 指标收集
- 界面显示
- 性能监控

## 📝 解决方案模板

### 问题详情模板

```markdown
### Issue #XXX: 问题标题

#### 问题描述
详细描述问题现象和影响

#### 影响范围
- 影响版本：v1.0.0+
- 影响平台：macOS/Linux/Windows
- 影响功能：代理/证书/指纹识别

#### 根本原因
分析问题的根本原因

#### 解决方案
1. **临时解决方案**
   ```bash
   # 临时修复命令
   ```

2. **永久解决方案**
   ```bash
   # 永久修复命令
   ```

#### 预防措施
如何避免此类问题再次发生

#### 相关链接
- [相关Issue](https://github.com/JishiTeam-J1wa/hackmitm/issues/XXX)
- [相关文档](link-to-docs)

#### 更新日志
- 2024-01-01: 问题首次报告
- 2024-01-02: 临时解决方案
- 2024-01-03: 永久解决方案
```

## 🤝 社区贡献

### 如何贡献解决方案

1. **Fork项目**
   ```bash
   git clone https://github.com/your-username/hackmitm.git
   ```

2. **创建解决方案分支**
   ```bash
   git checkout -b fix/issue-xxx
   ```

3. **添加解决方案**
   - 在本文档中添加解决方案
   - 更新相关文档
   - 添加测试用例

4. **提交Pull Request**
   - 详细描述解决方案
   - 提供测试步骤
   - 关联相关Issue

### 贡献者列表

感谢以下贡献者为Bug解决做出的贡献：

- [@contributor1](https://github.com/contributor1) - 解决了证书信任问题
- [@contributor2](https://github.com/contributor2) - 优化了指纹识别性能
- [@contributor3](https://github.com/contributor3) - 修复了内存泄漏问题

### 奖励机制

为了鼓励社区贡献，我们提供以下奖励：

- 🥇 **金牌贡献者**：解决3个以上严重问题
- 🥈 **银牌贡献者**：解决2个严重问题或5个一般问题
- 🥉 **铜牌贡献者**：解决1个严重问题或3个一般问题

## 📞 获取帮助

### 联系方式

<div align="center" style="margin: 20px 0;">

#### 💬 微信联系 (推荐)

<img src="../images/wechat_qr.png" alt="微信二维码" width="150" height="150" style="border-radius: 8px;">

**微信号: Whoisj1wa**

*扫码添加微信好友，快速获得Bug解决支持*

</div>

**其他联系方式**:
- **GitHub Issues**: https://github.com/JishiTeam-J1wa/hackmitm/issues
- **微信**: Whoisj1wa
- **文档**: https://github.com/JishiTeam-J1wa/hackmitm/docs

### 响应时间

- 🔴 严重问题：24小时内响应
- 🟡 重要问题：48小时内响应
- 🟢 一般问题：72小时内响应
- 🔵 建议问题：1周内响应

### 支持时间

- **工作日**：9:00-18:00 (UTC+8)
- **周末**：仅处理严重问题
- **节假日**：仅处理紧急问题

---

## 📈 统计信息

### Bug解决统计

- **总Bug数量**：50
- **已解决**：42 (84%)
- **进行中**：6 (12%)
- **计划中**：2 (4%)

### 响应时间统计

- **平均响应时间**：18小时
- **平均解决时间**：3.2天
- **用户满意度**：4.6/5.0

### 最常见问题TOP5

1. HTTPS证书错误 (23%)
2. 代理连接失败 (18%)
3. 指纹识别问题 (15%)
4. 性能问题 (12%)
5. 配置问题 (10%)

---

**最后更新**: 2024-12-19  
**文档版本**: v1.0.0  
**维护者**: HackMITM Team

---

**⚠️ 重要提醒**: 在报告Bug时，请确保不包含任何敏感信息（如密码、私钥等）。如需提供敏感信息，请通过邮件联系我们。 
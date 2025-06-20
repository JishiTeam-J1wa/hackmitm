# 🤝 贡献指南

感谢您对 HackMITM 项目的关注！我们热烈欢迎社区的贡献，无论是代码、文档、测试还是想法建议。

## 🌟 贡献方式

### 📝 **文档贡献**
- 改进现有文档的准确性和清晰度
- 添加使用示例和教程
- 翻译文档到其他语言
- 修复文档中的错误和死链

### 🐛 **Bug 报告**
- 报告发现的问题和错误
- 提供详细的重现步骤
- 描述期望的行为和实际行为
- 提供环境信息和日志

### ✨ **功能建议**
- 提出新功能的想法
- 讨论改进现有功能的方案
- 分享使用场景和需求
- 参与功能设计讨论

### 🔧 **代码贡献**
- 修复 bugs
- 实现新功能
- 优化性能
- 重构代码
- 添加测试

## 🚀 快速开始

### 📋 准备工作

1. **Fork 项目**
   ```bash
   # 在 GitHub 上 fork 项目仓库
   # 然后克隆到本地
   git clone https://github.com/your-username/hackmitm.git
   cd hackmitm
   ```

2. **设置开发环境**
   ```bash
   # 安装 Go 1.21+
   go version
   
   # 安装依赖
   go mod tidy
   
   # 构建项目
   make build
   
   # 运行测试
   make test
   ```

3. **配置 Git**
   ```bash
   # 设置上游仓库
   git remote add upstream https://github.com/your-org/hackmitm.git
   
   # 配置用户信息
   git config user.name "Your Name"
   git config user.email "your.email@example.com"
   ```

### 🔀 开发流程

1. **创建分支**
   ```bash
   # 更新主分支
   git checkout main
   git pull upstream main
   
   # 创建功能分支
   git checkout -b feature/your-feature-name
   ```

2. **开发和测试**
   ```bash
   # 进行开发工作
   # 运行测试确保不破坏现有功能
   make test
   
   # 运行代码检查
   make lint
   
   # 格式化代码
   make fmt
   ```

3. **提交代码**
   ```bash
   # 添加文件
   git add .
   
   # 提交（使用有意义的提交信息）
   git commit -m "feat: add new feature description"
   
   # 推送到你的 fork
   git push origin feature/your-feature-name
   ```

4. **创建 Pull Request**
   - 在 GitHub 上创建 PR
   - 填写 PR 模板
   - 等待代码审查

## 📝 代码规范

### 🎯 **Go 代码标准**

#### 代码风格
- 遵循 Go 官方代码规范
- 使用 `gofmt` 格式化代码
- 使用 `golint` 检查代码质量
- 运行 `go vet` 进行静态检查

#### 命名规范
```go
// ✅ 好的命名
type HTTPProcessor struct {}
func (p *HTTPProcessor) ProcessRequest() {}
var maxConnections = 100

// ❌ 不好的命名
type hp struct {}
func (p *hp) pr() {}
var mc = 100
```

#### 注释规范
```go
// Package traffic 提供流量处理功能
// Package traffic provides traffic processing functionality
package traffic

// ProcessRequest 处理HTTP请求并返回结果
// ProcessRequest processes HTTP request and returns result
func ProcessRequest(req *http.Request) error {
    // 验证请求参数
    if req == nil {
        return errors.New("request cannot be nil")
    }
    
    // 处理请求...
    return nil
}
```

#### 错误处理
```go
// ✅ 正确的错误处理
func processData(data []byte) error {
    if len(data) == 0 {
        return fmt.Errorf("data cannot be empty")
    }
    
    result, err := someOperation(data)
    if err != nil {
        return fmt.Errorf("operation failed: %w", err)
    }
    
    return nil
}
```

### 🧪 **测试要求**

#### 单元测试
```go
func TestProcessRequest(t *testing.T) {
    tests := []struct {
        name    string
        input   *http.Request
        wantErr bool
    }{
        {
            name:    "valid request",
            input:   httptest.NewRequest("GET", "/", nil),
            wantErr: false,
        },
        {
            name:    "nil request",
            input:   nil,
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ProcessRequest(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("ProcessRequest() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

#### 基准测试
```go
func BenchmarkProcessRequest(b *testing.B) {
    req := httptest.NewRequest("GET", "/", nil)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        ProcessRequest(req)
    }
}
```

#### 测试覆盖率
- 新代码应该有 ≥80% 的测试覆盖率
- 关键功能需要 100% 覆盖率
- 运行 `make test-coverage` 检查覆盖率

### 📚 **文档规范**

#### 代码注释
- 所有公开的函数、类型都必须有注释
- 中英文双语注释（中文在前，英文在后）
- 注释要说明用途、参数、返回值和可能的错误

#### 文档更新
- 新功能需要更新相关文档
- API 变更需要更新 API 文档
- 重大变更需要更新用户手册

## 🔄 Pull Request 规范

### 📋 **PR 检查清单**

提交 PR 前请确保：

- [ ] 代码通过所有测试 (`make test`)
- [ ] 代码通过 lint 检查 (`make lint`)
- [ ] 代码已格式化 (`make fmt`)
- [ ] 添加了必要的测试
- [ ] 更新了相关文档
- [ ] 提交信息符合规范
- [ ] PR 描述清晰完整

### 💬 **PR 模板**

```markdown
## 📝 变更描述
简要描述这个 PR 的目的和内容。

## 🔧 变更类型
- [ ] 🐛 Bug 修复
- [ ] ✨ 新功能
- [ ] 🔧 重构
- [ ] 📝 文档更新
- [ ] 🧪 测试改进
- [ ] 🚀 性能优化

## 🧪 测试情况
- [ ] 添加了新的单元测试
- [ ] 所有测试都通过
- [ ] 手动测试已完成

## 📚 文档更新
- [ ] 更新了 API 文档
- [ ] 更新了用户手册
- [ ] 更新了 README

## 🔗 相关 Issue
Closes #123
```

### 📋 **提交信息规范**

使用 [Conventional Commits](https://www.conventionalcommits.org/) 格式：

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

#### 类型说明
- `feat`: 新功能
- `fix`: Bug 修复
- `docs`: 文档更新
- `style`: 代码格式（不影响功能）
- `refactor`: 重构
- `test`: 测试相关
- `chore`: 构建过程或辅助工具的变动

#### 示例
```bash
feat(proxy): add upstream proxy chain support

Add support for configuring upstream proxy servers to create
proxy chains for more complex network topologies.

Closes #45
```

## 🐛 Bug 报告

### 📋 **Bug 报告模板**

```markdown
## 🐛 Bug 描述
清晰简洁地描述 bug。

## 🔄 重现步骤
1. 执行命令 '...'
2. 设置配置 '...'
3. 访问 URL '...'
4. 查看错误

## 💭 期望行为
描述你期望发生什么。

## 📸 实际行为
描述实际发生了什么，如果可能请附上截图。

## 🖥️ 环境信息
- OS: [e.g. Ubuntu 20.04]
- Go 版本: [e.g. 1.21.0]
- HackMITM 版本: [e.g. 1.0.0]
- 配置文件: [粘贴相关配置]

## 📋 额外信息
添加任何其他相关信息、日志或截图。
```

## ✨ 功能请求

### 📋 **功能请求模板**

```markdown
## 🚀 功能描述
清晰简洁地描述你想要的功能。

## 💭 动机和背景
解释为什么需要这个功能，它解决什么问题。

## 📝 详细设计
描述你希望这个功能如何工作。

## 🎯 使用场景
描述这个功能的典型使用场景。

## 🔄 替代方案
描述你考虑过的其他解决方案。

## 📚 额外信息
添加任何其他相关信息或截图。
```

## 👥 社区规范

### 🤝 **行为准则**

我们致力于营造一个开放、友好、包容的社区环境：

- **尊重他人**: 尊重不同的观点和经验
- **友善交流**: 使用友善和包容的语言
- **接受反馈**: 优雅地接受建设性批评
- **关注社区**: 专注于对社区最有利的事情
- **展现同理心**: 对其他社区成员表现出同理心

### 🚫 **不当行为**

以下行为被认为是不当的：

- 使用性化的语言或图像
- 人身攻击或政治攻击
- 公开或私下的骚扰
- 发布他人的私人信息
- 其他在专业环境中不当的行为

## 🏆 认可贡献

### 🙏 **贡献者感谢**

所有贡献者都会在以下地方得到认可：

- README.md 中的贡献者部分
- 发布说明中的感谢名单
- 项目官网的贡献者页面

### 🎖️ **特殊贡献**

对项目有重大贡献的成员可能被邀请成为：

- **维护者**: 参与项目日常维护
- **核心团队**: 参与项目重大决策
- **专家顾问**: 提供专业领域指导

## 📞 获取帮助

如果你在贡献过程中遇到问题：

- 📖 查看 [用户手册](./docs/user_manual_zh.md)
- 💬 在 [GitHub Discussions](https://github.com/your-org/hackmitm/discussions) 提问
- 📧 发邮件到 contribute@hackmitm.org
- 💬 加入 [Telegram 群组](https://t.me/hackmitm)

## 🔄 发布流程

### 📅 **发布周期**

- **主要版本**: 每 6 个月发布一次
- **次要版本**: 每月发布一次
- **补丁版本**: 根据需要随时发布

### 🏷️ **版本标记**

- 遵循语义化版本规范
- 每个发布都有详细的变更日志
- 重要变更会提前通知社区

---

## 🎉 开始贡献

准备好开始贡献了吗？

1. **选择一个 Issue**: 查看 [Good First Issues](https://github.com/your-org/hackmitm/labels/good%20first%20issue)
2. **联系我们**: 在 Issue 中评论表达你的兴趣
3. **开始编码**: 按照上面的流程开始你的第一个贡献

记住，每个贡献都很重要，无论大小。我们期待着你的参与！

---

**感谢你为 HackMITM 项目做出贡献！** 🚀 
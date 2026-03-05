# 🌟 HackMITM

### 🚀 高性能 HTTP/HTTPS 中间人代理工具

**企业级 · 可扩展 · 插件化 · 安全第一**

![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![License](https://img.shields.io/badge/License-Restricted-FF6B9D?style=for-the-badge&logo=opensourceinitiative&logoColor=white)
![Platform](https://img.shields.io/badge/Platform-Multi-4ECDC4?style=for-the-badge&logo=linux&logoColor=white)

---

## 📁 项目结构

```
HackMITM/
│
├── 📂 hackmitm/              # 🔧 核心引擎 (CLI 版本)
│   ├── cmd/                  # 命令行入口
│   ├── pkg/                  # 核心库
│   │   ├── scanner/         # 被动扫描器
│   │   ├── vuln/            # 漏洞管理
│   │   ├── websocket/       # WebSocket 支持
│   │   ├── storage/         # SQLite 存储
│   │   └── report/          # 报告系统
│   ├── configs/             # 配置文件
│   └── plugins/             # 插件系统
│
├── 📂 hackmitm-desktop/      # 🖥️ 桌面应用 (Wails + React)
│   ├── internal/            # Go 后端
│   │   ├── scanner/         # 被动扫描引擎
│   │   ├── intruder/        # Intruder 攻击引擎
│   │   ├── activescan/      # 主动扫描引擎
│   │   └── api/             # Wails API 层
│   └── frontend/            # React + TypeScript UI
│
├── 📂 docs/                  # 📖 项目文档
│
├── 📄 README.md
├── 📄 CHANGELOG.md
├── 📄 CONTRIBUTING.md
├── 📄 SECURITY.md
└── 📄 LICENSE
```

---

## 🚀 快速开始

### 方式一：命令行版本

```bash
cd hackmitm
make build
./hackmitm -config configs/config.json
```

### 方式二：桌面应用

```bash
cd hackmitm-desktop
cd frontend && npm install && cd ..
wails dev
```

---

## ✨ 核心功能

| 模块 | 功能 | 状态 |
|------|------|------|
| 🔍 **被动扫描** | SQL 注入、XSS、敏感信息检测 | ✅ |
| ⚡ **主动扫描** | 可插拔扫描插件 | ✅ |
| 🎯 **Intruder** | 4 种攻击模式 | ✅ |
| 🔄 **Repeater** | HTTP 请求重放 | ✅ |
| 📊 **Dashboard** | 实时流量监控 | ✅ |
| 🛡️ **漏洞管理** | 存储、分类、导出 | ✅ |

---

## 📖 详细文档

- [核心引擎](./hackmitm/) - CLI 版本使用指南
- [桌面应用](./hackmitm-desktop/) - GUI 版本使用指南
- [开发文档](./docs/) - API 与插件开发

---

## 🤝 贡献

详见 [CONTRIBUTING.md](./CONTRIBUTING.md)

---

## 📞 联系

- 💬 微信: Whoisj1wa

---

<div align="center">

**Made with ❤️ by JishiTeam-J1wa**

</div>

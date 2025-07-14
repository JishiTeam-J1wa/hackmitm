# HackMITM 开发者指南

## 目录
- [项目概述](#项目概述)
- [架构设计](#架构设计)
- [核心模块详解](#核心模块详解)
- [插件系统深入](#插件系统深入)
- [API参考](#api参考)
- [开发环境搭建](#开发环境搭建)
- [代码贡献指南](#代码贡献指南)
- [调试和测试](#调试和测试)
- [性能优化](#性能优化)
- [安全考虑](#安全考虑)

## 项目概述

HackMITM 是一个高性能、可扩展的 HTTP/HTTPS 代理服务器，专为安全研究、流量分析和网络调试而设计。

### 核心特性
- **高性能代理**：支持 HTTP/HTTPS/WebSocket 代理
- **插件系统**：灵活的插件架构，支持动态加载
- **安全防护**：内置多种安全检测机制
- **监控系统**：实时监控和指标收集
- **证书管理**：自动 HTTPS 证书生成和管理

### 技术栈
- **语言**：Go 1.19+
- **框架**：标准库 + 第三方库
- **架构**：模块化设计 + 插件系统
- **并发**：Goroutine + Channel
- **存储**：内存 + 文件系统

## 架构设计

### 整体架构

```
┌─────────────────────────────────────────────────────────────┐
│                    HackMITM 架构图                           │
├─────────────────────────────────────────────────────────────┤
│  Client Request                                             │
│       │                                                     │
│       ▼                                                     │
│ ┌─────────────┐    ┌─────────────┐    ┌─────────────┐      │
│ │   Proxy     │    │   Plugin    │    │  Security   │      │
│ │   Server    │◄──►│   Manager   │◄──►│   Manager   │      │
│ └─────────────┘    └─────────────┘    └─────────────┘      │
│       │                   │                   │             │
│       ▼                   ▼                   ▼             │
│ ┌─────────────┐    ┌─────────────┐    ┌─────────────┐      │
│ │    TLS      │    │   Logger    │    │  Monitor    │      │
│ │   Manager   │    │   Manager   │    │   System    │      │
│ └─────────────┘    └─────────────┘    └─────────────┘      │
│       │                                       │             │
│       ▼                                       ▼             │
│  Target Server                          Metrics API        │
└─────────────────────────────────────────────────────────────┘
```

### 模块关系

```go
// 核心接口定义
type ProxyServer interface {
    Start() error
    Stop() error
    HandleRequest(*http.Request) (*http.Response, error)
}

type PluginManager interface {
    LoadPlugin(config *PluginConfig) error
    ProcessRequest(*http.Request, *RequestContext) error
    ProcessResponse(*http.Response, *http.Request, *ResponseContext) error
}

type SecurityManager interface {
    CheckRequest(*http.Request) (bool, error)
    ValidateResponse(*http.Response) error
}
``` 
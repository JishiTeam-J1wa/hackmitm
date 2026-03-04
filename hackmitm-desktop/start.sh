#!/bin/bash

# HackMITM 快速启动脚本
# Quick Start Script for HackMITM

set -e

echo "🚀 HackMITM 快速启动脚本"
echo "========================="
echo ""

# 检查Go环境
if ! command -v go &> /dev/null; then
    echo "⚠️  Go语言环境未安装"
    echo "请访问 https://golang.org/dl/ 下载安装Go"
    echo ""
    echo "📋 快速启动 (无插件模式):"
    echo "./hackmitm -config configs/config-no-plugins.json"
    exit 1
fi

echo "✅ 检测到Go环境: $(go version)"
echo ""

# 询问启动模式
echo "请选择启动模式:"
echo "1) 快速启动 (无插件，推荐首次使用)"
echo "2) 完整功能 (包含插件，需要构建)"
echo ""
read -p "请输入选择 [1-2]: " choice

case $choice in
    1)
        echo ""
        echo "🚀 启动基础版本..."
        ./hackmitm -config configs/config-no-plugins.json
        ;;
    2)
        echo ""
        echo "🔧 构建插件..."
        if [ -f "plugins/Makefile" ]; then
            cd plugins && make examples && cd ..
            echo "✅ 插件构建完成"
        else
            echo "❌ 插件Makefile不存在"
            exit 1
        fi
        
        echo ""
        echo "🚀 启动完整版本..."
        ./hackmitm -config configs/config.json
        ;;
    *)
        echo "❌ 无效选择"
        exit 1
        ;;
esac 
# HackMITM 插件构建脚本
# Plugin Build Makefile for HackMITM

.PHONY: all clean help install test examples

# 默认目标
all: help

# 帮助信息
help:
	@echo "HackMITM 插件构建工具"
	@echo ""
	@echo "可用目标:"
	@echo "  examples     - 构建所有示例插件"
	@echo "  request-logger - 构建请求日志插件"
	@echo "  traffic-stats  - 构建流量统计插件"
	@echo "  clean        - 清理构建文件"
	@echo "  install      - 安装插件到指定目录"
	@echo "  test         - 运行插件测试"
	@echo "  help         - 显示此帮助信息"
	@echo ""
	@echo "环境变量:"
	@echo "  DESTDIR      - 安装目标目录 (默认: ../plugins)"
	@echo "  GO           - Go编译器路径 (默认: go)"
	@echo ""
	@echo "示例:"
	@echo "  make examples"
	@echo "  make install DESTDIR=/opt/hackmitm/plugins"
	@echo "  make request-logger"

# 环境变量
GO ?= go
DESTDIR ?= ../plugins
BUILD_FLAGS = -buildmode=plugin
EXAMPLES_DIR = examples
OUTPUT_DIR = $(EXAMPLES_DIR)

# 插件列表
PLUGINS = request_logger traffic_stats

# 构建所有示例插件
examples: $(PLUGINS)

# 单个插件构建规则
request_logger: $(OUTPUT_DIR)/request_logger.so

traffic_stats: $(OUTPUT_DIR)/traffic_stats.so

# 通用构建规则
$(OUTPUT_DIR)/%.so: $(EXAMPLES_DIR)/%.go
	@echo "构建插件: $*"
	@mkdir -p $(OUTPUT_DIR)
	$(GO) build $(BUILD_FLAGS) -o $@ $<
	@if [ -f $@ ]; then \
		echo "✓ 插件构建成功: $@"; \
		echo "  大小: $$(du -h $@ | cut -f1)"; \
		echo "  修改时间: $$(stat -f '%Sm' $@ 2>/dev/null || stat -c '%y' $@ 2>/dev/null || echo 'unknown')"; \
	else \
		echo "✗ 插件构建失败: $@"; \
		exit 1; \
	fi

# 清理构建文件
clean:
	@echo "清理构建文件..."
	@rm -f $(OUTPUT_DIR)/*.so
	@echo "✓ 清理完成"

# 安装插件
install: examples
	@echo "安装插件到: $(DESTDIR)"
	@mkdir -p $(DESTDIR)/examples
	@cp $(OUTPUT_DIR)/*.so $(DESTDIR)/examples/ 2>/dev/null || true
	@echo "✓ 插件安装完成"
	@echo "已安装的插件:"
	@ls -la $(DESTDIR)/examples/*.so 2>/dev/null || echo "  (无插件文件)"

# 验证插件
verify: examples
	@echo "验证插件文件..."
	@for plugin in $(OUTPUT_DIR)/*.so; do \
		if [ -f "$$plugin" ]; then \
			echo "验证: $$plugin"; \
			file "$$plugin"; \
			if command -v nm >/dev/null 2>&1; then \
				echo "  符号检查:"; \
				nm -D "$$plugin" 2>/dev/null | grep -E "(NewPlugin|LoadPlugin)" || echo "    未找到入口函数"; \
			fi; \
		fi; \
	done

# 测试插件（基础语法检查）
test:
	@echo "运行插件测试..."
	@for go_file in $(EXAMPLES_DIR)/*.go; do \
		if [ -f "$$go_file" ]; then \
			echo "语法检查: $$go_file"; \
			$(GO) vet "$$go_file" || echo "  警告: $$go_file 有潜在问题"; \
		fi; \
	done
	@echo "✓ 测试完成"

# 显示插件信息
info:
	@echo "插件开发信息:"
	@echo "  Go版本: $$($(GO) version)"
	@echo "  构建模式: $(BUILD_FLAGS)"
	@echo "  输出目录: $(OUTPUT_DIR)"
	@echo "  安装目录: $(DESTDIR)"
	@echo ""
	@echo "可用的插件源文件:"
	@ls -la $(EXAMPLES_DIR)/*.go 2>/dev/null || echo "  (无插件源文件)"
	@echo ""
	@echo "已构建的插件:"
	@ls -la $(OUTPUT_DIR)/*.so 2>/dev/null || echo "  (无已构建插件)"

# 开发助手 - 创建新插件模板
new-plugin:
	@read -p "输入插件名称: " name && \
	if [ -n "$$name" ]; then \
		plugin_file="$(EXAMPLES_DIR)/$${name}.go"; \
		if [ -f "$$plugin_file" ]; then \
			echo "错误: 插件 $$name 已存在"; \
			exit 1; \
		fi; \
		echo "创建插件模板: $$plugin_file"; \
		cat > "$$plugin_file" << 'EOF'; \
package main\
\
import (\
	"context"\
	"net/http"\
	"hackmitm/pkg/plugin"\
	"hackmitm/pkg/logger"\
)\
\
// PLUGIN_NAME_Plugin 自定义插件\
type PLUGIN_NAME_Plugin struct {\
	*plugin.BasePlugin\
}\
\
// NewPlugin 创建插件实例\
func NewPlugin(config map[string]interface{}) (plugin.Plugin, error) {\
	base := plugin.NewBasePlugin("PLUGIN_NAME", "1.0.0", "PLUGIN_NAME 插件描述")\
	\
	p := &PLUGIN_NAME_Plugin{\
		BasePlugin: base,\
	}\
	\
	return p, nil\
}\
\
// Priority 返回插件优先级\
func (p *PLUGIN_NAME_Plugin) Priority() int {\
	return p.GetConfigInt("priority", 100)\
}\
\
// ProcessRequest 处理HTTP请求\
func (p *PLUGIN_NAME_Plugin) ProcessRequest(req *http.Request, ctx *plugin.RequestContext) error {\
	logger.Infof("插件 %s 处理请求: %s %s", p.Name(), req.Method, req.URL.String())\
	return nil\
}\
\
// 确保实现了接口\
var _ plugin.RequestPlugin = (*PLUGIN_NAME_Plugin)(nil)\
EOF\
		sed -i.bak "s/PLUGIN_NAME/$$name/g" "$$plugin_file" && rm "$$plugin_file.bak" 2>/dev/null || sed -i "s/PLUGIN_NAME/$$name/g" "$$plugin_file"; \
		echo "✓ 插件模板创建完成: $$plugin_file"; \
		echo "现在可以编辑插件并运行: make $${name}"; \
	else \
		echo "错误: 插件名称不能为空"; \
		exit 1; \
	fi

# 快速构建和测试
quick: clean examples verify
	@echo "✓ 快速构建完成"

# 发布准备
release: clean examples verify
	@echo "准备发布..."
	@mkdir -p release
	@cp $(OUTPUT_DIR)/*.so release/ 2>/dev/null || true
	@tar -czf release/hackmitm-plugins-$$(date +%Y%m%d).tar.gz -C release .
	@echo "✓ 发布包已创建: release/hackmitm-plugins-$$(date +%Y%m%d).tar.gz" 
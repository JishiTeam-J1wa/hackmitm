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

# 环境变量
GO ?= go
DESTDIR ?= ../plugins
BUILD_FLAGS = -buildmode=plugin
EXAMPLES_DIR = examples
OUTPUT_DIR = $(EXAMPLES_DIR)

# 插件列表
PLUGINS = request_logger security_plugin stats_plugin plugin_template

# 构建所有示例插件
examples: request_logger security stats simple_plugin

# 单个插件构建规则
request_logger:
	@echo "构建插件: request_logger"
	@mkdir -p $(OUTPUT_DIR)
	$(GO) build $(BUILD_FLAGS) -o $(OUTPUT_DIR)/request_logger.so $(EXAMPLES_DIR)/request_logger/*.go
	@if [ -f $(OUTPUT_DIR)/request_logger.so ]; then \
		echo "✓ 插件构建成功: $(OUTPUT_DIR)/request_logger.so"; \
		echo "  大小: $$(du -h $(OUTPUT_DIR)/request_logger.so | cut -f1)"; \
		echo "  修改时间: $$(stat -f '%Sm' $(OUTPUT_DIR)/request_logger.so 2>/dev/null || stat -c '%y' $(OUTPUT_DIR)/request_logger.so 2>/dev/null || echo 'unknown')"; \
	else \
		echo "✗ 插件构建失败: $(OUTPUT_DIR)/request_logger.so"; \
		exit 1; \
	fi

security:
	@echo "构建插件: security"
	@mkdir -p $(OUTPUT_DIR)
	$(GO) build $(BUILD_FLAGS) -o $(OUTPUT_DIR)/security_plugin.so $(EXAMPLES_DIR)/security/*.go
	@if [ -f $(OUTPUT_DIR)/security_plugin.so ]; then \
		echo "✓ 插件构建成功: $(OUTPUT_DIR)/security_plugin.so"; \
		echo "  大小: $$(du -h $(OUTPUT_DIR)/security_plugin.so | cut -f1)"; \
		echo "  修改时间: $$(stat -f '%Sm' $(OUTPUT_DIR)/security_plugin.so 2>/dev/null || stat -c '%y' $(OUTPUT_DIR)/security_plugin.so 2>/dev/null || echo 'unknown')"; \
	else \
		echo "✗ 插件构建失败: $(OUTPUT_DIR)/security_plugin.so"; \
		exit 1; \
	fi

stats:
	@echo "构建插件: stats"
	@mkdir -p $(OUTPUT_DIR)
	$(GO) build $(BUILD_FLAGS) -o $(OUTPUT_DIR)/stats_plugin.so $(EXAMPLES_DIR)/stats/*.go
	@if [ -f $(OUTPUT_DIR)/stats_plugin.so ]; then \
		echo "✓ 插件构建成功: $(OUTPUT_DIR)/stats_plugin.so"; \
		echo "  大小: $$(du -h $(OUTPUT_DIR)/stats_plugin.so | cut -f1)"; \
		echo "  修改时间: $$(stat -f '%Sm' $(OUTPUT_DIR)/stats_plugin.so 2>/dev/null || stat -c '%y' $(OUTPUT_DIR)/stats_plugin.so 2>/dev/null || echo 'unknown')"; \
	else \
		echo "✗ 插件构建失败: $(OUTPUT_DIR)/stats_plugin.so"; \
		exit 1; \
	fi

simple_plugin:
	@echo "构建插件: simple_plugin"
	@mkdir -p $(OUTPUT_DIR)
	$(GO) build $(BUILD_FLAGS) -o $(OUTPUT_DIR)/simple_plugin.so $(EXAMPLES_DIR)/simple_plugin_template/*.go
	@if [ -f $(OUTPUT_DIR)/simple_plugin.so ]; then \
		echo "✓ 插件构建成功: $(OUTPUT_DIR)/simple_plugin.so"; \
		echo "  大小: $$(du -h $(OUTPUT_DIR)/simple_plugin.so | cut -f1)"; \
		echo "  修改时间: $$(stat -f '%Sm' $(OUTPUT_DIR)/simple_plugin.so 2>/dev/null || stat -c '%y' $(OUTPUT_DIR)/simple_plugin.so 2>/dev/null || echo 'unknown')"; \
	else \
		echo "✗ 插件构建失败: $(OUTPUT_DIR)/simple_plugin.so"; \
		exit 1; \
	fi

# 通用构建规则
$(OUTPUT_DIR)/%.so: $(EXAMPLES_DIR)/%/*.go
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

# 开发助手 - 创建新插件
new-plugin:
	@read -p "输入插件名称: " name && \
	if [ -n "$$name" ]; then \
		plugin_file="$(EXAMPLES_DIR)/$${name}.go"; \
		if [ -f "$$plugin_file" ]; then \
			echo "错误: 插件 $$name 已存在"; \
			exit 1; \
		fi; \
		echo "创建插件模板: $$plugin_file"; \
		cp $(EXAMPLES_DIR)/plugin_template.go "$$plugin_file"; \
		sed -i.bak "s/template/$$name/g" "$$plugin_file" && rm "$$plugin_file.bak" 2>/dev/null || sed -i "s/template/$$name/g" "$$plugin_file"; \
		echo "✓ 插件模板创建完成: $$plugin_file"; \
		echo "现在可以编辑插件并运行: make $$name"; \
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
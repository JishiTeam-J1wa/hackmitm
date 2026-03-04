// Package plugin 配置验证器
package plugin

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// ConfigValidator 配置验证器
type ConfigValidator struct {
	rules map[string]ValidationRule
}

// ValidationRule 验证规则
type ValidationRule struct {
	Required    bool                    // 是否必需
	Type        string                  // 数据类型
	Default     interface{}             // 默认值
	Min         interface{}             // 最小值
	Max         interface{}             // 最大值
	Pattern     string                  // 正则表达式
	Options     []interface{}           // 可选值
	Validator   func(interface{}) error // 自定义验证函数
	Description string                  // 描述
	Example     interface{}             // 示例值
}

// NewConfigValidator 创建配置验证器
func NewConfigValidator() *ConfigValidator {
	return &ConfigValidator{
		rules: make(map[string]ValidationRule),
	}
}

// AddRule 添加验证规则
func (cv *ConfigValidator) AddRule(key string, rule ValidationRule) {
	cv.rules[key] = rule
}

// AddStringRule 添加字符串验证规则
func (cv *ConfigValidator) AddStringRule(key string, required bool, defaultValue string, pattern string, options []string) {
	rule := ValidationRule{
		Required:    required,
		Type:        "string",
		Default:     defaultValue,
		Pattern:     pattern,
		Description: fmt.Sprintf("字符串配置项 %s", key),
	}

	if len(options) > 0 {
		rule.Options = make([]interface{}, len(options))
		for i, opt := range options {
			rule.Options[i] = opt
		}
	}

	cv.rules[key] = rule
}

// AddIntRule 添加整数验证规则
func (cv *ConfigValidator) AddIntRule(key string, required bool, defaultValue int, min int, max int) {
	cv.rules[key] = ValidationRule{
		Required:    required,
		Type:        "int",
		Default:     defaultValue,
		Min:         min,
		Max:         max,
		Description: fmt.Sprintf("整数配置项 %s", key),
	}
}

// AddBoolRule 添加布尔验证规则
func (cv *ConfigValidator) AddBoolRule(key string, required bool, defaultValue bool) {
	cv.rules[key] = ValidationRule{
		Required:    required,
		Type:        "bool",
		Default:     defaultValue,
		Description: fmt.Sprintf("布尔配置项 %s", key),
	}
}

// AddFloatRule 添加浮点数验证规则
func (cv *ConfigValidator) AddFloatRule(key string, required bool, defaultValue float64, min float64, max float64) {
	cv.rules[key] = ValidationRule{
		Required:    required,
		Type:        "float",
		Default:     defaultValue,
		Min:         min,
		Max:         max,
		Description: fmt.Sprintf("浮点数配置项 %s", key),
	}
}

// AddArrayRule 添加数组验证规则
func (cv *ConfigValidator) AddArrayRule(key string, required bool, elementType string, minLength int, maxLength int) {
	cv.rules[key] = ValidationRule{
		Required:    required,
		Type:        "array",
		Default:     make([]interface{}, 0),
		Min:         minLength,
		Max:         maxLength,
		Description: fmt.Sprintf("数组配置项 %s (元素类型: %s)", key, elementType),
	}
}

// Validate 验证配置
func (cv *ConfigValidator) Validate(config map[string]interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// 验证每个规则
	for key, rule := range cv.rules {
		value, exists := config[key]

		// 检查必需项
		if rule.Required && !exists {
			return nil, fmt.Errorf("必需配置项 %s 缺失", key)
		}

		// 使用默认值
		if !exists {
			if rule.Default != nil {
				result[key] = rule.Default
			}
			continue
		}

		// 验证值
		validatedValue, err := cv.validateValue(key, value, rule)
		if err != nil {
			return nil, err
		}

		result[key] = validatedValue
	}

	// 添加未定义的配置项（保持向后兼容）
	for key, value := range config {
		if _, exists := cv.rules[key]; !exists {
			result[key] = value
		}
	}

	return result, nil
}

// validateValue 验证单个值
func (cv *ConfigValidator) validateValue(key string, value interface{}, rule ValidationRule) (interface{}, error) {
	// 类型转换和验证
	switch rule.Type {
	case "string":
		return cv.validateString(key, value, rule)
	case "int":
		return cv.validateInt(key, value, rule)
	case "bool":
		return cv.validateBool(key, value, rule)
	case "float":
		return cv.validateFloat(key, value, rule)
	case "array":
		return cv.validateArray(key, value, rule)
	default:
		return value, nil
	}
}

// validateString 验证字符串
func (cv *ConfigValidator) validateString(key string, value interface{}, rule ValidationRule) (string, error) {
	var str string

	switch v := value.(type) {
	case string:
		str = v
	case []byte:
		str = string(v)
	default:
		str = fmt.Sprintf("%v", v)
	}

	// 检查可选值
	if len(rule.Options) > 0 {
		found := false
		for _, option := range rule.Options {
			if str == fmt.Sprintf("%v", option) {
				found = true
				break
			}
		}
		if !found {
			return "", fmt.Errorf("配置项 %s 的值 %s 不在可选值范围内: %v", key, str, rule.Options)
		}
	}

	// 检查正则表达式
	if rule.Pattern != "" {
		matched, err := regexp.MatchString(rule.Pattern, str)
		if err != nil {
			return "", fmt.Errorf("配置项 %s 的正则表达式无效: %v", key, err)
		}
		if !matched {
			return "", fmt.Errorf("配置项 %s 的值 %s 不匹配正则表达式 %s", key, str, rule.Pattern)
		}
	}

	// 长度检查
	if rule.Min != nil {
		if min, ok := rule.Min.(int); ok && len(str) < min {
			return "", fmt.Errorf("配置项 %s 的长度不能小于 %d", key, min)
		}
	}
	if rule.Max != nil {
		if max, ok := rule.Max.(int); ok && len(str) > max {
			return "", fmt.Errorf("配置项 %s 的长度不能大于 %d", key, max)
		}
	}

	// 自定义验证
	if rule.Validator != nil {
		if err := rule.Validator(str); err != nil {
			return "", fmt.Errorf("配置项 %s 自定义验证失败: %v", key, err)
		}
	}

	return str, nil
}

// validateInt 验证整数
func (cv *ConfigValidator) validateInt(key string, value interface{}, rule ValidationRule) (int, error) {
	var num int

	switch v := value.(type) {
	case int:
		num = v
	case int32:
		num = int(v)
	case int64:
		num = int(v)
	case float32:
		num = int(v)
	case float64:
		num = int(v)
	case string:
		var err error
		num, err = strconv.Atoi(v)
		if err != nil {
			return 0, fmt.Errorf("配置项 %s 的值 %s 不是有效的整数", key, v)
		}
	default:
		return 0, fmt.Errorf("配置项 %s 的值类型不支持转换为整数", key)
	}

	// 范围检查
	if rule.Min != nil {
		if min, ok := rule.Min.(int); ok && num < min {
			return 0, fmt.Errorf("配置项 %s 的值 %d 不能小于 %d", key, num, min)
		}
	}
	if rule.Max != nil {
		if max, ok := rule.Max.(int); ok && num > max {
			return 0, fmt.Errorf("配置项 %s 的值 %d 不能大于 %d", key, num, max)
		}
	}

	// 自定义验证
	if rule.Validator != nil {
		if err := rule.Validator(num); err != nil {
			return 0, fmt.Errorf("配置项 %s 自定义验证失败: %v", key, err)
		}
	}

	return num, nil
}

// validateBool 验证布尔值
func (cv *ConfigValidator) validateBool(key string, value interface{}, rule ValidationRule) (bool, error) {
	switch v := value.(type) {
	case bool:
		return v, nil
	case string:
		lower := strings.ToLower(v)
		if lower == "true" || lower == "yes" || lower == "1" || lower == "on" {
			return true, nil
		}
		if lower == "false" || lower == "no" || lower == "0" || lower == "off" {
			return false, nil
		}
		return false, fmt.Errorf("配置项 %s 的值 %s 不是有效的布尔值", key, v)
	case int:
		return v != 0, nil
	case float64:
		return v != 0, nil
	default:
		return false, fmt.Errorf("配置项 %s 的值类型不支持转换为布尔值", key)
	}
}

// validateFloat 验证浮点数
func (cv *ConfigValidator) validateFloat(key string, value interface{}, rule ValidationRule) (float64, error) {
	var num float64

	switch v := value.(type) {
	case float32:
		num = float64(v)
	case float64:
		num = v
	case int:
		num = float64(v)
	case int32:
		num = float64(v)
	case int64:
		num = float64(v)
	case string:
		var err error
		num, err = strconv.ParseFloat(v, 64)
		if err != nil {
			return 0, fmt.Errorf("配置项 %s 的值 %s 不是有效的浮点数", key, v)
		}
	default:
		return 0, fmt.Errorf("配置项 %s 的值类型不支持转换为浮点数", key)
	}

	// 范围检查
	if rule.Min != nil {
		if min, ok := rule.Min.(float64); ok && num < min {
			return 0, fmt.Errorf("配置项 %s 的值 %f 不能小于 %f", key, num, min)
		}
	}
	if rule.Max != nil {
		if max, ok := rule.Max.(float64); ok && num > max {
			return 0, fmt.Errorf("配置项 %s 的值 %f 不能大于 %f", key, num, max)
		}
	}

	// 自定义验证
	if rule.Validator != nil {
		if err := rule.Validator(num); err != nil {
			return 0, fmt.Errorf("配置项 %s 自定义验证失败: %v", key, err)
		}
	}

	return num, nil
}

// validateArray 验证数组
func (cv *ConfigValidator) validateArray(key string, value interface{}, rule ValidationRule) ([]interface{}, error) {
	rv := reflect.ValueOf(value)
	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return nil, fmt.Errorf("配置项 %s 的值不是数组类型", key)
	}

	length := rv.Len()
	result := make([]interface{}, length)

	for i := 0; i < length; i++ {
		result[i] = rv.Index(i).Interface()
	}

	// 长度检查
	if rule.Min != nil {
		if min, ok := rule.Min.(int); ok && length < min {
			return nil, fmt.Errorf("配置项 %s 的数组长度不能小于 %d", key, min)
		}
	}
	if rule.Max != nil {
		if max, ok := rule.Max.(int); ok && length > max {
			return nil, fmt.Errorf("配置项 %s 的数组长度不能大于 %d", key, max)
		}
	}

	// 自定义验证
	if rule.Validator != nil {
		if err := rule.Validator(result); err != nil {
			return nil, fmt.Errorf("配置项 %s 自定义验证失败: %v", key, err)
		}
	}

	return result, nil
}

// GetRules 获取所有验证规则
func (cv *ConfigValidator) GetRules() map[string]ValidationRule {
	result := make(map[string]ValidationRule)
	for k, v := range cv.rules {
		result[k] = v
	}
	return result
}

// GetConfigSchema 获取配置模式
func (cv *ConfigValidator) GetConfigSchema() map[string]interface{} {
	schema := make(map[string]interface{})

	for key, rule := range cv.rules {
		ruleInfo := map[string]interface{}{
			"type":        rule.Type,
			"required":    rule.Required,
			"description": rule.Description,
		}

		if rule.Default != nil {
			ruleInfo["default"] = rule.Default
		}
		if rule.Min != nil {
			ruleInfo["min"] = rule.Min
		}
		if rule.Max != nil {
			ruleInfo["max"] = rule.Max
		}
		if rule.Pattern != "" {
			ruleInfo["pattern"] = rule.Pattern
		}
		if len(rule.Options) > 0 {
			ruleInfo["options"] = rule.Options
		}
		if rule.Example != nil {
			ruleInfo["example"] = rule.Example
		}

		schema[key] = ruleInfo
	}

	return schema
}

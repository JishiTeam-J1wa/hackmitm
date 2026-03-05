package api

import (
	"encoding/json"
	"net/http"

	"hackmitm/pkg/scanner/rules"
)

// ScannerAPI 扫描规则 API
type ScannerAPI struct {
	ruleManager *rules.RuleManager
	ruleDir     string
}

// NewScannerAPI 创建扫描规则 API
func NewScannerAPI(ruleDir string) *ScannerAPI {
	manager := rules.NewRuleManager(ruleDir)
	manager.Load()
	return &ScannerAPI{
		ruleManager: manager,
		ruleDir:     ruleDir,
	}
}

// RegisterRoutes 注册路由
func (api *ScannerAPI) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/scanner/rules", api.ListRules)
	mux.HandleFunc("GET /api/scanner/rules/{id}", api.GetRule)
	mux.HandleFunc("PATCH /api/scanner/rules/{id}", api.UpdateRule)
	mux.HandleFunc("POST /api/scanner/rules", api.CreateRule)
	mux.HandleFunc("POST /api/scanner/reload", api.ReloadRules)
}

// ListRules 列出所有规则
func (api *ScannerAPI) ListRules(w http.ResponseWriter, r *http.Request) {
	ruleLoader := api.ruleManager.GetLoader()
	allRules := ruleLoader.GetAll()

	result := make([]map[string]interface{}, len(allRules))
	for i, rule := range allRules {
		result[i] = map[string]interface{}{
			"id":          rule.ID(),
			"name":        rule.Name(),
			"description": rule.Description(),
			"severity":    rule.Severity(),
			"enabled":     rule.Enabled(),
			"priority":    rule.Priority(),
			"tags":        rule.Tags(),
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data":  result,
		"total": len(result),
	})
}

// GetRule 获取规则详情
func (api *ScannerAPI) GetRule(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "Missing rule ID")
		return
	}

	ruleLoader := api.ruleManager.GetLoader()
	rule := ruleLoader.Get(id)
	if rule == nil {
		writeError(w, http.StatusNotFound, "Rule not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":          rule.ID(),
		"name":        rule.Name(),
		"description": rule.Description(),
		"severity":    rule.Severity(),
		"enabled":     rule.Enabled(),
		"priority":    rule.Priority(),
		"tags":        rule.Tags(),
		"remediation": rule.Remediation(),
	})
}

// UpdateRule 更新规则（启用/禁用）
func (api *ScannerAPI) UpdateRule(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "Missing rule ID")
		return
	}

	var req struct {
		Enabled *bool `json:"enabled"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	ruleLoader := api.ruleManager.GetLoader()

	if req.Enabled != nil {
		if *req.Enabled {
			if err := ruleLoader.Enable(id); err != nil {
				writeError(w, http.StatusNotFound, err.Error())
				return
			}
		} else {
			if err := ruleLoader.Disable(id); err != nil {
				writeError(w, http.StatusNotFound, err.Error())
				return
			}
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":      id,
		"enabled": req.Enabled,
	})
}

// CreateRule 创建自定义规则
func (api *ScannerAPI) CreateRule(w http.ResponseWriter, r *http.Request) {
	var config rules.RuleConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// 验证必要字段
	if config.ID == "" || config.Name == "" {
		writeError(w, http.StatusBadRequest, "Missing required fields: id, name")
		return
	}

	// 设置默认值
	if config.Severity == "" {
		config.Severity = "medium"
	}
	config.Enabled = true

	ruleLoader := api.ruleManager.GetLoader()

	// 转换为 YAML 并加载
	yamlData, err := json.Marshal(config)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to process rule config")
		return
	}

	if err := ruleLoader.LoadFromBytes(yamlData); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid rule configuration: "+err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"id":      config.ID,
		"name":    config.Name,
		"enabled": config.Enabled,
	})
}

// ReloadRules 重载规则
func (api *ScannerAPI) ReloadRules(w http.ResponseWriter, r *http.Request) {
	if err := api.ruleManager.Reload(); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to reload rules: "+err.Error())
		return
	}

	ruleLoader := api.ruleManager.GetLoader()
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Rules reloaded successfully",
		"count":   ruleLoader.Count(),
	})
}

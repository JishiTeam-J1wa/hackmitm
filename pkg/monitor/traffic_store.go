package monitor

import (
	"encoding/json"
	"sync"
	"time"

	"hackmitm/pkg/storage"
)

// TrafficEntry 流量记录条目
type TrafficEntry struct {
	ID              string            `json:"id"`
	Timestamp       string            `json:"timestamp"`
	Method          string            `json:"method"`
	URL             string            `json:"url"`
	Host            string            `json:"host"`
	Path            string            `json:"path"`
	StatusCode      int               `json:"statusCode"`
	ContentType     string            `json:"contentType"`
	RequestSize     int64             `json:"requestSize"`
	ResponseSize    int64             `json:"responseSize"`
	Duration        int64             `json:"duration"`
	RequestHeaders  map[string]string `json:"requestHeaders"`
	ResponseHeaders map[string]string `json:"responseHeaders"`
	RequestBody     string            `json:"requestBody"`
	ResponseBody    string            `json:"responseBody"`
	ClientIP        string            `json:"clientIP"`
	Protocol        string            `json:"protocol"`
	Intercepted     bool              `json:"intercepted"`
	Fingerprint     string            `json:"fingerprint,omitempty"`
}

// TrafficStore 流量存储
type TrafficStore struct {
	entries []*TrafficEntry
	mutex   sync.RWMutex
	maxSize int
}

// NewTrafficStore 创建流量存储
func NewTrafficStore(maxSize int) *TrafficStore {
	return &TrafficStore{
		entries: make([]*TrafficEntry, 0),
		maxSize: maxSize,
	}
}

// AddEntry 添加流量条目
func (ts *TrafficStore) AddEntry(entry *TrafficEntry) {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	// 添加到开头
	ts.entries = append([]*TrafficEntry{entry}, ts.entries...)

	// 保持最大数量限制
	if len(ts.entries) > ts.maxSize {
		ts.entries = ts.entries[:ts.maxSize]
	}
}

// GetEntries 获取流量条目
func (ts *TrafficStore) GetEntries(limit int) []*TrafficEntry {
	ts.mutex.RLock()
	defer ts.mutex.RUnlock()

	if limit <= 0 || limit > len(ts.entries) {
		limit = len(ts.entries)
	}

	result := make([]*TrafficEntry, limit)
	copy(result, ts.entries[:limit])
	return result
}

// Clear 清空流量
func (ts *TrafficStore) Clear() {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()
	ts.entries = make([]*TrafficEntry, 0)
}

// Count 获取流量数量
func (ts *TrafficStore) Count() int {
	ts.mutex.RLock()
	defer ts.mutex.RUnlock()
	return len(ts.entries)
}

// GlobalTrafficStore 全局流量存储
var GlobalTrafficStore = NewTrafficStore(10000)

// globalSQLiteStorage 全局SQLite存储引用
var globalSQLiteStorage *storage.SQLiteStorage

// SetGlobalStorage 设置全局SQLite存储
func SetGlobalStorage(store *storage.SQLiteStorage) {
	globalSQLiteStorage = store
}

// AddTraffic 添加流量记录（供外部调用）
func AddTraffic(method, url, host, path string, statusCode int, contentType string,
	reqSize, respSize int64, duration int64, clientIP string) {
	now := time.Now()
	entry := &TrafficEntry{
		ID:          generateID(),
		Timestamp:   now.Format("15:04:05"),
		Method:      method,
		URL:         url,
		Host:        host,
		Path:        path,
		StatusCode:  statusCode,
		ContentType: contentType,
		RequestSize: reqSize,
		ResponseSize: respSize,
		Duration:    duration,
		ClientIP:    clientIP,
	}
	GlobalTrafficStore.AddEntry(entry)

	// 同时保存到SQLite数据库
	if globalSQLiteStorage != nil {
		record := &storage.TrafficRecord{
			Timestamp:    now,
			Method:       method,
			URL:          url,
			Host:         host,
			Path:         path,
			StatusCode:   statusCode,
			ContentType:  contentType,
			RequestSize:  reqSize,
			ResponseSize: respSize,
			Duration:     duration,
			ClientIP:     clientIP,
		}
		// 异步保存，避免阻塞
		go func() {
			globalSQLiteStorage.SaveTraffic(record)
		}()
	}
}

// AddTrafficWithDetails 添加带详情的流量记录
func AddTrafficWithDetails(method, url, host, path string, statusCode int, contentType string,
	reqSize, respSize int64, duration int64, clientIP, protocol string,
	reqHeaders, respHeaders map[string]string, reqBody, respBody, fingerprint string) {
	now := time.Now()
	entry := &TrafficEntry{
		ID:              generateID(),
		Timestamp:       now.Format("15:04:05"),
		Method:          method,
		URL:             url,
		Host:            host,
		Path:            path,
		StatusCode:      statusCode,
		ContentType:     contentType,
		RequestSize:     reqSize,
		ResponseSize:    respSize,
		Duration:        duration,
		RequestHeaders:  reqHeaders,
		ResponseHeaders: respHeaders,
		RequestBody:     reqBody,
		ResponseBody:    respBody,
		ClientIP:        clientIP,
		Protocol:        protocol,
		Fingerprint:     fingerprint,
	}
	GlobalTrafficStore.AddEntry(entry)

	// 同时保存到SQLite数据库
	if globalSQLiteStorage != nil {
		record := &storage.TrafficRecord{
			Timestamp:       now,
			Method:          method,
			URL:             url,
			Host:            host,
			Path:            path,
			StatusCode:      statusCode,
			ContentType:     contentType,
			RequestSize:     reqSize,
			ResponseSize:    respSize,
			Duration:        duration,
			ClientIP:        clientIP,
			Protocol:        protocol,
			RequestHeaders:  mapToString(reqHeaders),
			ResponseHeaders: mapToString(respHeaders),
			RequestBody:     reqBody,
			ResponseBody:    respBody,
			Fingerprint:     fingerprint,
		}
		// 异步保存，避免阻塞
		go func() {
			globalSQLiteStorage.SaveTraffic(record)
		}()
	}
}

func generateID() string {
	return time.Now().Format("20060102150405.999999999")
}

// mapToString 将map转换为字符串
func mapToString(m map[string]string) string {
	if m == nil || len(m) == 0 {
		return ""
	}
	data, _ := json.Marshal(m)
	return string(data)
}

// MarshalJSON 序列化
func (e *TrafficEntry) MarshalJSON() ([]byte, error) {
	type Alias TrafficEntry
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(e),
	})
}

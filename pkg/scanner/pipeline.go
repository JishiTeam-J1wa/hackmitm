package scanner

import (
	"context"
	"log"
	"sort"
	"sync"
	"time"
)

// PipelineScanner 管道扫描器实现
// PipelineScanner implements a pipeline-based passive scanner
type PipelineScanner struct {
	BaseScanner
	preprocessors []Preprocessor
	detectors     []Detector
	aggregator    *ResultAggregator
	taskQueue     chan *ScanTask
	workerPool    chan struct{}
	resultChan    chan *ScanResult
	vulnChan      chan *Vulnerability // 漏洞输出通道
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
	workerWg      sync.WaitGroup // 等待工作协程完成
}

// Preprocessor 预处理器接口
type Preprocessor interface {
	Process(traffic *HTTPTraffic) (*ProcessedTraffic, error)
	Name() string
}

// Detector 检测器接口
type Detector interface {
	Detect(traffic *ProcessedTraffic) []*Vulnerability
	Name() string
}

// ProcessedTraffic 预处理后的流量
type ProcessedTraffic struct {
	HTTPTraffic
	// 预处理后的附加信息
	ParsedParams   map[string]string // 解析后的参数
	ParsedHeaders  map[string]string // 规范化的头
	ParsedBody     string            // 解析后的请求体
	IsAPI          bool              // 是否为 API 请求
	IsStatic       bool              // 是否为静态资源
	ContentPattern string            // 内容模式
}

// ScanTask 扫描任务
type ScanTask struct {
	ID        string
	Traffic   *HTTPTraffic
	Timestamp time.Time
}

// NewPipelineScanner 创建管道扫描器
func NewPipelineScanner(config *ScannerConfig) *PipelineScanner {
	if config == nil {
		config = DefaultScannerConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &PipelineScanner{
		BaseScanner: *NewBaseScanner("pipeline-scanner", config),
		taskQueue:   make(chan *ScanTask, config.QueueSize),
		workerPool:  make(chan struct{}, config.MaxConcurrent),
		resultChan:  make(chan *ScanResult, 100),
		vulnChan:    make(chan *Vulnerability, 100),
		preprocessors: make([]Preprocessor, 0),
		detectors:     make([]Detector, 0),
		aggregator:    NewResultAggregator(),
		ctx:           ctx,
		cancel:        cancel,
	}
}

// AddPreprocessor 添加预处理器
func (s *PipelineScanner) AddPreprocessor(p Preprocessor) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.preprocessors = append(s.preprocessors, p)
}

// AddDetector 添加检测器
func (s *PipelineScanner) AddDetector(d Detector) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.detectors = append(s.detectors, d)
}

// SetVulnChannel 设置漏洞输出通道
func (s *PipelineScanner) SetVulnChannel(ch chan *Vulnerability) {
	s.vulnChan = ch
}

// Start 启动扫描器
func (s *PipelineScanner) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return nil
	}
	s.running = true
	s.mu.Unlock()

	// 启动工作协程
	for i := 0; i < s.config.MaxConcurrent; i++ {
		s.workerPool <- struct{}{}
	}

	s.wg.Add(1)
	go s.processLoop()

	log.Printf("[Scanner] Pipeline scanner started with %d workers", s.config.MaxConcurrent)
	return nil
}

// Stop 停止扫描器
func (s *PipelineScanner) Stop(ctx context.Context) error {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return nil
	}
	s.running = false
	s.mu.Unlock()

	s.cancel()
	s.wg.Wait()       // 等待 processLoop 完成
	s.workerWg.Wait() // 等待所有工作协程完成

	close(s.taskQueue)
	close(s.resultChan)
	close(s.vulnChan)

	log.Printf("[Scanner] Pipeline scanner stopped")
	return nil
}

// Scan 提交扫描任务
func (s *PipelineScanner) Scan(ctx context.Context, traffic *HTTPTraffic) ([]*Vulnerability, error) {
	task := &ScanTask{
		ID:        traffic.ID,
		Traffic:   traffic,
		Timestamp: time.Now(),
	}

	select {
	case s.taskQueue <- task:
		return nil, nil // 异步扫描，结果通过通道返回
	default:
		// 队列满，丢弃任务
		log.Printf("[Scanner] Task queue full, dropping task: %s", traffic.ID)
		return nil, ErrQueueFull
	}
}

// processLoop 处理循环
func (s *PipelineScanner) processLoop() {
	defer s.wg.Done()

	for {
		select {
		case <-s.ctx.Done():
			return
		case task, ok := <-s.taskQueue:
			if !ok {
				return
			}
			<-s.workerPool // 获取工作槽
			s.workerWg.Add(1)
			go s.processTask(task)
		}
	}
}

// processTask 处理单个扫描任务
func (s *PipelineScanner) processTask(task *ScanTask) {
	defer func() {
		s.workerPool <- struct{}{} // 释放工作槽
		s.workerWg.Done()
	}()

	startTime := time.Now()

	// 1. 预处理
	processed := s.preprocess(task.Traffic)
	if processed == nil {
		return
	}

	// 2. 规则匹配
	matchedRules := s.matchRules(processed)

	// 3. 漏洞检测
	var vulnerabilities []*Vulnerability
	for _, match := range matchedRules {
		vulns := s.detect(processed, match)
		vulnerabilities = append(vulnerabilities, vulns...)
	}

	// 4. 聚合结果
	result := &ScanResult{
		TrafficID:      task.Traffic.ID,
		Vulnerabilities: vulnerabilities,
		Duration:       time.Since(startTime),
		Timestamp:      time.Now(),
	}

	// 5. 更新统计
	s.updateStats(vulnerabilities)

	// 6. 发送结果
	select {
	case s.resultChan <- result:
	default:
	}

	// 7. 发送漏洞
	for _, vuln := range vulnerabilities {
		select {
		case s.vulnChan <- vuln:
		default:
		}
	}
}

// preprocess 执行预处理
func (s *PipelineScanner) preprocess(traffic *HTTPTraffic) *ProcessedTraffic {
	processed := &ProcessedTraffic{
		HTTPTraffic:   *traffic,
		ParsedParams:  make(map[string]string),
		ParsedHeaders: make(map[string]string),
	}

	// 执行预处理器链
	s.mu.RLock()
	preprocessors := s.preprocessors
	s.mu.RUnlock()

	for _, p := range preprocessors {
		result, err := p.Process(traffic)
		if err != nil {
			log.Printf("[Scanner] Preprocessor %s error: %v", p.Name(), err)
			continue
		}
		if result != nil {
			processed = result
		}
	}

	return processed
}

// matchRules 匹配规则
func (s *PipelineScanner) matchRules(traffic *ProcessedTraffic) []*MatchResult {
	s.mu.RLock()
	rules := s.rules
	s.mu.RUnlock()

	var matches []*MatchResult
	for _, rule := range rules {
		if !rule.Enabled() {
			continue
		}
		if matched, result := rule.Match(&traffic.HTTPTraffic); matched {
			matches = append(matches, result)
		}
	}

	// 按优先级排序
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Confidence > matches[j].Confidence
	})

	return matches
}

// detect 执行检测
func (s *PipelineScanner) detect(traffic *ProcessedTraffic, match *MatchResult) []*Vulnerability {
	s.mu.RLock()
	detectors := s.detectors
	s.mu.RUnlock()

	var vulnerabilities []*Vulnerability
	for _, d := range detectors {
		vulns := d.Detect(traffic)
		for _, vuln := range vulns {
			// 补充置信度
			if vuln.Confidence == 0 {
				vuln.Confidence = match.Confidence
			}
			// 补充证据
			if vuln.Evidence == "" {
				vuln.Evidence = match.Evidence
			}
			vulnerabilities = append(vulnerabilities, vuln)
		}
	}

	return vulnerabilities
}

// updateStats 更新统计
func (s *PipelineScanner) updateStats(vulns []*Vulnerability) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.stats.TotalScanned++
	s.stats.TotalVulnsFound += int64(len(vulns))

	for _, v := range vulns {
		s.stats.BySeverity[v.Severity]++
	}
}

// GetResultChannel 获取结果通道
func (s *PipelineScanner) GetResultChannel() <-chan *ScanResult {
	return s.resultChan
}

// GetVulnChannel 获取漏洞通道
func (s *PipelineScanner) GetVulnChannel() <-chan *Vulnerability {
	return s.vulnChan
}

// ResultAggregator 结果聚合器
type ResultAggregator struct {
	results map[string]*ScanResult // traffic_id -> result
	mu      sync.RWMutex
}

// NewResultAggregator 创建结果聚合器
func NewResultAggregator() *ResultAggregator {
	return &ResultAggregator{
		results: make(map[string]*ScanResult),
	}
}

// Add 添加结果
func (a *ResultAggregator) Add(result *ScanResult) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.results[result.TrafficID] = result
}

// Get 获取结果
func (a *ResultAggregator) Get(trafficID string) (*ScanResult, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	result, ok := a.results[trafficID]
	return result, ok
}

// GetAll 获取所有结果
func (a *ResultAggregator) GetAll() []*ScanResult {
	a.mu.RLock()
	defer a.mu.RUnlock()
	results := make([]*ScanResult, 0, len(a.results))
	for _, r := range a.results {
		results = append(results, r)
	}
	return results
}

// 错误定义
var ErrQueueFull = &ScannerError{Msg: "task queue full"}

// ScannerError 扫描器错误
type ScannerError struct {
	Msg string
}

func (e *ScannerError) Error() string {
	return e.Msg
}

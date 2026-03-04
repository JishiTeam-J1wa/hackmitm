// Package intruder implements automated HTTP attack execution
package intruder

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// AttackType defines the type of attack
type AttackType string

const (
	AttackTypeSniper      AttackType = "sniper"
	AttackTypeBatteringRam AttackType = "battering_ram"
	AttackTypePitchfork   AttackType = "pitchfork"
	AttackTypeClusterBomb AttackType = "cluster_bomb"
)

// AttackStatus defines the status of an attack
type AttackStatus string

const (
	StatusPending  AttackStatus = "pending"
	StatusRunning  AttackStatus = "running"
	StatusPaused   AttackStatus = "paused"
	StatusComplete AttackStatus = "complete"
	StatusError    AttackStatus = "error"
)

// PayloadPosition represents a marked position in the request for payload injection
type PayloadPosition struct {
	Start int `json:"start"`
	End   int `json:"end"`
}

// AttackConfig defines the attack configuration
type AttackConfig struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	AttackType      AttackType        `json:"attackType"`
	BaseRequest     string            `json:"baseRequest"`     // Raw HTTP request with § markers
	Method          string            `json:"method"`
	URL             string            `json:"url"`
	Headers         map[string]string `json:"headers"`
	Body            string            `json:"body"`
	PayloadSets     [][]string        `json:"payloadSets"`     // Multiple payload sets for different positions
	Positions       []PayloadPosition `json:"positions"`       // Detected positions
	Concurrency     int               `json:"concurrency"`     // Number of concurrent requests
	RateLimit       int               `json:"rateLimit"`       // Requests per second (0 = unlimited)
	FollowRedirects bool              `json:"followRedirects"`
	Timeout         int               `json:"timeout"`         // Request timeout in seconds
}

// AttackResult represents the result of a single attack request
type AttackResult struct {
	ID           int           `json:"id"`
	Payload      []string      `json:"payload"`      // Payloads used for this request
	StatusCode   int           `json:"statusCode"`
	StatusText   string        `json:"statusText"`
	ResponseTime int64         `json:"responseTime"` // in milliseconds
	Length       int64         `json:"length"`       // response body length
	Error        string        `json:"error,omitempty"`
	Request      string        `json:"request"`
	Response     string        `json:"response"`
	Timestamp    time.Time     `json:"timestamp"`
}

// AttackProgress tracks the progress of an attack
type AttackProgress struct {
	Total       int32 `json:"total"`
	Completed   int32 `json:"completed"`
	Errors      int32 `json:"errors"`
	Status      AttackStatus `json:"status"`
	CurrentRPS  float64 `json:"currentRps"`  // Current requests per second
}

// Attack represents an active attack
type Attack struct {
	Config   *AttackConfig `json:"config"`
	Progress *AttackProgress `json:"progress"`
	Results  []*AttackResult `json:"results"`

	ctx        context.Context
	cancel     context.CancelFunc
	pauseChan  chan struct{}
	resumeChan chan struct{}
	client     *http.Client
	mu         sync.RWMutex

	// Callbacks
	OnProgress  func(*AttackProgress)
	OnResult    func(*AttackResult)
	OnComplete  func()
}

// Intruder manages attack execution
type Intruder struct {
	attacks map[string]*Attack
	mu      sync.RWMutex
}

// NewIntruder creates a new Intruder instance
func NewIntruder() *Intruder {
	return &Intruder{
		attacks: make(map[string]*Attack),
	}
}

// NewAttack creates a new attack with the given configuration
func NewAttack(config *AttackConfig) *Attack {
	if config.Concurrency <= 0 {
		config.Concurrency = 10
	}
	if config.Timeout <= 0 {
		config.Timeout = 30
	}

	// Detect positions if not provided
	if len(config.Positions) == 0 {
		config.Positions = DetectPositions(config.BaseRequest)
	}

	return &Attack{
		Config: config,
		Progress: &AttackProgress{
			Status: StatusPending,
		},
		Results:    make([]*AttackResult, 0),
		pauseChan:  make(chan struct{}),
		resumeChan: make(chan struct{}),
		client: &http.Client{
			Timeout: time.Duration(config.Timeout) * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if !config.FollowRedirects {
					return http.ErrUseLastResponse
				}
				return nil
			},
		},
	}
}

// PositionMarker is the character used to mark payload positions
const PositionMarker = "§"

// DetectPositions finds all §...§ markers in the request
func DetectPositions(request string) []PayloadPosition {
	positions := []PayloadPosition{}
	re := regexp.MustCompile(`§([^§]*)§`)

	matches := re.FindAllStringIndex(request, -1)
	for _, match := range matches {
		positions = append(positions, PayloadPosition{
			Start: match[0],
			End:   match[1],
		})
	}

	return positions
}

// Start begins the attack execution
func (a *Attack) Start(ctx context.Context) error {
	a.mu.Lock()
	if a.Progress.Status == StatusRunning {
		a.mu.Unlock()
		return fmt.Errorf("attack already running")
	}

	a.ctx, a.cancel = context.WithCancel(ctx)
	a.Progress.Status = StatusRunning
	a.mu.Unlock()

	// Generate payload combinations based on attack type
	combinations := a.generatePayloadCombinations()
	atomic.StoreInt32(&a.Progress.Total, int32(len(combinations)))

	// Create worker pool
	var wg sync.WaitGroup
	requestChan := make(chan int, a.Config.Concurrency) // Index into combinations

	// Rate limiter
	var lastRequest time.Time
	rateLimitInterval := time.Duration(0)
	if a.Config.RateLimit > 0 {
		rateLimitInterval = time.Second / time.Duration(a.Config.RateLimit)
	}

	// Start workers
	for i := 0; i < a.Config.Concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for idx := range requestChan {
				select {
				case <-a.ctx.Done():
					return
				case <-a.pauseChan:
					// Wait for resume
					<-a.resumeChan
				default:
				}

				// Rate limiting
				if rateLimitInterval > 0 {
					elapsed := time.Since(lastRequest)
					if elapsed < rateLimitInterval {
						time.Sleep(rateLimitInterval - elapsed)
					}
					lastRequest = time.Now()
				}

				result := a.executeRequest(combinations[idx])
				a.addResult(result)

				atomic.AddInt32(&a.Progress.Completed, 1)
				if a.OnProgress != nil {
					a.OnProgress(a.Progress)
				}
			}
		}()
	}

	// Start time for RPS calculation
	startTime := time.Now()

	// Send requests to workers
	go func() {
		for i := range combinations {
			select {
			case <-a.ctx.Done():
				break
			default:
				requestChan <- i
			}
		}
		close(requestChan)
	}()

	// Wait for completion in background
	go func() {
		wg.Wait()

		a.mu.Lock()
		a.Progress.Status = StatusComplete
		elapsed := time.Since(startTime).Seconds()
		if elapsed > 0 {
			a.Progress.CurrentRPS = float64(a.Progress.Completed) / elapsed
		}
		a.mu.Unlock()

		if a.OnComplete != nil {
			a.OnComplete()
		}
	}()

	return nil
}

// generatePayloadCombinations creates payload combinations based on attack type
func (a *Attack) generatePayloadCombinations() [][]string {
	switch a.Config.AttackType {
	case AttackTypeSniper:
		return a.generateSniperCombinations()
	case AttackTypeBatteringRam:
		return a.generateBatteringRamCombinations()
	case AttackTypePitchfork:
		return a.generatePitchforkCombinations()
	case AttackTypeClusterBomb:
		return a.generateClusterBombCombinations()
	default:
		return a.generateSniperCombinations()
	}
}

// generateSniperCombinations: Tests each payload at each position individually
// Position 1 with Payload1, Position 1 with Payload2, ..., Position 2 with Payload1, ...
func (a *Attack) generateSniperCombinations() [][]string {
	var combinations [][]string
	payloads := a.Config.PayloadSets[0]

	for _, pos := range a.Config.Positions {
		for _, payload := range payloads {
			// Create combination with empty strings, fill in current position
			comb := make([]string, len(a.Config.Positions))
			comb[posIndex(pos, a.Config.Positions)] = payload
			combinations = append(combinations, comb)
		}
	}

	return combinations
}

// generateBatteringRamCombinations: Same payload at all positions simultaneously
// All positions with Payload1, All positions with Payload2, ...
func (a *Attack) generateBatteringRamCombinations() [][]string {
	var combinations [][]string
	payloads := a.Config.PayloadSets[0]

	for _, payload := range payloads {
		comb := make([]string, len(a.Config.Positions))
		for i := range comb {
			comb[i] = payload
		}
		combinations = append(combinations, comb)
	}

	return combinations
}

// generatePitchforkCombinations: Parallel iteration through payload sets
// Position1 with PayloadSet1[0], Position2 with PayloadSet2[0], ...
func (a *Attack) generatePitchforkCombinations() [][]string {
	if len(a.Config.PayloadSets) != len(a.Config.Positions) {
		return nil
	}

	// Find minimum length
	minLen := len(a.Config.PayloadSets[0])
	for _, ps := range a.Config.PayloadSets {
		if len(ps) < minLen {
			minLen = len(ps)
		}
	}

	var combinations [][]string
	for i := 0; i < minLen; i++ {
		comb := make([]string, len(a.Config.Positions))
		for j := range a.Config.Positions {
			comb[j] = a.Config.PayloadSets[j][i]
		}
		combinations = append(combinations, comb)
	}

	return combinations
}

// generateClusterBombCombinations: All combinations of all payload sets
// Cartesian product: Position1 × Position2 × ... × PositionN
func (a *Attack) generateClusterBombCombinations() [][]string {
	if len(a.Config.PayloadSets) != len(a.Config.Positions) {
		return nil
	}

	return cartesianProduct(a.Config.PayloadSets)
}

// cartesianProduct generates all combinations of payload sets
func cartesianProduct(payloadSets [][]string) [][]string {
	if len(payloadSets) == 0 {
		return nil
	}

	if len(payloadSets) == 1 {
		var result [][]string
		for _, p := range payloadSets[0] {
			result = append(result, []string{p})
		}
		return result
	}

	rest := cartesianProduct(payloadSets[1:])
	var result [][]string

	for _, p := range payloadSets[0] {
		for _, r := range rest {
			comb := append([]string{p}, r...)
			result = append(result, comb)
		}
	}

	return result
}

// posIndex finds the index of a position in the positions slice
func posIndex(pos PayloadPosition, positions []PayloadPosition) int {
	for i, p := range positions {
		if p.Start == pos.Start && p.End == pos.End {
			return i
		}
	}
	return 0
}

// executeRequest executes a single attack request
func (a *Attack) executeRequest(payloads []string) *AttackResult {
	// Build request with payloads injected
	requestStr := a.injectPayloads(payloads)
	startTime := time.Now()

	// Parse the request
	req, err := http.NewRequest(a.Config.Method, a.Config.URL, bytes.NewBufferString(requestStr))
	if err != nil {
		return &AttackResult{
			Payload:   payloads,
			Error:     err.Error(),
			Timestamp: time.Now(),
		}
	}

	// Add headers
	for key, value := range a.Config.Headers {
		req.Header.Set(key, value)
	}

	// Execute request
	resp, err := a.client.Do(req)
	if err != nil {
		return &AttackResult{
			Payload:   payloads,
			Error:     err.Error(),
			Timestamp: time.Now(),
		}
	}
	defer resp.Body.Close()

	// Read response
	body, _ := io.ReadAll(resp.Body)
	elapsed := time.Since(startTime)

	result := &AttackResult{
		Payload:      payloads,
		StatusCode:   resp.StatusCode,
		StatusText:   resp.Status,
		ResponseTime: elapsed.Milliseconds(),
		Length:       int64(len(body)),
		Request:      requestStr,
		Response:     string(body),
		Timestamp:    time.Now(),
	}

	return result
}

// injectPayloads replaces § markers with payloads
func (a *Attack) injectPayloads(payloads []string) string {
	if len(a.Config.Positions) == 0 || len(payloads) == 0 {
		return a.Config.Body
	}

	result := a.Config.BaseRequest
	offset := 0

	for i, pos := range a.Config.Positions {
		if i >= len(payloads) {
			break
		}
		payload := payloads[i]

		// Calculate adjusted positions
		start := pos.Start + offset
		end := pos.End + offset

		// Replace the marker with payload
		before := result[:start]
		after := result[end:]
		result = before + payload + after

		// Adjust offset for length difference
		offset += len(payload) - (pos.End - pos.Start)
	}

	return result
}

// addResult adds a result to the attack
func (a *Attack) addResult(result *AttackResult) {
	a.mu.Lock()
	defer a.mu.Unlock()

	result.ID = len(a.Results)
	a.Results = append(a.Results, result)

	if result.Error != "" {
		atomic.AddInt32(&a.Progress.Errors, 1)
	}

	if a.OnResult != nil {
		go a.OnResult(result)
	}
}

// Pause pauses the attack
func (a *Attack) Pause() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.Progress.Status == StatusRunning {
		a.Progress.Status = StatusPaused
		a.pauseChan <- struct{}{}
	}
}

// Resume resumes a paused attack
func (a *Attack) Resume() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.Progress.Status == StatusPaused {
		a.Progress.Status = StatusRunning
		a.resumeChan <- struct{}{}
	}
}

// Stop stops the attack
func (a *Attack) Stop() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.cancel != nil {
		a.cancel()
	}
	a.Progress.Status = StatusComplete
}

// GetProgress returns current progress
func (a *Attack) GetProgress() *AttackProgress {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.Progress
}

// GetResults returns all results
func (a *Attack) GetResults() []*AttackResult {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.Results
}

// Intruder management methods

// CreateAttack creates and stores a new attack
func (i *Intruder) CreateAttack(config *AttackConfig) *Attack {
	attack := NewAttack(config)
	i.mu.Lock()
	i.attacks[config.ID] = attack
	i.mu.Unlock()
	return attack
}

// GetAttack retrieves an attack by ID
func (i *Intruder) GetAttack(id string) *Attack {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.attacks[id]
}

// RemoveAttack removes an attack
func (i *Intruder) RemoveAttack(id string) {
	i.mu.Lock()
	defer i.mu.Unlock()
	delete(i.attacks, id)
}

// ListAttacks returns all attack IDs
func (i *Intruder) ListAttacks() []string {
	i.mu.RLock()
	defer i.mu.RUnlock()

	ids := make([]string, 0, len(i.attacks))
	for id := range i.attacks {
		ids = append(ids, id)
	}
	return ids
}

// Helper: Parse raw HTTP request string
func ParseRawRequest(raw string) (method, url string, headers map[string]string, body string) {
	headers = make(map[string]string)
	parts := strings.SplitN(raw, "\r\n\r\n", 2)

	if len(parts) > 1 {
		body = parts[1]
	}

	lines := strings.Split(parts[0], "\r\n")
	if len(lines) > 0 {
		// Parse request line
		requestLine := strings.Fields(lines[0])
		if len(requestLine) >= 2 {
			method = requestLine[0]
			url = requestLine[1]
		}

		// Parse headers
		for _, line := range lines[1:] {
			if line == "" {
				break
			}
			idx := strings.Index(line, ":")
			if idx > 0 {
				key := strings.TrimSpace(line[:idx])
				value := strings.TrimSpace(line[idx+1:])
				headers[key] = value
			}
		}
	}

	return
}

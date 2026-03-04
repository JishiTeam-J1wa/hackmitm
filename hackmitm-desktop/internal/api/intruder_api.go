package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"hackmitm-desktop/internal/intruder"
)

// IntruderAPI handles intruder attack operations
type IntruderAPI struct {
	ctx     context.Context
	db      *sql.DB
	intruder *intruder.Intruder
	mu      sync.RWMutex
}

// NewIntruderAPI creates a new IntruderAPI instance
func NewIntruderAPI() *IntruderAPI {
	return &IntruderAPI{
		intruder: intruder.NewIntruder(),
	}
}

// SetContext sets the application context
func (a *IntruderAPI) SetContext(ctx context.Context) {
	a.ctx = ctx
}

// SetDB sets the database connection
func (a *IntruderAPI) SetDB(db *sql.DB) {
	a.db = db
}

// ============ Attack Management ============

// CreateAttack creates a new attack with the given configuration
func (a *IntruderAPI) CreateAttack(config intruder.AttackConfig) (*intruder.Attack, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	attack := a.intruder.CreateAttack(&config)
	return attack, nil
}

// StartAttack starts an attack by ID
func (a *IntruderAPI) StartAttack(attackID string) error {
	a.mu.RLock()
	attack := a.intruder.GetAttack(attackID)
	a.mu.RUnlock()

	if attack == nil {
		return fmt.Errorf("attack not found: %s", attackID)
	}

	// Set up callbacks
	attack.OnProgress = func(progress *intruder.AttackProgress) {
		// Could emit event via context if available
	}

	attack.OnResult = func(result *intruder.AttackResult) {
		// Store result in database
		if a.db != nil {
			go a.storeAttackResult(attackID, result)
		}
	}

	return attack.Start(a.ctx)
}

// PauseAttack pauses a running attack
func (a *IntruderAPI) PauseAttack(attackID string) error {
	a.mu.RLock()
	attack := a.intruder.GetAttack(attackID)
	a.mu.RUnlock()

	if attack == nil {
		return fmt.Errorf("attack not found: %s", attackID)
	}

	attack.Pause()
	return nil
}

// ResumeAttack resumes a paused attack
func (a *IntruderAPI) ResumeAttack(attackID string) error {
	a.mu.RLock()
	attack := a.intruder.GetAttack(attackID)
	a.mu.RUnlock()

	if attack == nil {
		return fmt.Errorf("attack not found: %s", attackID)
	}

	attack.Resume()
	return nil
}

// StopAttack stops an attack
func (a *IntruderAPI) StopAttack(attackID string) error {
	a.mu.RLock()
	attack := a.intruder.GetAttack(attackID)
	a.mu.RUnlock()

	if attack == nil {
		return fmt.Errorf("attack not found: %s", attackID)
	}

	attack.Stop()
	return nil
}

// GetAttackProgress returns the progress of an attack
func (a *IntruderAPI) GetAttackProgress(attackID string) (*intruder.AttackProgress, error) {
	a.mu.RLock()
	attack := a.intruder.GetAttack(attackID)
	a.mu.RUnlock()

	if attack == nil {
		return nil, fmt.Errorf("attack not found: %s", attackID)
	}

	return attack.GetProgress(), nil
}

// GetAttackResults returns all results from an attack
func (a *IntruderAPI) GetAttackResults(attackID string) ([]*intruder.AttackResult, error) {
	a.mu.RLock()
	attack := a.intruder.GetAttack(attackID)
	a.mu.RUnlock()

	if attack == nil {
		return nil, fmt.Errorf("attack not found: %s", attackID)
	}

	return attack.GetResults(), nil
}

// GetAttackConfig returns the configuration of an attack
func (a *IntruderAPI) GetAttackConfig(attackID string) (*intruder.AttackConfig, error) {
	a.mu.RLock()
	attack := a.intruder.GetAttack(attackID)
	a.mu.RUnlock()

	if attack == nil {
		return nil, fmt.Errorf("attack not found: %s", attackID)
	}

	return attack.Config, nil
}

// ListAttacks returns all attack IDs
func (a *IntruderAPI) ListAttacks() []string {
	return a.intruder.ListAttacks()
}

// RemoveAttack removes an attack
func (a *IntruderAPI) RemoveAttack(attackID string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.intruder.RemoveAttack(attackID)
	return nil
}

// ============ Position Detection ============

// DetectPositions finds all payload positions in a request
func (a *IntruderAPI) DetectPositions(request string) []intruder.PayloadPosition {
	return intruder.DetectPositions(request)
}

// ============ Database Storage ============

// storeAttackResult stores an attack result in the database
func (a *IntruderAPI) storeAttackResult(attackID string, result *intruder.AttackResult) error {
	if a.db == nil {
		return nil
	}

	payloadJSON, _ := json.Marshal(result.Payload)

	_, err := a.db.Exec(`
		INSERT INTO intruder_results (attack_id, payload, status_code, status_text, response_time, length, error, request, response, timestamp)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		attackID, string(payloadJSON), result.StatusCode, result.StatusText,
		result.ResponseTime, result.Length, result.Error, result.Request, result.Response,
		result.Timestamp.Format(time.RFC3339),
	)

	return err
}

// GetStoredAttackResults retrieves attack results from database
func (a *IntruderAPI) GetStoredAttackResults(attackID string, limit int) ([]AttackResultRow, error) {
	if a.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	query := `SELECT id, attack_id, payload, status_code, status_text, response_time, length, error, request, response, timestamp
			  FROM intruder_results WHERE attack_id = ? ORDER BY timestamp DESC LIMIT ?`

	rows, err := a.db.Query(query, attackID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []AttackResultRow
	for rows.Next() {
		var r AttackResultRow
		var payloadStr, timestamp sql.NullString
		var errStr, request, response sql.NullString

		err := rows.Scan(
			&r.ID, &r.AttackID, &payloadStr, &r.StatusCode, &r.StatusText,
			&r.ResponseTime, &r.Length, &errStr, &request, &response, &timestamp,
		)
		if err != nil {
			continue
		}

		if payloadStr.Valid {
			json.Unmarshal([]byte(payloadStr.String), &r.Payload)
		}
		r.Error = errStr.String
		r.Request = request.String
		r.Response = response.String
		r.Timestamp = timestamp.String

		results = append(results, r)
	}

	return results, nil
}

// AttackResultRow represents a stored attack result
type AttackResultRow struct {
	ID           int64    `json:"id"`
	AttackID     string   `json:"attackId"`
	Payload      []string `json:"payload"`
	StatusCode   int      `json:"statusCode"`
	StatusText   string   `json:"statusText"`
	ResponseTime int64    `json:"responseTime"`
	Length       int64    `json:"length"`
	Error        string   `json:"error,omitempty"`
	Request      string   `json:"request"`
	Response     string   `json:"response"`
	Timestamp    string   `json:"timestamp"`
}

// ============ Attack Types ============

// GetAttackTypes returns available attack types
func (a *IntruderAPI) GetAttackTypes() []AttackTypeInfo {
	return []AttackTypeInfo{
		{
			ID:          string(intruder.AttackTypeSniper),
			Name:        "Sniper",
			Description: "Tests each payload at each position individually. Best for single-position testing.",
		},
		{
			ID:          string(intruder.AttackTypeBatteringRam),
			Name:        "Battering Ram",
			Description: "Uses the same payload at all positions simultaneously. Useful for credential stuffing.",
		},
		{
			ID:          string(intruder.AttackTypePitchfork),
			Name:        "Pitchfork",
			Description: "Iterates through multiple payload sets in parallel. Each position gets its own payload list.",
		},
		{
			ID:          string(intruder.AttackTypeClusterBomb),
			Name:        "Cluster Bomb",
			Description: "Tests all combinations of payload sets. Maximum coverage but can generate many requests.",
		},
	}
}

// AttackTypeInfo provides information about an attack type
type AttackTypeInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// ============ Utility Functions ============

// ParseRawRequest parses a raw HTTP request string
func (a *IntruderAPI) ParseRawRequest(raw string) (method, url string, headers map[string]string, body string) {
	return intruder.ParseRawRequest(raw)
}

// EstimateRequestCount estimates the total number of requests for an attack
func (a *IntruderAPI) EstimateRequestCount(attackType intruder.AttackType, positionCount int, payloadSetSizes []int) int {
	switch attackType {
	case intruder.AttackTypeSniper:
		if len(payloadSetSizes) == 0 {
			return 0
		}
		return positionCount * payloadSetSizes[0]
	case intruder.AttackTypeBatteringRam:
		if len(payloadSetSizes) == 0 {
			return 0
		}
		return payloadSetSizes[0]
	case intruder.AttackTypePitchfork:
		if len(payloadSetSizes) == 0 {
			return 0
		}
		minSize := payloadSetSizes[0]
		for _, size := range payloadSetSizes {
			if size < minSize {
				minSize = size
			}
		}
		return minSize
	case intruder.AttackTypeClusterBomb:
		if len(payloadSetSizes) == 0 {
			return 0
		}
		total := 1
		for _, size := range payloadSetSizes {
			total *= size
		}
		return total
	default:
		return 0
	}
}

// CreateAttackConfig is a helper to create an attack configuration
func (a *IntruderAPI) CreateAttackConfig(
	id, name string,
	attackType intruder.AttackType,
	baseRequest, method, url string,
	headers map[string]string,
	body string,
	payloadSets [][]string,
	concurrency, rateLimit, timeout int,
	followRedirects bool,
) *intruder.AttackConfig {
	return &intruder.AttackConfig{
		ID:              id,
		Name:            name,
		AttackType:      attackType,
		BaseRequest:     baseRequest,
		Method:          method,
		URL:             url,
		Headers:         headers,
		Body:            body,
		PayloadSets:     payloadSets,
		Concurrency:     concurrency,
		RateLimit:       rateLimit,
		Timeout:         timeout,
		FollowRedirects: followRedirects,
	}
}

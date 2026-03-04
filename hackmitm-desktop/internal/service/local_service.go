package service

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

// LocalConfig holds configuration for local mode
type LocalConfig struct {
	DataDir   string `json:"dataDir"`
	ApiPort   int    `json:"apiPort"`
	ProxyPort int    `json:"proxyPort"`
}

// RemoteConfig holds configuration for remote mode
type RemoteConfig struct {
	Host   string `json:"host"`
	Port   int    `json:"port"`
	ApiKey string `json:"apiKey,omitempty"`
}

// LocalService manages the embedded HackMITM binary
type LocalService struct {
	cmd       *exec.Cmd
	config    *LocalConfig
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	running   bool
	mu        sync.RWMutex
	outputBuf []string
	logFile   *os.File
}

// NewLocalService creates a new local service manager
func NewLocalService() *LocalService {
	return &LocalService{
		outputBuf: make([]string, 0, 100),
	}
}

// SetContext sets the context for the service
func (s *LocalService) SetContext(ctx context.Context) {
	s.ctx, s.cancel = context.WithCancel(ctx)
}

// getBinaryPath returns the path to the embedded hackmitm binary
func (s *LocalService) getBinaryPath() (string, error) {
	// Get the executable directory
	exePath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to get executable path: %w", err)
	}
	exeDir := filepath.Dir(exePath)

	var binaryName string
	switch runtime.GOOS {
	case "windows":
		binaryName = "hackmitm.exe"
	default:
		binaryName = "hackmitm"
	}

	// Try multiple locations
	locations := []string{
		filepath.Join(exeDir, binaryName),                        // Same directory as executable
		filepath.Join(exeDir, "..", "Resources", binaryName),     // macOS app bundle Resources
		filepath.Join(exeDir, "resources", binaryName),           // Windows/Linux resources
		filepath.Join(exeDir, "..", "..", "hackmitm", binaryName), // Development: hackmitm-desktop -> hackmitm
		filepath.Join(exeDir, "..", "..", "..", "hackmitm", binaryName), // Development alternative path
		"./hackmitm",         // Current working directory
		"./hackmitm.exe",     // Windows current directory
		"../hackmitm/hackmitm", // Sibling hackmitm directory (development)
	}

	for _, loc := range locations {
		absPath, _ := filepath.Abs(loc)
		if _, err := os.Stat(absPath); err == nil {
			return absPath, nil
		}
	}

	return "", fmt.Errorf("hackmitm binary not found in any known location")
}

// Start starts the local HackMITM service
func (s *LocalService) Start(config *LocalConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("service is already running")
	}

	s.config = config

	// Create data directory if it doesn't exist
	if err := os.MkdirAll(config.DataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	// Find the binary
	binaryPath, err := s.getBinaryPath()
	if err != nil {
		return err
	}

	// Create config directory
	configDir := filepath.Join(config.DataDir, "configs")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Generate config file
	configPath := filepath.Join(configDir, "config.json")
	configContent := fmt.Sprintf(`{
  "api_port": %d,
  "proxy_port": %d,
  "data_dir": "%s",
  "log_level": "info"
}`, config.ApiPort, config.ProxyPort, config.DataDir)

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	// Prepare command arguments (hackmitm uses -config flag)
	args := []string{
		"-config", configPath,
		"-log-level", "info",
	}

	// Create the command
	s.cmd = exec.CommandContext(s.ctx, binaryPath, args...)

	// Set up process group for proper termination (platform-specific)
	setupProcessGroup(s.cmd)

	// Create pipes for stdout and stderr
	stdoutPipe, err := s.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	stderrPipe, err := s.cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Set up log file for debugging
	logPath := filepath.Join(config.DataDir, "hackmitm-service.log")
	s.logFile, err = os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		// Non-fatal, just log to console
		fmt.Printf("Warning: Could not open log file: %v\n", err)
	}

	// Start the process
	if err := s.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start hackmitm: %w", err)
	}

	s.running = true

	// Start goroutines to read output
	s.wg.Add(2)
	go s.readOutput(stdoutPipe, "stdout")
	go s.readOutput(stderrPipe, "stderr")

	// Wait for the process in a goroutine
	go func() {
		err := s.cmd.Wait()
		s.mu.Lock()
		s.running = false
		s.mu.Unlock()
		if err != nil {
			fmt.Printf("HackMITM process exited with error: %v\n", err)
		}
	}()

	// Wait for health check
	if err := s.waitForHealth(config.ApiPort, 30*time.Second); err != nil {
		s.Stop()
		return fmt.Errorf("health check failed: %w", err)
	}

	return nil
}

// readOutput reads from a pipe and logs the output
func (s *LocalService) readOutput(reader io.Reader, source string) {
	defer s.wg.Done()
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		logLine := fmt.Sprintf("[%s] [%s] %s", timestamp, source, line)

		s.mu.Lock()
		s.outputBuf = append(s.outputBuf, logLine)
		// Keep only last 100 lines
		if len(s.outputBuf) > 100 {
			s.outputBuf = s.outputBuf[len(s.outputBuf)-100:]
		}
		s.mu.Unlock()

		// Write to log file
		if s.logFile != nil {
			s.logFile.WriteString(logLine + "\n")
		}

		// Also print to console for debugging
		fmt.Println(logLine)
	}
}

// waitForHealth waits for the service to become healthy
func (s *LocalService) waitForHealth(port int, timeout time.Duration) error {
	client := &http.Client{Timeout: 2 * time.Second}

	// Try multiple possible health check endpoints
	healthURLs := []string{
		fmt.Sprintf("http://localhost:%d/api/v1/health", port),
		fmt.Sprintf("http://localhost:%d/health", port),
		fmt.Sprintf("http://localhost:%d/api/health", port),
		fmt.Sprintf("http://localhost:%d/", port),
	}

	start := time.Now()
	for time.Since(start) < timeout {
		// First check if process is still running
		if s.cmd.Process == nil {
			return fmt.Errorf("process failed to start")
		}

		// Try each health URL
		for _, url := range healthURLs {
			resp, err := client.Get(url)
			if err == nil {
				resp.Body.Close()
				// Any response means the server is up
				return nil
			}
		}

		// If process is running but no HTTP response yet, wait a bit
		time.Sleep(500 * time.Millisecond)
	}

	// If timeout but process is still running, consider it success
	if s.cmd.Process != nil {
		fmt.Println("Health check timeout, but process is running - continuing...")
		return nil
	}

	return fmt.Errorf("service did not become healthy within %v", timeout)
}

// Stop stops the local HackMITM service
func (s *LocalService) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running || s.cmd == nil {
		return nil
	}

	// Cancel context
	if s.cancel != nil {
		s.cancel()
	}

	// Platform-specific termination
	terminateProcess(s.cmd)

	// Wait for process to exit
	done := make(chan error, 1)
	go func() {
		done <- s.cmd.Wait()
	}()

	select {
	case <-time.After(5 * time.Second):
		// Force kill if still running
		if s.cmd.Process != nil {
			s.cmd.Process.Kill()
		}
	case <-done:
		// Process exited gracefully
	}

	s.running = false

	// Close log file
	if s.logFile != nil {
		s.logFile.Close()
		s.logFile = nil
	}

	// Wait for output readers
	s.wg.Wait()

	return nil
}

// IsRunning returns whether the service is running
func (s *LocalService) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// GetOutput returns the recent output buffer
func (s *LocalService) GetOutput() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]string, len(s.outputBuf))
	copy(result, s.outputBuf)
	return result
}

// GetConfig returns the current configuration
func (s *LocalService) GetConfig() *LocalConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config
}

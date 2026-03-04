// +build !windows

package service

import (
	"os/exec"
	"syscall"
)

// setupProcessGroup sets up the process group for Unix systems
func setupProcessGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
}

// terminateProcess gracefully terminates the process on Unix systems
func terminateProcess(cmd *exec.Cmd) {
	if cmd.Process != nil {
		// Send SIGTERM to process group
		syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM)
	}
}

// +build windows

package service

import (
	"os/exec"
)

// setupProcessGroup sets up the process for Windows systems
func setupProcessGroup(cmd *exec.Cmd) {
	// On Windows, we don't need special setup
	// The process will be killed directly when needed
}

// terminateProcess terminates the process on Windows systems
func terminateProcess(cmd *exec.Cmd) {
	if cmd.Process != nil {
		// On Windows, just kill the process
		cmd.Process.Kill()
	}
}

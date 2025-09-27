package utils

import (
	"bytes"
	"os/exec"
)

func RunCommand(command string, args ...string) (int, string, string, error) {
	var stdout, stderr bytes.Buffer

	// Construct command
	cmd := exec.Command(command, args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Run the command
	err := cmd.Run()

	// Exit code (default to -1 if unknown)
	exitCode := -1
	if cmd.ProcessState != nil {
		exitCode = cmd.ProcessState.ExitCode()
	}

	return exitCode, stdout.String(), stderr.String(), err
}

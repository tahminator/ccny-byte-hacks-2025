package utils

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

func RunCommand(cmdStr string) (int, string, string, error) {
	var stdout, stderr bytes.Buffer

	// Print command execution header
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Printf("🚀 EXECUTING SHELL COMMAND\n")
	fmt.Printf("⏰ Timestamp: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Printf("💻 Command: %s\n", cmdStr)
	fmt.Println(strings.Repeat("=", 80))

	// Construct command
	cmd := exec.Command("bash", "-c", cmdStr)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Run the command
	err := cmd.Run()

	// Exit code (default to -1 if unknown)
	exitCode := -1
	if cmd.ProcessState != nil {
		exitCode = cmd.ProcessState.ExitCode()
	}

	// Capture output strings
	stdoutStr := stdout.String()
	stderrStr := stderr.String()

	// Print detailed output with clear formatting
	fmt.Println("\n" + strings.Repeat("-", 80))
	fmt.Printf("📊 COMMAND EXECUTION RESULTS\n")
	fmt.Printf("🔢 Exit Code: %d\n", exitCode)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	} else {
		fmt.Printf("✅ Execution completed successfully\n")
	}
	fmt.Println(strings.Repeat("-", 80))

	// Print STDOUT with clear formatting
	if stdoutStr != "" {
		fmt.Println("\n📤 STDOUT OUTPUT:")
		fmt.Println(strings.Repeat("─", 40))
		// Split by lines and add line numbers for better readability
		lines := strings.Split(strings.TrimRight(stdoutStr, "\n"), "\n")
		for i, line := range lines {
			fmt.Printf("%3d | %s\n", i+1, line)
		}
		fmt.Println(strings.Repeat("─", 40))
	} else {
		fmt.Println("\n📤 STDOUT OUTPUT: (empty)")
	}

	// Print STDERR with clear formatting
	if stderrStr != "" {
		fmt.Println("\n📥 STDERR OUTPUT:")
		fmt.Println(strings.Repeat("─", 40))
		// Split by lines and add line numbers for better readability
		lines := strings.Split(strings.TrimRight(stderrStr, "\n"), "\n")
		for i, line := range lines {
			fmt.Printf("%3d | %s\n", i+1, line)
		}
		fmt.Println(strings.Repeat("─", 40))
	} else {
		fmt.Println("\n📥 STDERR OUTPUT: (empty)")
	}

	// Print execution summary
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Printf("🏁 COMMAND EXECUTION COMPLETED\n")
	fmt.Printf("⏰ Completed at: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println(strings.Repeat("=", 80) + "\n")

	return exitCode, stdoutStr, stderrStr, err
}

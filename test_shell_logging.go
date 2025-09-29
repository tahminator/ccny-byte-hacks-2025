package main

import (
	"fmt"

	"github.com/tahminator/go-react-template/utils"
)

func main() {
	fmt.Println("Testing enhanced shell logging...")

	// Test 1: Simple command with stdout
	fmt.Println("\n=== Test 1: Simple command with stdout ===")
	exitCode, stdout, stderr, err := utils.RunCommand("echo 'Hello, World!'")
	fmt.Printf("Return values - Exit Code: %d, Stdout: %q, Stderr: %q, Error: %v\n", exitCode, stdout, stderr, err)

	// Test 2: Command with stderr
	fmt.Println("\n=== Test 2: Command with stderr ===")
	exitCode, stdout, stderr, err = utils.RunCommand("ls /nonexistent/directory 2>&1")
	fmt.Printf("Return values - Exit Code: %d, Stdout: %q, Stderr: %q, Error: %v\n", exitCode, stdout, stderr, err)

	// Test 3: Command with both stdout and stderr
	fmt.Println("\n=== Test 3: Command with both stdout and stderr ===")
	exitCode, stdout, stderr, err = utils.RunCommand("echo 'This goes to stdout' && echo 'This goes to stderr' >&2")
	fmt.Printf("Return values - Exit Code: %d, Stdout: %q, Stderr: %q, Error: %v\n", exitCode, stdout, stderr, err)

	// Test 4: Multi-line output
	fmt.Println("\n=== Test 4: Multi-line output ===")
	exitCode, stdout, stderr, err = utils.RunCommand("echo -e 'Line 1\nLine 2\nLine 3'")
	fmt.Printf("Return values - Exit Code: %d, Stdout: %q, Stderr: %q, Error: %v\n", exitCode, stdout, stderr, err)

	fmt.Println("\n=== All tests completed ===")
}

// Package main_test tests the CLI entry point
//
// Coverage Note:
// This package shows 0% coverage because main.go contains minimal code:
//   - Version variables (injected at build time)
//   - A single call to cli.SetVersion()
//   - A single call to cli.Execute()
//
// Integration tests that build and run the binary as a subprocess
// cannot be measured by Go's coverage tool. The actual business logic
// is in internal/cli and other internal packages, which have their own tests.
//
// These tests verify:
//   - Binary builds successfully
//   - Version flag works correctly
//   - Help output contains expected commands
//   - Invalid commands return non-zero exit codes
//   - All CLI flags are recognized
package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

var binaryPath string

// TestMain builds the binary once for all tests
func TestMain(m *testing.M) {
	// Build the binary once for all tests
	tmpDir, err := os.MkdirTemp("", "todo-tracker-test")
	if err != nil {
		panic("failed to create temp dir: " + err.Error())
	}
	defer os.RemoveAll(tmpDir)

	binaryName := "todo"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}
	binaryPath = filepath.Join(tmpDir, binaryName)

	// Build the binary
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Dir = filepath.Join("..", "..", "cmd", "todo")
	output, err := cmd.CombinedOutput()
	if err != nil {
		panic("failed to build binary: " + err.Error() + "\n" + string(output))
	}

	os.Exit(m.Run())
}

// TestVersionFlag tests that --version flag works correctly
func TestVersionFlag(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	cmd := exec.Command(binaryPath, "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("--version flag failed: %v\nOutput: %s", err, output)
	}

	// Check version output format
	outputStr := strings.TrimSpace(string(output))
	if !strings.Contains(outputStr, "todo") && !strings.Contains(outputStr, "dev") {
		t.Errorf("unexpected version output: %s", outputStr)
	}
}

// TestHelpFlag tests that --help flag works correctly
func TestHelpFlag(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	cmd := exec.Command(binaryPath, "--help")
	output, err := cmd.CombinedOutput()
	// --help returns exit code 0 in cobra
	if err != nil {
		t.Fatalf("--help flag failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)

	// Check for expected content in help output
	expectedPhrases := []string{
		"TODO",
		"scan",
		"stale",
		"orphaned",
		"report",
	}

	for _, phrase := range expectedPhrases {
		if !strings.Contains(outputStr, phrase) {
			t.Errorf("help output missing expected phrase %q", phrase)
		}
	}
}

// TestHelpShortFlag tests that -h flag works correctly
func TestHelpShortFlag(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	cmd := exec.Command(binaryPath, "-h")
	output, err := cmd.CombinedOutput()
	// -h returns exit code 0 in cobra
	if err != nil {
		t.Fatalf("-h flag failed: %v\nOutput: %s", err, output)
	}

	// Should contain same help content as --help
	outputStr := string(output)
	if !strings.Contains(outputStr, "TODO") {
		t.Errorf("help output missing expected content")
	}
}

// TestInvalidCommand tests that invalid commands return non-zero exit code
func TestInvalidCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	cmd := exec.Command(binaryPath, "nonexistent-command-xyz")
	output, _ := cmd.CombinedOutput()

	// Should return non-zero exit code
	if cmd.ProcessState == nil {
		t.Fatal("process state is nil")
	}

	if cmd.ProcessState.Success() {
		t.Errorf("expected non-zero exit code for invalid command, got success.\nOutput: %s", output)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "unknown command") {
		t.Errorf("expected 'unknown command' error message, got: %s", outputStr)
	}
}

// TestNoArgs tests running the CLI with no arguments (should show help)
func TestNoArgs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	cmd := exec.Command(binaryPath)
	output, err := cmd.CombinedOutput()
	// No args typically shows help and exits successfully in cobra
	if err != nil {
		t.Fatalf("no args failed: %v\nOutput: %s", err, output)
	}

	// Should show help content
	outputStr := string(output)
	if !strings.Contains(outputStr, "TODO") {
		t.Errorf("expected help content in output, got: %s", outputStr)
	}
}

// TestVerboseFlag tests that --verbose flag is accepted
func TestVerboseFlag(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// --verbose with help should work
	cmd := exec.Command(binaryPath, "--verbose", "--help")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("--verbose --help failed: %v\nOutput: %s", err, output)
	}

	// Should show help content
	outputStr := string(output)
	if !strings.Contains(outputStr, "TODO") {
		t.Errorf("expected help content in output, got: %s", outputStr)
	}
}

// TestScanHelp tests that scan subcommand has help
func TestScanHelp(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	cmd := exec.Command(binaryPath, "scan", "--help")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("scan --help failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	expectedPhrases := []string{"scan", "TODO"}

	for _, phrase := range expectedPhrases {
		if !strings.Contains(outputStr, phrase) {
			t.Errorf("scan help missing expected phrase %q", phrase)
		}
	}
}

// TestStaleHelp tests that stale subcommand has help
func TestStaleHelp(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	cmd := exec.Command(binaryPath, "stale", "--help")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("stale --help failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "stale") {
		t.Errorf("stale help missing expected phrase 'stale'")
	}
}

// TestOrphanedHelp tests that orphaned subcommand has help
func TestOrphanedHelp(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	cmd := exec.Command(binaryPath, "orphaned", "--help")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("orphaned --help failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "orphaned") {
		t.Errorf("orphaned help missing expected phrase 'orphaned'")
	}
}

// TestReportHelp tests that report subcommand has help
func TestReportHelp(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	cmd := exec.Command(binaryPath, "report", "--help")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("report --help failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "report") {
		t.Errorf("report help missing expected phrase 'report'")
	}
}

// TestConfigFlag tests that --config flag is recognized
func TestConfigFlag(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Using --config with a non-existent file with --help should still show help
	cmd := exec.Command(binaryPath, "--config", "/nonexistent/config.yaml", "--help")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("--config --help failed: %v\nOutput: %s", err, output)
	}

	// Should show help content
	outputStr := string(output)
	if !strings.Contains(outputStr, "TODO") {
		t.Errorf("expected help content in output, got: %s", outputStr)
	}
}

// TestOutputFlag tests that --output flag is recognized
func TestOutputFlag(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Using --output with --help should still show help
	cmd := exec.Command(binaryPath, "--output", "/tmp/test-output.txt", "--help")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("--output --help failed: %v\nOutput: %s", err, output)
	}

	// Should show help content
	outputStr := string(output)
	if !strings.Contains(outputStr, "TODO") {
		t.Errorf("expected help content in output, got: %s", outputStr)
	}
}
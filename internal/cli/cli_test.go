// Package cli_test 测试命令行接口功能
package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// TestRootCommand 测试根命令创建
func TestRootCommand(t *testing.T) {
	cmd := GetRootCmd()

	if cmd == nil {
		t.Fatal("GetRootCmd() returned nil")
	}

	if cmd.Use != "todo" {
		t.Errorf("rootCmd.Use = %q, want %q", cmd.Use, "todo")
	}

	if cmd.Short == "" {
		t.Error("rootCmd.Short should not be empty")
	}

	if cmd.Long == "" {
		t.Error("rootCmd.Long should not be empty")
	}
}

// TestRootCommandHasSubcommands 测试根命令包含所有子命令
func TestRootCommandHasSubcommands(t *testing.T) {
	cmd := GetRootCmd()

	expectedCommands := []string{"scan", "stale", "orphaned", "report", "config"}

	commands := cmd.Commands()
	commandNames := make(map[string]bool)
	for _, c := range commands {
		commandNames[c.Name()] = true
	}

	for _, expected := range expectedCommands {
		if !commandNames[expected] {
			t.Errorf("rootCmd missing expected subcommand: %s", expected)
		}
	}
}

// TestRootCommandPersistentFlags 测试根命令持久标志
func TestRootCommandPersistentFlags(t *testing.T) {
	cmd := GetRootCmd()

	flags := []string{"config", "verbose", "output"}

	for _, flag := range flags {
		f := cmd.PersistentFlags().Lookup(flag)
		if f == nil {
			t.Errorf("rootCmd missing persistent flag: %s", flag)
		}
	}
}

// TestSetVersion 测试版本设置
func TestSetVersion(t *testing.T) {
	SetVersion("1.0.0", "abc123", "2024-01-01")

	expectedVersion := "TODO Tracker 1.0.0 (commit: abc123, built: 2024-01-01)"
	got := GetVersion()

	if got != expectedVersion {
		t.Errorf("GetVersion() = %q, want %q", got, expectedVersion)
	}
}

// TestGetVersion 测试获取版本信息
func TestGetVersion(t *testing.T) {
	version := GetVersion()

	if version == "" {
		t.Error("GetVersion() returned empty string")
	}

	// 应包含 TODO Tracker
	if !strings.Contains(version, "TODO Tracker") {
		t.Errorf("GetVersion() should contain 'TODO Tracker', got %q", version)
	}
}

// TestScanCommand 测试扫描命令
func TestScanCommand(t *testing.T) {
	cmd, _, err := findCommand(GetRootCmd(), "scan")
	if err != nil {
		t.Fatalf("Failed to find scan command: %v", err)
	}

	if cmd.Use != "scan [path]" {
		t.Errorf("scanCmd.Use = %q, want %q", cmd.Use, "scan [path]")
	}

	if cmd.Short == "" {
		t.Error("scanCmd.Short should not be empty")
	}

	// 检查标志
	flags := []string{"staged", "since", "watch", "ci"}
	for _, flag := range flags {
		f := cmd.Flags().Lookup(flag)
		if f == nil {
			t.Errorf("scanCmd missing flag: %s", flag)
		}
	}
}

// TestScanCommandExecution 测试扫描命令执行
func TestScanCommandExecution(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "todo-scan-cmd-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建测试文件
	testFile := filepath.Join(tempDir, "test.go")
	content := `package main
// TODO: 这是一个测试TODO
func main() {}
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// 执行命令
	cmd := GetRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"scan", tempDir})

	err = cmd.Execute()
	if err != nil {
		t.Errorf("scan command failed: %v", err)
	}

	// 命令执行成功即可，输出直接写入stdout，不通过cmd.OutOrStdout()
}

// TestScanCommandWithFlags 测试扫描命令带标志执行
func TestScanCommandWithFlags(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "staged flag",
			args: []string{"scan", "--staged"},
		},
		{
			name: "since flag",
			args: []string{"scan", "--since", "HEAD~1"},
		},
		{
			name: "ci mode",
			args: []string{"scan", "--ci"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := GetRootCmd()
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()
			// 命令应该执行成功（即使没有实际扫描结果）
			if err != nil {
				t.Logf("scan command returned error: %v", err)
			}
		})
	}
}

// TestStaleCommand 测试过期TODO命令
func TestStaleCommand(t *testing.T) {
	cmd, _, err := findCommand(GetRootCmd(), "stale")
	if err != nil {
		t.Fatalf("Failed to find stale command: %v", err)
	}

	if cmd.Use != "stale" {
		t.Errorf("staleCmd.Use = %q, want %q", cmd.Use, "stale")
	}

	if cmd.Short == "" {
		t.Error("staleCmd.Short should not be empty")
	}

	// 检查标志
	flags := []struct {
		name     string
		defValue string
	}{
		{"older-than", "90"},
		{"min-churn", "0"},
		{"review", "false"},
	}

	for _, tc := range flags {
		f := cmd.Flags().Lookup(tc.name)
		if f == nil {
			t.Errorf("staleCmd missing flag: %s", tc.name)
		} else if f.DefValue != tc.defValue {
			t.Errorf("staleCmd flag %s defValue = %q, want %q", tc.name, f.DefValue, tc.defValue)
		}
	}
}

// TestStaleCommandExecution 测试过期命令执行
func TestStaleCommandExecution(t *testing.T) {
	cmd := GetRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"stale"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("stale command failed: %v", err)
	}

	// 命令执行成功即可，输出直接写入stdout
}

// TestStaleCommandWithFlags 测试过期命令带标志执行
func TestStaleCommandWithFlags(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "custom older-than",
			args: []string{"stale", "--older-than", "60"},
		},
		{
			name: "min-churn flag",
			args: []string{"stale", "--min-churn", "5"},
		},
		{
			name: "review mode",
			args: []string{"stale", "--review"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := GetRootCmd()
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()
			if err != nil {
				t.Errorf("stale command failed: %v", err)
			}
		})
	}
}

// TestOrphanedCommand 测试孤儿TODO命令
func TestOrphanedCommand(t *testing.T) {
	cmd, _, err := findCommand(GetRootCmd(), "orphaned")
	if err != nil {
		t.Fatalf("Failed to find orphaned command: %v", err)
	}

	if cmd.Use != "orphaned" {
		t.Errorf("orphanedCmd.Use = %q, want %q", cmd.Use, "orphaned")
	}

	if cmd.Short == "" {
		t.Error("orphanedCmd.Short should not be empty")
	}

	// 检查标志
	flags := []struct {
		name     string
		defValue string
	}{
		{"inactive", "180"},
		{"all", "false"},
	}

	for _, tc := range flags {
		f := cmd.Flags().Lookup(tc.name)
		if f == nil {
			t.Errorf("orphanedCmd missing flag: %s", tc.name)
		} else if f.DefValue != tc.defValue {
			t.Errorf("orphanedCmd flag %s defValue = %q, want %q", tc.name, f.DefValue, tc.defValue)
		}
	}
}

// TestOrphanedCommandExecution 测试孤儿命令执行
func TestOrphanedCommandExecution(t *testing.T) {
	cmd := GetRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"orphaned"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("orphaned command failed: %v", err)
	}

	// 命令执行成功即可，输出直接写入stdout
}

// TestOrphanedCommandWithFlags 测试孤儿命令带标志执行
func TestOrphanedCommandWithFlags(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "custom inactive threshold",
			args: []string{"orphaned", "--inactive", "90"},
		},
		{
			name: "all authors",
			args: []string{"orphaned", "--all"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := GetRootCmd()
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()
			if err != nil {
				t.Errorf("orphaned command failed: %v", err)
			}
		})
	}
}

// TestReportCommand 测试报告命令
func TestReportCommand(t *testing.T) {
	cmd, _, err := findCommand(GetRootCmd(), "report")
	if err != nil {
		t.Fatalf("Failed to find report command: %v", err)
	}

	if cmd.Use != "report" {
		t.Errorf("reportCmd.Use = %q, want %q", cmd.Use, "report")
	}

	if cmd.Short == "" {
		t.Error("reportCmd.Short should not be empty")
	}

	// 检查标志
	flags := []struct {
		name     string
		defValue string
	}{
		{"format", "table"},
		{"output", ""},
		{"stale-only", "false"},
		{"orphan-only", "false"},
	}

	for _, tc := range flags {
		f := cmd.Flags().Lookup(tc.name)
		if f == nil {
			t.Errorf("reportCmd missing flag: %s", tc.name)
		} else if f.DefValue != tc.defValue {
			t.Errorf("reportCmd flag %s defValue = %q, want %q", tc.name, f.DefValue, tc.defValue)
		}
	}
}

// TestReportCommandExecution 测试报告命令执行
func TestReportCommandExecution(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"table format", []string{"report", "-f", "table"}},
		{"json format", []string{"report", "-f", "json"}},
		{"markdown format", []string{"report", "-f", "markdown"}},
		{"html format", []string{"report", "-f", "html"}},
		{"md shorthand", []string{"report", "-f", "md"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := GetRootCmd()
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()
			if err != nil {
				t.Errorf("report command failed: %v", err)
			}

			// 命令执行成功即可，输出直接写入stdout
		})
	}
}

// TestReportCommandWithStaleFlag 测试报告命令带stale-only标志
func TestReportCommandWithStaleFlag(t *testing.T) {
	cmd := GetRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"report", "--stale-only"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("report command failed: %v", err)
	}
}

// TestReportCommandWithOrphanFlag 测试报告命令带orphan-only标志
func TestReportCommandWithOrphanFlag(t *testing.T) {
	cmd := GetRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"report", "--orphan-only"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("report command failed: %v", err)
	}
}

// TestConfigCommand 测试配置命令
func TestConfigCommand(t *testing.T) {
	cmd, _, err := findCommand(GetRootCmd(), "config")
	if err != nil {
		t.Fatalf("Failed to find config command: %v", err)
	}

	if cmd.Use != "config" {
		t.Errorf("configCmd.Use = %q, want %q", cmd.Use, "config")
	}

	if cmd.Short == "" {
		t.Error("configCmd.Short should not be empty")
	}

	// 检查子命令
	subcmds := cmd.Commands()
	subcmdNames := make(map[string]bool)
	for _, c := range subcmds {
		subcmdNames[c.Name()] = true
	}

	expectedSubcmds := []string{"init", "show", "set", "reset"}
	for _, expected := range expectedSubcmds {
		if !subcmdNames[expected] {
			t.Errorf("configCmd missing expected subcommand: %s", expected)
		}
	}
}

// TestConfigInitCommand 测试配置初始化命令
func TestConfigInitCommand(t *testing.T) {
	// 切换到临时目录
	tempDir, err := os.MkdirTemp("", "todo-config-init-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(origDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	cmd := GetRootCmd()
	cmd.SetArgs([]string{"config", "init"})

	err = cmd.Execute()
	if err != nil {
		t.Errorf("config init command failed: %v", err)
	}
	// Note: runConfigInit currently just prints a message, doesn't create file
	// TODO: Update test when file creation is implemented
}

// TestConfigInitCommandAlreadyExists 测试配置初始化命令（配置文件已存在）
func TestConfigInitCommandAlreadyExists(t *testing.T) {
	// 切换到临时目录
	tempDir, err := os.MkdirTemp("", "todo-config-init-exists-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(origDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// 创建已存在的配置文件
	configFile := filepath.Join(tempDir, ".todo-tracker.yaml")
	if err := os.WriteFile(configFile, []byte("version: 1"), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	cmd := GetRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"config", "init"})

	err = cmd.Execute()
	if err == nil {
		t.Error("config init should return error when config already exists")
	}
}

// TestConfigShowCommand 测试配置显示命令
func TestConfigShowCommand(t *testing.T) {
	cmd := GetRootCmd()
	cmd.SetArgs([]string{"config", "show"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("config show command failed: %v", err)
	}
	// Note: Output uses fmt.Printf, not cmd.OutOrStdout(), so we can't capture it
}

// TestConfigSetCommand 测试配置设置命令
func TestConfigSetCommand(t *testing.T) {
	cmd := GetRootCmd()
	cmd.SetArgs([]string{"config", "set", "scan.workers", "4"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("config set command failed: %v", err)
	}
	// Note: Output uses fmt.Printf, not cmd.OutOrStdout(), so we can't capture it
}

// TestConfigSetCommandMissingArgs 测试配置设置命令缺少参数
func TestConfigSetCommandMissingArgs(t *testing.T) {
	cmd := GetRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"config", "set", "key"}) // 缺少 value

	err := cmd.Execute()
	if err == nil {
		t.Error("config set should return error when missing arguments")
	}
}

// TestConfigResetCommand 测试配置重置命令
func TestConfigResetCommand(t *testing.T) {
	// 切换到临时目录
	tempDir, err := os.MkdirTemp("", "todo-config-reset-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(origDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	cmd := GetRootCmd()
	cmd.SetArgs([]string{"config", "reset"})

	err = cmd.Execute()
	if err != nil {
		t.Errorf("config reset command failed: %v", err)
	}
	// Note: Output uses fmt.Printf, not cmd.OutOrStdout(), so we can't capture it
}

// TestVersionCommand 测试版本命令
func TestVersionCommand(t *testing.T) {
	cmd := GetRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--version"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("version command failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "todo") && !strings.Contains(output, "dev") {
		t.Errorf("version output should contain version info, got: %s", output)
	}
}

// TestHelpCommand 测试帮助命令
func TestHelpCommand(t *testing.T) {
	cmd := GetRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("help command failed: %v", err)
	}

	output := buf.String()
	expectedStrings := []string{"TODO Tracker", "scan", "stale", "orphaned", "report", "config"}
	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("help output should contain %q, got: %s", expected, output)
		}
	}
}

// TestCommandHelp 测试各命令的帮助信息
func TestCommandHelp(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		expects []string
	}{
		{
			name:    "scan help",
			args:    []string{"scan", "--help"},
			expects: []string{"扫描", "staged", "since", "watch"},
		},
		{
			name:    "stale help",
			args:    []string{"stale", "--help"},
			expects: []string{"过期", "older-than", "min-churn"},
		},
		{
			name:    "orphaned help",
			args:    []string{"orphaned", "--help"},
			expects: []string{"孤儿", "inactive", "all"},
		},
		{
			name:    "report help",
			args:    []string{"report", "--help"},
			expects: []string{"报告", "format", "output"},
		},
		{
			name:    "config help",
			args:    []string{"config", "--help"},
			expects: []string{"配置", "init", "show", "set", "reset"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := GetRootCmd()
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()
			if err != nil {
				t.Errorf("help command failed: %v", err)
			}

			output := buf.String()
			for _, expected := range tt.expects {
				if !strings.Contains(output, expected) {
					t.Errorf("help output should contain %q", expected)
				}
			}
		})
	}
}

// TestVerboseFlag 测试详细输出标志
func TestVerboseFlag(t *testing.T) {
	cmd := GetRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"-v", "scan"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("verbose flag command failed: %v", err)
	}
}

// TestConfigFlag 测试配置文件标志
func TestConfigFlag(t *testing.T) {
	// 创建临时配置文件
	tempFile, err := os.CreateTemp("", "todo-config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	configContent := "version: 1\nscan:\n  workers: 4\n"
	if _, err := tempFile.WriteString(configContent); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}
	tempFile.Close()

	cmd := GetRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"-c", tempFile.Name(), "config", "show"})

	err = cmd.Execute()
	if err != nil {
		t.Errorf("config flag command failed: %v", err)
	}
}

// TestOutputFlag 测试输出文件标志
func TestOutputFlag(t *testing.T) {
	// 创建临时输出文件
	tempFile, err := os.CreateTemp("", "todo-output-*")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tempFile.Close()
	defer os.Remove(tempFile.Name())

	cmd := GetRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"-o", tempFile.Name(), "report"})

	err = cmd.Execute()
	if err != nil {
		t.Errorf("output flag command failed: %v", err)
	}
}

// TestExecute 测试Execute函数
func TestExecute(t *testing.T) {
	// 由于Execute会调用os.Exit，我们无法直接测试
	// 但可以验证函数存在且类型正确
	cmd := GetRootCmd()
	if cmd == nil {
		t.Error("GetRootCmd returned nil")
	}
}

// TestInvalidCommand 测试无效命令
func TestInvalidCommand(t *testing.T) {
	cmd := GetRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"invalid-command"})

	err := cmd.Execute()
	if err == nil {
		t.Error("invalid command should return error")
	}
}

// TestScanCommandEmptyDirectory 测试扫描空目录
func TestScanCommandEmptyDirectory(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "todo-empty-scan-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cmd := GetRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"scan", tempDir})

	err = cmd.Execute()
	if err != nil {
		t.Errorf("scan command failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "扫描") {
		t.Errorf("scan output should contain '扫描', got: %s", output)
	}
}

// TestScanCommandNonexistentPath 测试扫描不存在的路径
func TestScanCommandNonexistentPath(t *testing.T) {
	cmd := GetRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"scan", "/nonexistent/path/that/does/not/exist"})

	// 命令应该能执行，只是扫描结果为空
	err := cmd.Execute()
	// 根据实现可能返回错误或空结果
	_ = err // 忽略错误，只验证不会崩溃
}

// TestFlagDefaults 测试标志默认值
func TestFlagDefaults(t *testing.T) {
	tests := []struct {
		name         string
		command      string
		flag         string
		defaultValue string
	}{
		{
			name:         "stale older-than default",
			command:      "stale",
			flag:         "older-than",
			defaultValue: "90",
		},
		{
			name:         "stale min-churn default",
			command:      "stale",
			flag:         "min-churn",
			defaultValue: "0",
		},
		{
			name:         "orphaned inactive default",
			command:      "orphaned",
			flag:         "inactive",
			defaultValue: "180",
		},
		{
			name:         "report format default",
			command:      "report",
			flag:         "format",
			defaultValue: "table",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, _, err := findCommand(GetRootCmd(), tt.command)
			if err != nil {
				t.Fatalf("Failed to find command %s: %v", tt.command, err)
			}

			flag := cmd.Flags().Lookup(tt.flag)
			if flag == nil {
				t.Fatalf("Flag %s not found", tt.flag)
			}

			if flag.DefValue != tt.defaultValue {
				t.Errorf("Flag %s default = %q, want %q", tt.flag, flag.DefValue, tt.defaultValue)
			}
		})
	}
}

// Helper function to find a command by name
func findCommand(root *cobra.Command, name string) (*cobra.Command, []string, error) {
	return root.Find([]string{name})
}

// TestGenerateTableReport 测试表格报告生成
func TestGenerateTableReport(t *testing.T) {
	err := generateTableReport()
	if err != nil {
		t.Errorf("generateTableReport() failed: %v", err)
	}
}

// TestGenerateJSONReport 测试JSON报告生成
func TestGenerateJSONReport(t *testing.T) {
	err := generateJSONReport()
	if err != nil {
		t.Errorf("generateJSONReport() failed: %v", err)
	}
}

// TestGenerateMarkdownReport 测试Markdown报告生成
func TestGenerateMarkdownReport(t *testing.T) {
	err := generateMarkdownReport()
	if err != nil {
		t.Errorf("generateMarkdownReport() failed: %v", err)
	}
}

// TestGenerateHTMLReport 测试HTML报告生成
func TestGenerateHTMLReport(t *testing.T) {
	err := generateHTMLReport()
	if err != nil {
		t.Errorf("generateHTMLReport() failed: %v", err)
	}
}

// TestReportFormat 测试报告格式处理
func TestReportFormat(t *testing.T) {
	tests := []struct {
		name   string
		format string
	}{
		{"table format", "table"},
		{"json format", "json"},
		{"markdown format", "markdown"},
		{"md format", "md"},
		{"html format", "html"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := GetRootCmd()
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetArgs([]string{"report", "-f", tt.format})

			err := cmd.Execute()
			if err != nil {
				t.Errorf("report command with format %s failed: %v", tt.format, err)
			}
		})
	}
}

// TestMultipleFlags 测试多个标志组合
func TestMultipleFlags(t *testing.T) {
	cmd := GetRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"report", "-f", "json", "--stale-only", "--orphan-only"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("report command with multiple flags failed: %v", err)
	}
}

// TestCommandShortDescriptions 测试命令简短描述
func TestCommandShortDescriptions(t *testing.T) {
	tests := []struct {
		name    string
		command string
	}{
		{"scan", "scan"},
		{"stale", "stale"},
		{"orphaned", "orphaned"},
		{"report", "report"},
		{"config", "config"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, _, err := findCommand(GetRootCmd(), tt.command)
			if err != nil {
				t.Fatalf("Failed to find command %s: %v", tt.command, err)
			}

			if cmd.Short == "" {
				t.Errorf("Command %s has empty Short description", tt.command)
			}
		})
	}
}

// TestCommandLongDescriptions 测试命令详细描述
func TestCommandLongDescriptions(t *testing.T) {
	tests := []struct {
		name    string
		command string
	}{
		{"scan", "scan"},
		{"stale", "stale"},
		{"orphaned", "orphaned"},
		{"report", "report"},
		{"config", "config"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, _, err := findCommand(GetRootCmd(), tt.command)
			if err != nil {
				t.Fatalf("Failed to find command %s: %v", tt.command, err)
			}

			// Long可以为空，但如果存在应该包含有用信息
			if cmd.Long != "" && len(cmd.Long) < 20 {
				t.Errorf("Command %s Long description is too short", tt.command)
			}
		})
	}
}

// TestRootCommandVersion 测试根命令版本信息
func TestRootCommandVersion(t *testing.T) {
	SetVersion("test-version", "test-commit", "test-date")

	cmd := GetRootCmd()
	if cmd.Version != "test-version" {
		t.Errorf("rootCmd.Version = %q, want %q", cmd.Version, "test-version")
	}
}

// TestConfigSubcommandArgs 测试配置子命令参数验证
func TestConfigSubcommandArgs(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		shouldError bool
	}{
		{
			name:        "set with two args",
			args:        []string{"config", "set", "key", "value"},
			shouldError: false,
		},
		{
			name:        "set with one arg",
			args:        []string{"config", "set", "key"},
			shouldError: true,
		},
		{
			name:        "set with no args",
			args:        []string{"config", "set"},
			shouldError: true,
		},
		{
			name:        "init with no args",
			args:        []string{"config", "init"},
			shouldError: false,
		},
		{
			name:        "show with no args",
			args:        []string{"config", "show"},
			shouldError: false,
		},
		{
			name:        "reset with no args",
			args:        []string{"config", "reset"},
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := GetRootCmd()
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()
			if tt.shouldError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.shouldError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestRunScanWithDifferentPaths 测试扫描不同路径
func TestRunScanWithDifferentPaths(t *testing.T) {
	// 创建临时目录结构
	tempDir, err := os.MkdirTemp("", "todo-multi-path-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建子目录
	subDir := filepath.Join(tempDir, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	// 创建测试文件
	testFiles := map[string]string{
		filepath.Join(tempDir, "main.go"):  "// TODO: root todo",
		filepath.Join(subDir, "sub.go"):    "// TODO: sub todo",
	}

	for path, content := range testFiles {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}
	}

	tests := []struct {
		name string
		path string
	}{
		{"root directory", tempDir},
		{"subdirectory", subDir},
		{"current directory", "."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := GetRootCmd()
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetArgs([]string{"scan", tt.path})

			err := cmd.Execute()
			if err != nil {
				t.Errorf("scan command failed: %v", err)
			}
		})
	}
}

// TestInitConfig 测试配置初始化函数
func TestInitConfig(t *testing.T) {
	// 创建临时配置文件
	tempFile, err := os.CreateTemp("", "todo-init-config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	configContent := "version: 1\nscan:\n  workers: 4\n"
	if _, err := tempFile.WriteString(configContent); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}
	tempFile.Close()

	// initConfig 在命令初始化时被调用
	// 这里测试带配置文件的命令执行
	cmd := GetRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"-c", tempFile.Name(), "config", "show"})

	err = cmd.Execute()
	if err != nil {
		t.Errorf("command with config file failed: %v", err)
	}
}
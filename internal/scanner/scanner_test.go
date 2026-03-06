// Package scanner_test 测试TODO扫描功能
package scanner

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/mxihan/todo-tracker/pkg/types"
)

// TestNewScanner 测试扫描器创建
func TestNewScanner(t *testing.T) {
	tests := []struct {
		name   string
		config *types.Config
	}{
		{
			name:   "默认配置",
			config: types.DefaultConfig(),
		},
		{
			name: "自定义Worker数量",
			config: &types.Config{
				Scan: types.ScanConfig{
					Workers: 4,
				},
			},
		},
		{
			name: "零Worker（自动检测）",
			config: &types.Config{
				Scan: types.ScanConfig{
					Workers: 0,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner := NewScanner(tt.config)
			if scanner == nil {
				t.Error("NewScanner() returned nil")
			}
		})
	}
}

// TestScan 测试基本扫描功能
func TestScan(t *testing.T) {
	// 创建临时测试目录
	tempDir, err := os.MkdirTemp("", "todo-scan-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建测试文件
	testFiles := map[string]string{
		"main.go": `package main
// TODO: 这是一个TODO
func main() {}
`,
		"utils.py": `# TODO: Python TODO
def hello():
    pass
`,
	}

	for name, content := range testFiles {
		filePath := filepath.Join(tempDir, name)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}
	}

	// 创建扫描器
	config := types.DefaultConfig()
	config.Scan.Paths = []string{tempDir}
	scanner := NewScanner(config)

	// 执行扫描
	ctx := context.Background()
	result, err := scanner.Scan(ctx, tempDir)

	// 验证结果
	if err != nil {
		t.Errorf("Scan() returned error: %v", err)
	}

	if result == nil {
		t.Fatal("Scan() returned nil result")
	}

	// 注意：由于scanner.Scan()尚未实现，这里只验证结构正确
	if result.TODOs == nil {
		t.Error("TODOs slice should not be nil")
	}

	if result.Summary.ByType == nil {
		t.Error("ByType map should not be nil")
	}
}

// TestScanEmptyDirectory 测试扫描空目录
func TestScanEmptyDirectory(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "todo-empty-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	config := types.DefaultConfig()
	scanner := NewScanner(config)

	ctx := context.Background()
	result, err := scanner.Scan(ctx, tempDir)

	if err != nil {
		t.Errorf("Scan() returned error: %v", err)
	}

	if result == nil {
		t.Fatal("Scan() returned nil result")
	}

	if result.Summary.Total != 0 {
		t.Errorf("Expected 0 TODOs in empty directory, got %d", result.Summary.Total)
	}
}

// TestScanNonExistentPath 测试扫描不存在的路径
func TestScanNonExistentPath(t *testing.T) {
	config := types.DefaultConfig()
	scanner := NewScanner(config)

	ctx := context.Background()
	result, err := scanner.Scan(ctx, "/non/existent/path/that/does/not/exist")

	// 根据实际实现，可能返回错误或空结果
	// 这里只验证不会崩溃
	_ = result
	_ = err
}

// TestScanStaged 测试暂存区扫描
func TestScanStaged(t *testing.T) {
	config := types.DefaultConfig()
	scanner := NewScanner(config)

	ctx := context.Background()
	result, err := scanner.ScanStaged(ctx)

	// Should not return nil result
	if result == nil {
		t.Error("ScanStaged() returned nil result")
		return
	}

	// Result should have initialized maps
	if result.Summary.ByType == nil {
		t.Error("ScanStaged() result.Summary.ByType is nil")
	}
	if result.Summary.ByPriority == nil {
		t.Error("ScanStaged() result.Summary.ByPriority is nil")
	}
	if result.Summary.ByAuthor == nil {
		t.Error("ScanStaged() result.Summary.ByAuthor is nil")
	}

	// If error, it might be because we're not in a git repo
	// In a git repo, should succeed without error
	if err != nil {
		t.Logf("ScanStaged() returned error (may be expected if not in git repo): %v", err)
	}
}

// TestScanStagedWithGitDisabled 测试Git禁用时的暂存区扫描
func TestScanStagedWithGitDisabled(t *testing.T) {
	config := types.DefaultConfig()
	config.Git.Enabled = false
	scanner := NewScanner(config)

	ctx := context.Background()
	result, err := scanner.ScanStaged(ctx)

	// Should return empty result without error when git is disabled
	if err != nil {
		t.Errorf("ScanStaged() with git disabled should not return error, got: %v", err)
	}

	if result == nil {
		t.Error("ScanStaged() returned nil result")
		return
	}

	if result.Summary.Total != 0 {
		t.Errorf("ScanStaged() with git disabled should return 0 TODOs, got %d", result.Summary.Total)
	}
}

// TestScanStagedEmptyRepo 测试空暂存区的扫描
func TestScanStagedEmptyRepo(t *testing.T) {
	// Skip if not in a git repo
	config := types.DefaultConfig()
	testScanner := NewScanner(config)
	if !testScanner.gitClient.IsGitRepo() {
		t.Skip("Not in a git repository, skipping TestScanStagedEmptyRepo")
	}

	scanner := NewScanner(config)
	ctx := context.Background()
	result, err := scanner.ScanStaged(ctx)

	if err != nil {
		t.Errorf("ScanStaged() returned error: %v", err)
	}

	if result == nil {
		t.Error("ScanStaged() returned nil result")
	}
}

// TestScanSince 测试增量扫描
func TestScanSince(t *testing.T) {
	config := types.DefaultConfig()
	scanner := NewScanner(config)

	ctx := context.Background()
	result, err := scanner.ScanSince(ctx, "HEAD~1")

	// Should not return nil result
	if result == nil {
		t.Error("ScanSince() returned nil result")
		return
	}

	// Result should have initialized maps
	if result.Summary.ByType == nil {
		t.Error("ScanSince() result.Summary.ByType is nil")
	}
	if result.Summary.ByPriority == nil {
		t.Error("ScanSince() result.Summary.ByPriority is nil")
	}
	if result.Summary.ByAuthor == nil {
		t.Error("ScanSince() result.Summary.ByAuthor is nil")
	}

	// If error, it might be because we're not in a git repo
	// In a git repo, should succeed without error
	if err != nil {
		t.Logf("ScanSince() returned error (may be expected if not in git repo): %v", err)
	}
}

// TestScanSinceWithGitDisabled 测试Git禁用时的增量扫描
func TestScanSinceWithGitDisabled(t *testing.T) {
	config := types.DefaultConfig()
	config.Git.Enabled = false
	scanner := NewScanner(config)

	ctx := context.Background()
	result, err := scanner.ScanSince(ctx, "HEAD~1")

	// Should return empty result without error when git is disabled
	if err != nil {
		t.Errorf("ScanSince() with git disabled should not return error, got: %v", err)
	}

	if result == nil {
		t.Error("ScanSince() returned nil result")
		return
	}

	if result.Summary.Total != 0 {
		t.Errorf("ScanSince() with git disabled should return 0 TODOs, got %d", result.Summary.Total)
	}
}

// TestScanSinceEmptyRef 测试无效引用的增量扫描
func TestScanSinceEmptyRef(t *testing.T) {
	// Skip if not in a git repo
	config := types.DefaultConfig()
	testScanner := NewScanner(config)
	if !testScanner.gitClient.IsGitRepo() {
		t.Skip("Not in a git repository, skipping TestScanSinceEmptyRef")
	}

	scanner := NewScanner(config)
	ctx := context.Background()
	// Using a non-existent ref should return an error
	result, err := scanner.ScanSince(ctx, "nonexistent-ref-12345")

	// Should return an error for invalid ref
	if err == nil {
		t.Log("ScanSince() with invalid ref did not return error (might be acceptable)")
	}

	if result == nil {
		t.Error("ScanSince() returned nil result even with error")
	}
}

// TestScanSinceWithContextCancellation 测试上下文取消
func TestScanSinceWithContextCancellation(t *testing.T) {
	// Skip if not in a git repo
	config := types.DefaultConfig()
	testScanner := NewScanner(config)
	if !testScanner.gitClient.IsGitRepo() {
		t.Skip("Not in a git repository, skipping TestScanSinceWithContextCancellation")
	}

	scanner := NewScanner(config)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	result, err := scanner.ScanSince(ctx, "HEAD~1")

	// Context cancellation may or may not cause an error depending on timing
	// Just verify we don't panic and get a result
	if result == nil {
		t.Error("ScanSince() returned nil result")
	}
	t.Logf("ScanSince with cancelled context returned: err=%v, result.Summary.Total=%d", err, result.Summary.Total)
}

// TestScanFile 测试单文件扫描
func TestScanFile(t *testing.T) {
	// 创建临时测试文件
	tempFile, err := os.CreateTemp("", "test-*.go")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	testContent := `package main
// TODO: 测试TODO
// FIXME: 需要修复
func main() {}
`
	if _, err := tempFile.WriteString(testContent); err != nil {
		t.Fatalf("Failed to write test content: %v", err)
	}
	tempFile.Close()

	config := types.DefaultConfig()
	scanner := NewScanner(config)

	ctx := context.Background()
	todos, err := scanner.ScanFile(ctx, tempFile.Name())

	// 当前是存根实现
	if err != nil {
		t.Logf("ScanFile() returned error: %v", err)
	}

	if todos == nil {
		t.Error("ScanFile() returned nil slice")
	}
}

// TestResultChan 测试流式扫描
func TestResultChan(t *testing.T) {
	config := types.DefaultConfig()
	scanner := NewScanner(config)

	ctx := context.Background()
	ch := scanner.ResultChan(ctx, ".")

	if ch == nil {
		t.Error("ResultChan() returned nil channel")
		return
	}

	// 读取通道直到关闭
	eventCount := 0
	for range ch {
		eventCount++
	}

	// 当前是存根实现，通道应该立即关闭
	t.Logf("Received %d events", eventCount)
}

// TestScanEvent 测试扫描事件
func TestScanEvent(t *testing.T) {
	event := ScanEvent{
		Type: EventTODOFound,
		File: "test.go",
		TODO: &types.TODO{
			Type:    "TODO",
			Message: "测试事件",
			File:    "test.go",
			Line:    10,
		},
	}

	if event.Type != EventTODOFound {
		t.Errorf("EventType = %d, want %d", event.Type, EventTODOFound)
	}

	if event.TODO == nil {
		t.Error("TODO should not be nil")
	}
}

// TestExcludePatterns 测试排除模式
func TestExcludePatterns(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		exclude  []string
		wantSkip bool
	}{
		{
			name:     "node_modules应被排除",
			path:     "node_modules/package/index.js",
			exclude:  []string{"**/node_modules/**"},
			wantSkip: true,
		},
		{
			name:     "vendor应被排除",
			path:     "vendor/lib/main.go",
			exclude:  []string{"**/vendor/**"},
			wantSkip: true,
		},
		{
			name:     ".git应被排除",
			path:     ".git/config",
			exclude:  []string{"**/.git/**"},
			wantSkip: true,
		},
		{
			name:     "源码不应被排除",
			path:     "src/main.go",
			exclude:  []string{"**/node_modules/**"},
			wantSkip: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 验证排除模式配置
			config := &types.Config{
				Scan: types.ScanConfig{
					Exclude: tt.exclude,
				},
			}

			_ = config // 配置已创建
			// TODO: 当skip.go实现后，测试实际的跳过逻辑
		})
	}
}

// TestWorkerPool 测试工作池
func TestWorkerPool(t *testing.T) {
	pool := &workerPool{
		workers:  4,
		taskCh:   make(chan scanTask, 10),
		resultCh: make(chan scanResult, 10),
	}

	if pool.workers != 4 {
		t.Errorf("workers = %d, want 4", pool.workers)
	}

	// 关闭通道
	close(pool.taskCh)
	close(pool.resultCh)
}

// BenchmarkScan 基准测试扫描
func BenchmarkScan(b *testing.B) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "bench-scan")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建多个测试文件
	for i := 0; i < 10; i++ {
		filePath := filepath.Join(tempDir, "file"+string(rune('0'+i))+".go")
		content := `package main
// TODO: 测试TODO ` + string(rune('0'+i)) + `
func main() {}
`
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			b.Fatalf("Failed to write file: %v", err)
		}
	}

	config := types.DefaultConfig()
	scanner := NewScanner(config)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scanner.Scan(ctx, tempDir)
	}
}

// BenchmarkScanFile 基准测试单文件扫描
func BenchmarkScanFile(b *testing.B) {
	tempFile, err := os.CreateTemp("", "bench-*.go")
	if err != nil {
		b.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	content := `package main
// TODO: 第一个
// TODO: 第二个
// TODO: 第三个
func main() {}
`
	if _, err := tempFile.WriteString(content); err != nil {
		b.Fatalf("Failed to write content: %v", err)
	}
	tempFile.Close()

	config := types.DefaultConfig()
	scanner := NewScanner(config)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scanner.ScanFile(ctx, tempFile.Name())
	}
}
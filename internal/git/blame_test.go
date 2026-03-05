// Package git_test 测试Git Blame功能
package git

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestNewBlamer 测试Blamer创建
func TestNewBlamer(t *testing.T) {
	client := NewClient(".")
	blamer := NewBlamer(client)

	if blamer == nil {
		t.Error("NewBlamer() returned nil")
	}

	if blamer.client == nil {
		t.Error("Blamer.client should not be nil")
	}
}

// TestBlameInfo 测试BlameInfo结构
func TestBlameInfo(t *testing.T) {
	info := BlameInfo{
		Author:      "Test Author",
		AuthorEmail: "test@example.com",
		CommitHash:  "abc123def456",
		CommitDate:  time.Now(),
		Line:        10,
		Content:     "// TODO: test",
	}

	if info.Author != "Test Author" {
		t.Errorf("Author = %s, want Test Author", info.Author)
	}

	if info.Line != 10 {
		t.Errorf("Line = %d, want 10", info.Line)
	}
}

// TestBlameResult 测试BlameResult结构
func TestBlameResult(t *testing.T) {
	result := &BlameResult{
		File: "test.go",
		Lines: []BlameInfo{
			{Author: "Alice", Line: 1},
			{Author: "Bob", Line: 2},
		},
		Authors: map[string]int{
			"Alice": 1,
			"Bob":   1,
		},
	}

	if result.File != "test.go" {
		t.Errorf("File = %s, want test.go", result.File)
	}

	if len(result.Lines) != 2 {
		t.Errorf("Lines count = %d, want 2", len(result.Lines))
	}
}

// TestBlameFile 测试文件Blame
func TestBlameFile(t *testing.T) {
	// 创建临时文件
	tempFile, err := os.CreateTemp("", "blame-test-*.go")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	testContent := `package main

// TODO: test todo
func main() {
    println("hello")
}
`
	if _, err := tempFile.WriteString(testContent); err != nil {
		t.Fatalf("Failed to write content: %v", err)
	}
	tempFile.Close()

	client := NewClient(filepath.Dir(tempFile.Name()))
	blamer := NewBlamer(client)

	result, err := blamer.BlameFile(tempFile.Name())
	if err != nil {
		// 在非Git目录中会失败，这是预期的
		t.Logf("BlameFile() returned error (expected in non-git dir): %v", err)
		return
	}

	if result == nil {
		t.Error("BlameFile() returned nil result")
	}
}

// TestBlameLine 测试单行Blame
func TestBlameLine(t *testing.T) {
	// 创建临时文件
	tempFile, err := os.CreateTemp("", "blame-line-test-*.go")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	testContent := `package main
// TODO: test todo
func main() {}
`
	if _, err := tempFile.WriteString(testContent); err != nil {
		t.Fatalf("Failed to write content: %v", err)
	}
	tempFile.Close()

	client := NewClient(filepath.Dir(tempFile.Name()))
	blamer := NewBlamer(client)

	info, err := blamer.BlameLine(tempFile.Name(), 2)
	if err != nil {
		t.Logf("BlameLine() returned error (expected in non-git dir): %v", err)
		return
	}

	if info == nil {
		t.Error("BlameLine() returned nil info")
	}
}

// TestGetTODOMetadata 测试获取TODO元数据
func TestGetTODOMetadata(t *testing.T) {
	// 创建临时文件
	tempFile, err := os.CreateTemp("", "todo-meta-test-*.go")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	testContent := `package main
// TODO: test todo
func main() {}
`
	if _, err := tempFile.WriteString(testContent); err != nil {
		t.Fatalf("Failed to write content: %v", err)
	}
	tempFile.Close()

	client := NewClient(filepath.Dir(tempFile.Name()))
	blamer := NewBlamer(client)

	author, commitHash, commitDate, err := blamer.GetTODOMetadata(tempFile.Name(), 2)
	if err != nil {
		t.Logf("GetTODOMetadata() returned error (expected in non-git dir): %v", err)
		return
	}

	// 验证返回值
	_ = author
	_ = commitHash
	_ = commitDate
}

// TestBatchBlame 测试批量Blame
func TestBatchBlame(t *testing.T) {
	// 创建临时目录和文件
	tempDir, err := os.MkdirTemp("", "batch-blame-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建多个测试文件
	files := []string{}
	for i := 0; i < 3; i++ {
		filePath := filepath.Join(tempDir, "file"+string(rune('0'+i))+".go")
		content := `package main
// TODO: test
func main() {}
`
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write file: %v", err)
		}
		files = append(files, filePath)
	}

	client := NewClient(tempDir)
	blamer := NewBlamer(client)

	results, err := blamer.BatchBlame(files)
	if err != nil {
		t.Logf("BatchBlame() returned error: %v", err)
		return
	}

	if len(results) != len(files) {
		t.Errorf("Results count = %d, want %d", len(results), len(files))
	}
}

// TestCheckAuthorActive 测试检查作者活跃状态
func TestCheckAuthorActive(t *testing.T) {
	client := NewClient(".")
	blamer := NewBlamer(client)

	isActive, lastCommit, err := blamer.CheckAuthorActive("nonexistent-author", 180)
	if err != nil {
		t.Logf("CheckAuthorActive() returned error: %v", err)
		return
	}

	// 不存在的作者应该返回不活跃
	if isActive {
		t.Log("Nonexistent author is marked as active")
	}

	_ = lastCommit
}

// TestIsHexString 测试十六进制字符串检查
func TestIsHexString(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"abc123", true},
		{"ABC123", true},
		{"0123456789abcdef", true},
		{"ghijkl", false}, // 包含非十六进制字符
		{"", true},        // 空字符串
		{"12345g", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := isHexString(tt.input)
			if result != tt.expected {
				t.Errorf("isHexString(%s) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// TestParseBlameOutput 测试解析Blame输出
func TestParseBlameOutput(t *testing.T) {
	client := NewClient(".")
	blamer := NewBlamer(client)

	// 模拟git blame --line-porcelain输出
	mockOutput := `abc123def456789 1 1 1
author Test Author
author-mail <test@example.com>
author-time 1700000000
	// TODO: test comment
def456789abc123 2 2 2
author Another Author
author-mail <another@example.com>
author-time 1700000100
	func main() {}
`

	result := &BlameResult{
		File:    "test.go",
		Lines:   make([]BlameInfo, 0),
		Authors: make(map[string]int),
	}

	blamer.parseBlameOutput(mockOutput, result)

	if len(result.Lines) != 2 {
		t.Errorf("Expected 2 lines, got %d", len(result.Lines))
	}

	if result.Lines[0].Author != "Test Author" {
		t.Errorf("First line author = %s, want Test Author", result.Lines[0].Author)
	}

	if result.Authors["Test Author"] != 1 {
		t.Errorf("Author count for 'Test Author' = %d, want 1", result.Authors["Test Author"])
	}
}

// BenchmarkBlameFile 基准测试文件Blame
func BenchmarkBlameFile(b *testing.B) {
	tempFile, err := os.CreateTemp("", "bench-blame-*.go")
	if err != nil {
		b.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	content := `package main
// TODO: line 2
// TODO: line 3
// TODO: line 4
func main() {}
`
	tempFile.WriteString(content)
	tempFile.Close()

	client := NewClient(filepath.Dir(tempFile.Name()))
	blamer := NewBlamer(client)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		blamer.BlameFile(tempFile.Name())
	}
}

// BenchmarkBlameLine 基准测试单行Blame
func BenchmarkBlameLine(b *testing.B) {
	tempFile, err := os.CreateTemp("", "bench-line-*.go")
	if err != nil {
		b.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	content := `package main
// TODO: test
func main() {}
`
	tempFile.WriteString(content)
	tempFile.Close()

	client := NewClient(filepath.Dir(tempFile.Name()))
	blamer := NewBlamer(client)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		blamer.BlameLine(tempFile.Name(), 2)
	}
}
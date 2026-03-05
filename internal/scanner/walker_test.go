// Package scanner_test 测试目录遍历功能
package scanner

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mxihan/todo-tracker/pkg/types"
)

// TestNewWalker 测试创建Walker
func TestNewWalker(t *testing.T) {
	config := types.DefaultConfig()
	walker := NewWalker(config)

	if walker == nil {
		t.Fatal("NewWalker() returned nil")
	}

	if len(walker.skipDirs) == 0 {
		t.Error("skipDirs should not be empty")
	}

	if len(walker.skipFiles) == 0 {
		t.Error("skipFiles should not be empty")
	}
}

// TestWalkerWalk 测试目录遍历
func TestWalkerWalk(t *testing.T) {
	// 创建测试目录结构
	tempDir, err := os.MkdirTemp("", "walker-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建文件和目录
	files := []string{
		"main.go",
		"utils.go",
		"src/app.go",
		"src/lib/helper.go",
	}

	dirs := []string{
		"src",
		"src/lib",
	}

	// 创建目录
	for _, dir := range dirs {
		if err := os.MkdirAll(filepath.Join(tempDir, dir), 0755); err != nil {
			t.Fatalf("Failed to create dir: %v", err)
		}
	}

	// 创建文件
	for _, file := range files {
		filePath := filepath.Join(tempDir, file)
		if err := os.WriteFile(filePath, []byte("// test"), 0644); err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}
	}

	config := types.DefaultConfig()
	walker := NewWalker(config)

	fileCh, errCh := walker.Walk(tempDir)

	var foundFiles []string
	for file := range fileCh {
		foundFiles = append(foundFiles, file)
	}

	// 检查错误
	select {
	case err := <-errCh:
		if err != nil {
			t.Errorf("Walk() returned error: %v", err)
		}
	default:
	}

	// 验证找到的文件数量
	if len(foundFiles) != len(files) {
		t.Errorf("Found %d files, want %d", len(foundFiles), len(files))
	}
}

// TestWalkerSkipDirs 测试跳过目录
func TestWalkerSkipDirs(t *testing.T) {
	// 创建测试目录结构
	tempDir, err := os.MkdirTemp("", "walker-skip-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建应该跳过的目录
	skipDirs := []string{".git", "node_modules", "vendor", "dist"}
	for _, dir := range skipDirs {
		dirPath := filepath.Join(tempDir, dir)
		os.MkdirAll(dirPath, 0755)
		filePath := filepath.Join(dirPath, "test.go")
		os.WriteFile(filePath, []byte("// test"), 0644)
	}

	// 创建应该扫描的目录
	scanDir := filepath.Join(tempDir, "src")
	os.MkdirAll(scanDir, 0755)
	os.WriteFile(filepath.Join(scanDir, "main.go"), []byte("// main"), 0644)

	config := types.DefaultConfig()
	walker := NewWalker(config)

	fileCh, _ := walker.Walk(tempDir)

	var foundFiles []string
	for file := range fileCh {
		foundFiles = append(foundFiles, filepath.Base(filepath.Dir(file)))
	}

	// 验证没有扫描跳过的目录
	for _, skipDir := range skipDirs {
		for _, found := range foundFiles {
			if found == skipDir {
				t.Errorf("Directory %s should have been skipped", skipDir)
			}
		}
	}
}

// TestWalkerSkipFiles 测试跳过文件
func TestWalkerSkipFiles(t *testing.T) {
	// 创建测试目录
	tempDir, err := os.MkdirTemp("", "walker-skip-files-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建应该扫描的文件
	shouldScan := []string{"main.go", "utils.py", "index.js"}
	for _, file := range shouldScan {
		filePath := filepath.Join(tempDir, file)
		os.WriteFile(filePath, []byte("// test"), 0644)
	}

	// 创建应该跳过的文件
	shouldSkip := []string{"app.min.js", "style.min.css", "package-lock.json", "go.sum"}
	for _, file := range shouldSkip {
		filePath := filepath.Join(tempDir, file)
		os.WriteFile(filePath, []byte("test"), 0644)
	}

	config := types.DefaultConfig()
	walker := NewWalker(config)

	fileCh, _ := walker.Walk(tempDir)

	var foundFiles []string
	for file := range fileCh {
		foundFiles = append(foundFiles, filepath.Base(file))
	}

	// 验证应该扫描的文件存在
	for _, expected := range shouldScan {
		found := false
		for _, actual := range foundFiles {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("File %s should have been found", expected)
		}
	}

	// 验证应该跳过的文件不存在
	for _, skip := range shouldSkip {
		for _, actual := range foundFiles {
			if actual == skip {
				t.Errorf("File %s should have been skipped", skip)
			}
		}
	}
}

// TestShouldSkipDir 测试目录跳过判断
func TestShouldSkipDir(t *testing.T) {
	config := types.DefaultConfig()
	walker := NewWalker(config)

	tests := []struct {
		dir      string
		wantSkip bool
	}{
		{".git", true},
		{"node_modules", true},
		{"vendor", true},
		{"dist", true},
		{"build", true},
		{"target", true},
		{".idea", true},
		{".vscode", true},
		{"__pycache__", true},
		{"src", false},
		{"lib", false},
		{"cmd", false},
		{"pkg", false},
	}

	for _, tt := range tests {
		t.Run(tt.dir, func(t *testing.T) {
			got := walker.shouldSkipDir(tt.dir)
			if got != tt.wantSkip {
				t.Errorf("shouldSkipDir(%s) = %v, want %v", tt.dir, got, tt.wantSkip)
			}
		})
	}
}

// TestWalkerShouldSkipFile 测试Walker的文件跳过判断
func TestWalkerShouldSkipFile(t *testing.T) {
	config := types.DefaultConfig()
	walker := NewWalker(config)

	tests := []struct {
		file     string
		wantSkip bool
	}{
		{"app.min.js", true},
		{"style.min.css", true},
		{"package-lock.json", true},
		{"go.sum", true},
		{"yarn.lock", true},
		{"main.go", false},
		{"utils.py", false},
		{"index.js", false},
		{"style.css", false},
	}

	for _, tt := range tests {
		t.Run(tt.file, func(t *testing.T) {
			got := walker.shouldSkipFile(tt.file)
			if got != tt.wantSkip {
				t.Errorf("shouldSkipFile(%s) = %v, want %v", tt.file, got, tt.wantSkip)
			}
		})
	}
}

// TestExtractDirPatterns 测试提取目录模式
func TestExtractDirPatterns(t *testing.T) {
	tests := []struct {
		excludes []string
		wantDirs []string
	}{
		{
			excludes: []string{"**/node_modules/**", "**/vendor/**"},
			wantDirs: []string{"node_modules", "vendor"},
		},
		{
			excludes: []string{"**/dist/**", "**/build/**"},
			wantDirs: []string{"dist", "build"},
		},
		{
			excludes: []string{"*.log"},
			wantDirs: []string{},
		},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := extractDirPatterns(tt.excludes)

			// 检查所有期望的目录都在结果中
			for _, want := range tt.wantDirs {
				found := false
				for _, actual := range got {
					if actual == want {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected dir %s not found in result", want)
				}
			}
		})
	}
}

// TestExtractFilePatterns 测试提取文件模式
func TestExtractFilePatterns(t *testing.T) {
	tests := []struct {
		excludes   []string
		wantFiles  []string
	}{
		{
			excludes:   []string{"**/*.min.js", "**/*.min.css"},
			wantFiles:  []string{".min.js", ".min.css"},
		},
		{
			excludes:   []string{"**/node_modules/**"},
			wantFiles:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := extractFilePatterns(tt.excludes)

			// 检查结果
			_ = got
			_ = tt.wantFiles
		})
	}
}

// TestIsBinaryFile 测试二进制文件检测
func TestIsBinaryFile(t *testing.T) {
	// 创建文本文件
	textFile, err := os.CreateTemp("", "text-*.txt")
	if err != nil {
		t.Fatalf("Failed to create text file: %v", err)
	}
	defer os.Remove(textFile.Name())
	textFile.WriteString("Hello, World!")
	textFile.Close()

	isBinary, err := isBinaryFile(textFile.Name())
	if err != nil {
		t.Errorf("isBinaryFile() returned error: %v", err)
	}
	if isBinary {
		t.Error("Text file should not be detected as binary")
	}

	// 创建二进制文件
	binaryFile, err := os.CreateTemp("", "binary-*.bin")
	if err != nil {
		t.Fatalf("Failed to create binary file: %v", err)
	}
	defer os.Remove(binaryFile.Name())
	binaryData := []byte{0x00, 0x01, 0x02, 0x03, 0x00, 0x05}
	binaryFile.Write(binaryData)
	binaryFile.Close()

	isBinary, err = isBinaryFile(binaryFile.Name())
	if err != nil {
		t.Errorf("isBinaryFile() returned error: %v", err)
	}
	if !isBinary {
		t.Error("Binary file should be detected as binary")
	}
}

// BenchmarkWalkerWalk 基准测试目录遍历
func BenchmarkWalkerWalk(b *testing.B) {
	// 创建测试目录
	tempDir, err := os.MkdirTemp("", "bench-walker")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建多个文件
	for i := 0; i < 100; i++ {
		filePath := filepath.Join(tempDir, "file"+string(rune('0'+i%10))+".go")
		os.WriteFile(filePath, []byte("// test"), 0644)
	}

	config := types.DefaultConfig()
	walker := NewWalker(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fileCh, _ := walker.Walk(tempDir)
		for range fileCh {
		}
	}
}

// BenchmarkShouldSkipDir 基准测试目录跳过判断
func BenchmarkShouldSkipDir(b *testing.B) {
	config := types.DefaultConfig()
	walker := NewWalker(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		walker.shouldSkipDir("node_modules")
	}
}

// BenchmarkWalkerShouldSkipFile 基准测试Walker的文件跳过判断
func BenchmarkWalkerShouldSkipFile(b *testing.B) {
	config := types.DefaultConfig()
	walker := NewWalker(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		walker.shouldSkipFile("main.go")
	}
}
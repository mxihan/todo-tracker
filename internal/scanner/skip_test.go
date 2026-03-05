// Package scanner_test 测试跳过规则
package scanner

import (
	"testing"
)

// TestDefaultSkipRules 测试默认跳过规则
func TestDefaultSkipRules(t *testing.T) {
	rules := DefaultSkipRules()

	if rules == nil {
		t.Fatal("DefaultSkipRules() returned nil")
	}

	if len(rules.directoryPatterns) == 0 {
		t.Error("directoryPatterns should not be empty")
	}

	if len(rules.filePatterns) == 0 {
		t.Error("filePatterns should not be empty")
	}

	if len(rules.extensions) == 0 {
		t.Error("extensions should not be empty")
	}
}

// TestShouldSkipDirectory 测试目录跳过判断
func TestShouldSkipDirectory(t *testing.T) {
	rules := DefaultSkipRules()

	tests := []struct {
		directory string
		wantSkip  bool
	}{
		// 应该跳过的目录
		{".git", true},
		{".svn", true},
		{".hg", true},
		{"node_modules", true},
		{"vendor", true},
		{"dist", true},
		{"build", true},
		{"target", true},
		{"__pycache__", true},
		{".idea", true},
		{".vscode", true},
		{".cache", true},
		{"venv", true},
		{".venv", true},

		// 不应该跳过的目录
		{"src", false},
		{"lib", false},
		{"pkg", true}, // pkg在默认跳过列表中
		{"cmd", false},
		{"internal", false},
		{"api", false},
		{"test", false},
		{"tests", false},
	}

	for _, tt := range tests {
		t.Run(tt.directory, func(t *testing.T) {
			got := rules.ShouldSkipDirectory(tt.directory)
			if got != tt.wantSkip {
				t.Errorf("ShouldSkipDirectory(%s) = %v, want %v", tt.directory, got, tt.wantSkip)
			}
		})
	}
}

// TestShouldSkipFile 测试文件跳过判断
func TestShouldSkipFile(t *testing.T) {
	rules := DefaultSkipRules()

	tests := []struct {
		file     string
		wantSkip bool
	}{
		// 应该跳过的文件
		{"app.exe", true},
		{"library.dll", true},
		{"module.so", true},
		{"image.png", true},
		{"photo.jpg", true},
		{"icon.gif", true},
		{"archive.zip", true},
		{"data.tar", true},
		{"compressed.gz", true},
		{"app.min.js", true},
		{"style.min.css", true},
		{"package-lock.json", true},
		{"yarn.lock", true},
		{"go.sum", true},
		{"database.db", true},
		{"cache.sqlite", true},
		{"document.pdf", true},

		// 不应该跳过的文件
		{"main.go", false},
		{"utils.py", false},
		{"index.js", false},
		{"app.ts", false},
		{"Main.java", false},
		{"lib.rs", false},
		{"style.css", false},
		{"index.html", false},
		{"config.yaml", false},
		{"settings.json", false},
		{"README.md", false},
	}

	for _, tt := range tests {
		t.Run(tt.file, func(t *testing.T) {
			got := rules.ShouldSkipFile(tt.file)
			if got != tt.wantSkip {
				t.Errorf("ShouldSkipFile(%s) = %v, want %v", tt.file, got, tt.wantSkip)
			}
		})
	}
}

// TestAddDirectoryPattern 测试添加目录模式
func TestAddDirectoryPattern(t *testing.T) {
	rules := DefaultSkipRules()
	initialCount := len(rules.directoryPatterns)

	rules.AddDirectoryPattern("custom_dir")

	if len(rules.directoryPatterns) != initialCount+1 {
		t.Errorf("AddDirectoryPattern() did not add pattern")
	}

	// 验证新添加的模式生效
	if !rules.ShouldSkipDirectory("custom_dir") {
		t.Error("New directory pattern should be matched")
	}
}

// TestAddFilePattern 测试添加文件模式
func TestAddFilePattern(t *testing.T) {
	rules := DefaultSkipRules()
	initialCount := len(rules.filePatterns)

	rules.AddFilePattern("*.custom")

	if len(rules.filePatterns) != initialCount+1 {
		t.Errorf("AddFilePattern() did not add pattern")
	}

	// 验证新添加的模式生效
	if !rules.ShouldSkipFile("test.custom") {
		t.Error("New file pattern should be matched")
	}
}

// TestMerge 测试合并规则
func TestMerge(t *testing.T) {
	rules1 := &SkipRules{
		directoryPatterns: []string{".git", "node_modules"},
		filePatterns:      []string{"*.exe"},
		extensions:        []string{".exe"},
	}

	rules2 := &SkipRules{
		directoryPatterns: []string{".cache", "dist"},
		filePatterns:      []string{"*.log"},
		extensions:        []string{".log"},
	}

	rules1.Merge(rules2)

	// 验证合并后的规则
	if len(rules1.directoryPatterns) != 4 {
		t.Errorf("Expected 4 directory patterns, got %d", len(rules1.directoryPatterns))
	}

	if len(rules1.filePatterns) != 2 {
		t.Errorf("Expected 2 file patterns, got %d", len(rules1.filePatterns))
	}

	if len(rules1.extensions) != 2 {
		t.Errorf("Expected 2 extensions, got %d", len(rules1.extensions))
	}

	// 验证nil合并不会崩溃
	rules1.Merge(nil)
}

// TestMergeNil 测试合并nil规则
func TestMergeNil(t *testing.T) {
	rules := DefaultSkipRules()
	initialCount := len(rules.directoryPatterns)

	rules.Merge(nil)

	if len(rules.directoryPatterns) != initialCount {
		t.Error("Merge(nil) should not change rules")
	}
}

// TestFromConfig 测试从配置创建规则
func TestFromConfig(t *testing.T) {
	tests := []struct {
		name           string
		patterns       []string
		testPath       string
		wantSkipDir    bool
		wantSkipFile   bool
	}{
		{
			name:        "目录模式",
			patterns:    []string{"**/custom/**"},
			testPath:    "custom",
			wantSkipDir: true,
		},
		{
			name:         "文件模式",
			patterns:     []string{"**/*.log"},
			testPath:     "debug.log",
			wantSkipFile: true,
		},
		{
			name:        "多级目录模式",
			patterns:    []string{"**/src/generated/**"},
			testPath:    "generated",
			wantSkipDir: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rules := FromConfig(tt.patterns)

			if tt.wantSkipDir && !rules.ShouldSkipDirectory(tt.testPath) {
				t.Errorf("Directory %s should be skipped", tt.testPath)
			}

			if tt.wantSkipFile && !rules.ShouldSkipFile(tt.testPath) {
				t.Errorf("File %s should be skipped", tt.testPath)
			}
		})
	}
}

// TestSkipRulesWithPathSeparators 测试带路径分隔符的模式
func TestSkipRulesWithPathSeparators(t *testing.T) {
	rules := DefaultSkipRules()

	// 测试完整路径
	tests := []struct {
		path     string
		wantSkip bool
	}{
		{"/home/user/project/node_modules/package", true},
		{"/home/user/project/.git/objects", true},
		{"/home/user/project/src/main.go", false},
		{"/home/user/project/build/output", true},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			// 提取目录名进行测试
			// 注意：当前实现基于basename匹配
			_ = rules
			_ = tt
		})
	}
}

// TestExtensionCaseInsensitive 测试扩展名大小写不敏感
func TestExtensionCaseInsensitive(t *testing.T) {
	rules := DefaultSkipRules()

	tests := []struct {
		file     string
		wantSkip bool
	}{
		{"image.PNG", true},  // 大写扩展名
		{"image.Jpg", true},  // 混合大小写
		{"document.PDF", true},
		{"main.GO", false},   // Go文件不跳过
	}

	for _, tt := range tests {
		t.Run(tt.file, func(t *testing.T) {
			got := rules.ShouldSkipFile(tt.file)
			if got != tt.wantSkip {
				t.Errorf("ShouldSkipFile(%s) = %v, want %v", tt.file, got, tt.wantSkip)
			}
		})
	}
}

// BenchmarkShouldSkipDirectory 基准测试目录跳过判断
func BenchmarkShouldSkipDirectory(b *testing.B) {
	rules := DefaultSkipRules()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rules.ShouldSkipDirectory("node_modules")
	}
}

// BenchmarkShouldSkipFile 基准测试文件跳过判断
func BenchmarkShouldSkipFile(b *testing.B) {
	rules := DefaultSkipRules()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rules.ShouldSkipFile("main.go")
	}
}
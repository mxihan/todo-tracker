// Package scanner_test 集成测试 - 端到端扫描测试
package scanner

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/mxihan/todo-tracker/internal/parser"
	"github.com/mxihan/todo-tracker/pkg/types"
)

// TestEndToEndScan 测试端到端扫描流程
func TestEndToEndScan(t *testing.T) {
	// 创建测试项目结构
	tempDir, err := os.MkdirTemp("", "e2e-scan-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建测试文件
	testFiles := map[string]string{
		"src/main.go": `package main
// TODO: 主函数待实现
// TODO!: 紧急修复
// TODO(@alice): 分配给alice
func main() {
	// FIXME: 这里需要修复
}
`,
		"src/utils.py": `# TODO: Python文件待办
# FIXME: 需要修复
def helper():
    pass
`,
		"src/index.html": `<!-- TODO: HTML注释中的待办 -->
<html>
<body>
<!-- FIXME: 需要修复的问题 -->
</body>
</html>
`,
		"src/style.css": `/* TODO: CSS待办 */
body {
    margin: 0;
}
`,
	}

	for path, content := range testFiles {
		fullPath := filepath.Join(tempDir, path)
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create dir: %v", err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write file: %v", err)
		}
	}

	// 创建扫描器和解析器
	config := types.DefaultConfig()
	config.Scan.Paths = []string{tempDir}
	scanner := NewScanner(config)

	ctx := context.Background()
	result, err := scanner.Scan(ctx, tempDir)
	if err != nil {
		t.Errorf("Scan() returned error: %v", err)
	}

	// 验证结果结构
	if result == nil {
		t.Fatal("Scan() returned nil result")
	}

	t.Logf("Scan result: %d TODOs found", result.Summary.Total)
}

// TestParserEndToEnd 测试解析器端到端
func TestParserEndToEnd(t *testing.T) {
	// 创建测试文件
	testContent := `package main

// TODO: 基本TODO
// TODO!: 高优先级TODO
// TODO(@alice): 分配给alice
// TODO(#123): 关联Issue
// TODO(JIRA-456): 关联Jira
// TODO(@bob) #789!: 组合格式

// FIXME: 需要修复
// HACK: 临时方案
// BUG: 已知问题
// XXX: 警告标记

func main() {
	// TODO: 函数内的TODO
	/*
	   TODO: 多行注释中的TODO
	   FIXME: 多行FIXME
	*/
}
`

	p := parser.NewParser(nil)
	todos := p.ParseFile(testContent, "test.go")

	// 验证解析结果
	expectedMinCount := 10 // 至少应该解析出10个TODO
	if len(todos) < expectedMinCount {
		t.Errorf("ParseFile() found %d TODOs, expected at least %d", len(todos), expectedMinCount)
	}

	// 统计各类型
	typeCount := make(map[string]int)
	priorityCount := make(map[string]int)

	for _, todo := range todos {
		typeCount[todo.Type]++
		priorityCount[todo.Priority]++
	}

	t.Logf("Type distribution: %v", typeCount)
	t.Logf("Priority distribution: %v", priorityCount)

	// 验证TODO类型存在
	if typeCount["TODO"] == 0 {
		t.Error("No TODO type found")
	}

	if typeCount["FIXME"] == 0 {
		t.Error("No FIXME type found")
	}
}

// TestMultiLanguageScan 测试多语言扫描
func TestMultiLanguageScan(t *testing.T) {
	p := parser.NewParser(nil)

	tests := []struct {
		name          string
		content       string
		filePath      string
		expectedTypes map[string]bool
	}{
		{
			name: "Go文件",
			content: `package main
// TODO: Go注释
func main() {}
`,
			filePath:      "main.go",
			expectedTypes: map[string]bool{"TODO": true},
		},
		{
			name: "Python文件",
			content: `# TODO: Python注释
# FIXME: 需要修复
def main():
    pass
`,
			filePath:      "main.py",
			expectedTypes: map[string]bool{"TODO": true, "FIXME": true},
		},
		{
			name: "JavaScript文件",
			content: `// TODO: JS注释
/* FIXME: 块注释 */
function main() {}
`,
			filePath:      "main.js",
			expectedTypes: map[string]bool{"TODO": true, "FIXME": true},
		},
		{
			name: "HTML文件",
			content: `<!-- TODO: HTML注释 -->
<html></html>
`,
			filePath:      "index.html",
			expectedTypes: map[string]bool{"TODO": true},
		},
		{
			name: "SQL文件",
			content: `-- TODO: SQL注释
SELECT * FROM users;
`,
			filePath:      "query.sql",
			expectedTypes: map[string]bool{"TODO": true},
		},
		{
			name: "Shell文件",
			content: `#!/bin/bash
# TODO: Shell注释
echo "hello"
`,
			filePath:      "script.sh",
			expectedTypes: map[string]bool{"TODO": true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			todos := p.ParseFile(tt.content, tt.filePath)

			foundTypes := make(map[string]bool)
			for _, todo := range todos {
				foundTypes[todo.Type] = true
			}

			for expectedType := range tt.expectedTypes {
				if !foundTypes[expectedType] {
					t.Errorf("Expected type %s not found in parsed TODOs", expectedType)
				}
			}
		})
	}
}

// TestPriorityParsing 测试优先级解析
func TestPriorityParsing(t *testing.T) {
	p := parser.NewParser(nil)

	tests := []struct {
		content        string
		expectedPriority string
	}{
		{"// TODO!: 高优先级", "high"},
		{"// TODO URGENT: 紧急", "high"},
		{"// TODO CRITICAL: 关键", "high"},
		{"// TODO>: 中等优先级", "medium"},
		{"// TODO MEDIUM: 中等", "medium"},
		{"// TODO: 低优先级", "low"},
	}

	for _, tt := range tests {
		t.Run(tt.content, func(t *testing.T) {
			todos := p.ParseFile(tt.content, "test.go")
			if len(todos) == 0 {
				t.Error("No TODO parsed")
				return
			}

			if todos[0].Priority != tt.expectedPriority {
				t.Errorf("Priority = %s, want %s", todos[0].Priority, tt.expectedPriority)
			}
		})
	}
}

// TestMetadataParsing 测试元数据解析
func TestMetadataParsing(t *testing.T) {
	p := parser.NewParser(nil)

	tests := []struct {
		name          string
		content       string
		wantAssignee  string
		wantTicket    string
	}{
		{
			name:         "带负责人",
			content:      "// TODO(@alice): 任务",
			wantAssignee: "alice",
		},
		{
			name:        "带Issue号",
			content:     "// TODO(#123): 任务",
			wantTicket:  "#123",
		},
		{
			name:        "带Jira号",
			content:     "// TODO(JIRA-456): 任务",
			wantTicket:  "JIRA-456",
		},
		{
			name:         "组合格式",
			content:      "// TODO(@bob) #789!: 任务",
			wantAssignee: "bob",
			wantTicket:   "#789",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			todos := p.ParseFile(tt.content, "test.go")
			if len(todos) == 0 {
				t.Error("No TODO parsed")
				return
			}

			todo := todos[0]

			if tt.wantAssignee != "" && todo.Assignee != tt.wantAssignee {
				t.Errorf("Assignee = %s, want %s", todo.Assignee, tt.wantAssignee)
			}

			if tt.wantTicket != "" && todo.TicketRef != tt.wantTicket {
				t.Errorf("TicketRef = %s, want %s", todo.TicketRef, tt.wantTicket)
			}
		})
	}
}

// TestSkipRulesIntegration 测试跳过规则集成
func TestSkipRulesIntegration(t *testing.T) {
	// 创建测试目录结构
	tempDir, err := os.MkdirTemp("", "skip-rules-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建应该跳过的目录和文件
	skipDirs := []string{"node_modules", "vendor", ".git", "dist", "build"}
	for _, dir := range skipDirs {
		dirPath := filepath.Join(tempDir, dir)
		os.MkdirAll(dirPath, 0755)

		// 在每个目录中创建包含TODO的文件
		filePath := filepath.Join(dirPath, "test.go")
		content := `// TODO: should be skipped`
		os.WriteFile(filePath, []byte(content), 0644)
	}

	// 创建应该扫描的目录
	scanDir := filepath.Join(tempDir, "src")
	os.MkdirAll(scanDir, 0755)
	filePath := filepath.Join(scanDir, "main.go")
	content := `// TODO: should be found`
	os.WriteFile(filePath, []byte(content), 0644)

	// 测试跳过规则
	rules := DefaultSkipRules()

	for _, dir := range skipDirs {
		if !rules.ShouldSkipDirectory(dir) {
			t.Errorf("Directory %s should be skipped", dir)
		}
	}

	if rules.ShouldSkipDirectory("src") {
		t.Error("Directory 'src' should not be skipped")
	}
}

// TestScanWithRealTestdata 测试使用真实测试数据
func TestScanWithRealTestdata(t *testing.T) {
	// 查找testdata目录
	testdataPath := filepath.Join("..", "..", "testdata", "project1")
	if _, err := os.Stat(testdataPath); os.IsNotExist(err) {
		t.Skip("testdata directory not found")
	}

	p := parser.NewParser(nil)

	// 测试main.go
	mainPath := filepath.Join(testdataPath, "main.go")
	if content, err := os.ReadFile(mainPath); err == nil {
		todos := p.ParseFile(string(content), mainPath)
		t.Logf("main.go: found %d TODOs", len(todos))
		for _, todo := range todos {
			t.Logf("  - [%s] %s at line %d", todo.Type, todo.Message, todo.Line)
		}
	}

	// 测试utils.py
	utilsPath := filepath.Join(testdataPath, "utils.py")
	if content, err := os.ReadFile(utilsPath); err == nil {
		todos := p.ParseFile(string(content), utilsPath)
		t.Logf("utils.py: found %d TODOs", len(todos))
		for _, todo := range todos {
			t.Logf("  - [%s] %s at line %d", todo.Type, todo.Message, todo.Line)
		}
	}

	// 测试index.html
	htmlPath := filepath.Join(testdataPath, "index.html")
	if content, err := os.ReadFile(htmlPath); err == nil {
		todos := p.ParseFile(string(content), htmlPath)
		t.Logf("index.html: found %d TODOs", len(todos))
		for _, todo := range todos {
			t.Logf("  - [%s] %s at line %d", todo.Type, todo.Message, todo.Line)
		}
	}
}

// TestConcurrentScan 测试并发扫描
func TestConcurrentScan(t *testing.T) {
	// 创建大量测试文件
	tempDir, err := os.MkdirTemp("", "concurrent-scan-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建50个测试文件
	for i := 0; i < 50; i++ {
		filePath := filepath.Join(tempDir, "file"+string(rune('0'+i%10))+".go")
		content := `package main
// TODO: test todo ` + string(rune('A'+i%26)) + `
func main() {}
`
		os.WriteFile(filePath, []byte(content), 0644)
	}

	config := types.DefaultConfig()
	config.Scan.Workers = 4 // 使用4个worker
	scanner := NewScanner(config)

	ctx := context.Background()
	result, err := scanner.Scan(ctx, tempDir)

	if err != nil {
		t.Errorf("Concurrent scan failed: %v", err)
	}

	if result == nil {
		t.Error("Scan returned nil result")
	}

	t.Logf("Scanned %d files concurrently", result.Summary.FilesScanned)
}
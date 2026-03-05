// Package parser_test 测试TODO解析功能
package parser

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/todo-tracker/todo-tracker/pkg/types"
)

// TestNewParser 测试解析器创建
func TestNewParser(t *testing.T) {
	tests := []struct {
		name   string
		config *types.PatternConfig
	}{
		{
			name:   "默认配置",
			config: nil,
		},
		{
			name: "自定义配置",
			config: &types.PatternConfig{
				Types:          []string{"TODO", "FIXME"},
				AssigneePattern: `@(\w+)`,
				TicketPattern:   `#(\d+)`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser(tt.config)
			if parser == nil {
				t.Error("NewParser() returned nil")
			}
		})
	}
}

// TestParseFile 测试文件解析
func TestParseFile(t *testing.T) {
	parser := NewParser(nil)

	tests := []struct {
		name          string
		content       string
		filePath      string
		wantCount     int
		wantTypes     map[string]bool
		wantPriority  map[string]int
	}{
		{
			name: "基本TODO",
			content: `package main
// TODO: 这是一个基本的TODO
func main() {}
`,
			filePath:  "main.go",
			wantCount: 1,
			wantTypes: map[string]bool{"TODO": true},
		},
		{
			name: "多种类型",
			content: `// TODO: 第一个
// FIXME: 第二个
// HACK: 第三个
`,
			filePath:  "test.go",
			wantCount: 3,
			wantTypes: map[string]bool{"TODO": true, "FIXME": true, "HACK": true},
		},
		{
			name: "高优先级TODO",
			content: `// TODO!: 这是高优先级
// FIXME!: 也是高优先级
`,
			filePath:     "urgent.go",
			wantCount:    2,
			wantPriority: map[string]int{"high": 2},
		},
		{
			name: "带负责人的TODO",
			content: `// TODO(@alice): 分配给alice
// TODO(bob): 分配给bob
`,
			filePath:  "assigned.go",
			wantCount: 2,
		},
		{
			name: "带工单号的TODO",
			content: `// TODO(#123): 关联Issue
// TODO(JIRA-456): 关联Jira
`,
			filePath:  "linked.go",
			wantCount: 2,
		},
		{
			name: "组合格式",
			content: `// TODO(@alice) #789!: 组合格式
`,
			filePath:  "combined.go",
			wantCount: 1,
		},
		{
			name: "空文件",
			content: `package main
`,
			filePath:  "empty.go",
			wantCount: 0,
		},
		{
			name: "无TODO",
			content: `package main
// 这是一个普通注释
func main() {}
`,
			filePath:  "notodo.go",
			wantCount: 0,
		},
		{
			name: "Python风格",
			content: `# TODO: Python注释
# FIXME: 需要修复
`,
			filePath:  "script.py",
			wantCount: 2,
			wantTypes: map[string]bool{"TODO": true, "FIXME": true},
		},
		{
			name: "HTML风格",
			content: `<!-- TODO: HTML注释 -->
<div>content</div>
<!-- FIXME: 需要修复 -->
`,
			filePath:  "page.html",
			wantCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			todos := parser.ParseFile(tt.content, tt.filePath)

			if len(todos) != tt.wantCount {
				t.Errorf("ParseFile() returned %d todos, want %d", len(todos), tt.wantCount)
			}

			// 验证类型
			if tt.wantTypes != nil {
				for _, todo := range todos {
					if !tt.wantTypes[todo.Type] {
						t.Errorf("Unexpected TODO type: %s", todo.Type)
					}
				}
			}

			// 验证优先级
			if tt.wantPriority != nil {
				priorityCount := make(map[string]int)
				for _, todo := range todos {
					priorityCount[todo.Priority]++
				}
				for priority, count := range tt.wantPriority {
					if priorityCount[priority] != count {
						t.Errorf("Priority %s count = %d, want %d", priority, priorityCount[priority], count)
					}
				}
			}
		})
	}
}

// TestParseLine 测试单行解析
func TestParseLine(t *testing.T) {
	parser := NewParser(nil)
	langConfig := &LanguageConfig{
		SingleLine: []string{"//"},
	}

	tests := []struct {
		name         string
		line         string
		wantNil      bool
		wantType     string
		wantMessage  string
		wantPriority string
		wantAssignee string
		wantTicket   string
	}{
		{
			name:        "基本TODO",
			line:        "// TODO: 这是一个测试",
			wantNil:     false,
			wantType:    "TODO",
			wantMessage: "这是一个测试",
		},
		{
			name:         "高优先级",
			line:         "// TODO!: 高优先级任务",
			wantNil:      false,
			wantType:     "TODO",
			wantPriority: "high",
			wantMessage:  "高优先级任务",
		},
		{
			name:         "中等优先级",
			line:         "// TODO>: 中等优先级",
			wantNil:      false,
			wantType:     "TODO",
			wantPriority: "medium",
		},
		{
			name:         "带负责人",
			line:         "// TODO(@alice): 分配的任务",
			wantNil:      false,
			wantType:     "TODO",
			wantAssignee: "alice",
		},
		{
			name:       "带GitHub Issue",
			line:       "// TODO(#123): 关联Issue",
			wantNil:    false,
			wantType:   "TODO",
			wantTicket: "#123",
		},
		{
			name:       "带Jira工单",
			line:       "// TODO(JIRA-456): Jira任务",
			wantNil:    false,
			wantType:   "TODO",
			wantTicket: "JIRA-456",
		},
		{
			name:    "非TODO注释",
			line:    "// 这是一个普通注释",
			wantNil: true,
		},
		{
			name:    "空行",
			line:    "",
			wantNil: true,
		},
		{
			name:     "FIXME",
			line:     "// FIXME: 需要修复",
			wantNil:  false,
			wantType: "FIXME",
		},
		{
			name:     "HACK",
			line:     "// HACK: 临时方案",
			wantNil:  false,
			wantType: "HACK",
		},
		{
			name:     "BUG",
			line:     "// BUG: 已知问题",
			wantNil:  false,
			wantType: "BUG",
		},
		{
			name:     "XXX",
			line:     "// XXX: 警告标记",
			wantNil:  false,
			wantType: "XXX",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			todo := parser.parseLine(tt.line, "test.go", 1, langConfig)

			if tt.wantNil {
				if todo != nil {
					t.Errorf("parseLine() should return nil, got %+v", todo)
				}
				return
			}

			if todo == nil {
				t.Error("parseLine() returned nil, expected non-nil")
				return
			}

			if todo.Type != tt.wantType {
				t.Errorf("Type = %s, want %s", todo.Type, tt.wantType)
			}

			if tt.wantMessage != "" && todo.Message != tt.wantMessage {
				t.Errorf("Message = %s, want %s", todo.Message, tt.wantMessage)
			}

			if tt.wantPriority != "" && todo.Priority != tt.wantPriority {
				t.Errorf("Priority = %s, want %s", todo.Priority, tt.wantPriority)
			}

			if tt.wantAssignee != "" && todo.Assignee != tt.wantAssignee {
				t.Errorf("Assignee = %s, want %s", todo.Assignee, tt.wantAssignee)
			}

			if tt.wantTicket != "" && todo.TicketRef != tt.wantTicket {
				t.Errorf("TicketRef = %s, want %s", todo.TicketRef, tt.wantTicket)
			}
		})
	}
}

// TestPriorityExtraction 测试优先级提取
func TestPriorityExtraction(t *testing.T) {
	parser := NewParser(nil)
	langConfig := &LanguageConfig{
		SingleLine: []string{"//"},
	}

	tests := []struct {
		line         string
		wantPriority string
	}{
		{"// TODO!: 高优先级", "high"},
		{"// TODO URGENT: 紧急", "high"},
		{"// TODO CRITICAL: 关键", "high"},
		{"// TODO>: 中等", "medium"},
		{"// TODO MEDIUM: 中等", "medium"},
		{"// TODO: 低优先级", "low"},
	}

	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			todo := parser.parseLine(tt.line, "test.go", 1, langConfig)
			if todo == nil {
				t.Error("parseLine() returned nil")
				return
			}
			if todo.Priority != tt.wantPriority {
				t.Errorf("Priority = %s, want %s", todo.Priority, tt.wantPriority)
			}
		})
	}
}

// TestAssigneeExtraction 测试负责人提取
func TestAssigneeExtraction(t *testing.T) {
	parser := NewParser(nil)
	langConfig := &LanguageConfig{
		SingleLine: []string{"//"},
	}

	tests := []struct {
		line         string
		wantAssignee string
	}{
		{"// TODO(@alice): 任务", "alice"},
		{"// TODO(bob): 任务", "bob"},
		{"// TODO @charlie: 任务", "charlie"},
		{"// TODO(@david) #123: 任务", "david"},
	}

	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			todo := parser.parseLine(tt.line, "test.go", 1, langConfig)
			if todo == nil {
				t.Error("parseLine() returned nil")
				return
			}
			if todo.Assignee != tt.wantAssignee {
				t.Errorf("Assignee = %s, want %s", todo.Assignee, tt.wantAssignee)
			}
		})
	}
}

// TestTicketExtraction 测试工单号提取
func TestTicketExtraction(t *testing.T) {
	parser := NewParser(nil)
	langConfig := &LanguageConfig{
		SingleLine: []string{"//"},
	}

	tests := []struct {
		line       string
		wantTicket string
	}{
		{"// TODO(#123): 任务", "#123"},
		{"// TODO(#456): 任务", "#456"},
		{"// TODO(JIRA-789): 任务", "JIRA-789"},
		{"// TODO(PROJ-100): 任务", "PROJ-100"},
	}

	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			todo := parser.parseLine(tt.line, "test.go", 1, langConfig)
			if todo == nil {
				t.Error("parseLine() returned nil")
				return
			}
			if todo.TicketRef != tt.wantTicket {
				t.Errorf("TicketRef = %s, want %s", todo.TicketRef, tt.wantTicket)
			}
		})
	}
}

// TestLanguageSupport 测试语言支持检测
func TestLanguageSupport(t *testing.T) {
	tests := []struct {
		ext         string
		wantSupport bool
	}{
		{".go", true},
		{".py", true},
		{".js", true},
		{".ts", true},
		{".java", true},
		{".rs", true},
		{".html", true},
		{".css", true},
		{".unknown", false},
		{".xyz", false},
	}

	for _, tt := range tests {
		t.Run(tt.ext, func(t *testing.T) {
			got := IsSupported(tt.ext)
			if got != tt.wantSupport {
				t.Errorf("IsSupported(%s) = %v, want %v", tt.ext, got, tt.wantSupport)
			}
		})
	}
}

// TestGetLanguageByExtension 测试获取语言配置
func TestGetLanguageByExtension(t *testing.T) {
	tests := []struct {
		ext      string
		wantName string
		wantNil  bool
	}{
		{".go", "Go", false},
		{".py", "Python", false},
		{".js", "JavaScript", false},
		{".unknown", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.ext, func(t *testing.T) {
			lang := GetLanguageByExtension(tt.ext)
			if tt.wantNil {
				if lang != nil {
					t.Errorf("GetLanguageByExtension(%s) should return nil", tt.ext)
				}
				return
			}

			if lang == nil {
				t.Errorf("GetLanguageByExtension(%s) returned nil", tt.ext)
				return
			}

			if lang.Name != tt.wantName {
				t.Errorf("Name = %s, want %s", lang.Name, tt.wantName)
			}
		})
	}
}

// TestParseRealFiles 测试解析真实测试文件
func TestParseRealFiles(t *testing.T) {
	// 获取testdata目录路径
	testdataDir := filepath.Join("..", "..", "testdata", "project1")

	// 检查目录是否存在
	if _, err := os.Stat(testdataDir); os.IsNotExist(err) {
		t.Skip("testdata directory not found")
	}

	parser := NewParser(nil)

	tests := []struct {
		fileName  string
		minTodos  int
	}{
		{"main.go", 10},  // 根据实际内容调整
		{"utils.py", 8},
		{"index.html", 3},
	}

	for _, tt := range tests {
		t.Run(tt.fileName, func(t *testing.T) {
			filePath := filepath.Join(testdataDir, tt.fileName)
			content, err := os.ReadFile(filePath)
			if err != nil {
				t.Skipf("Cannot read file %s: %v", filePath, err)
			}

			todos := parser.ParseFile(string(content), filePath)
			if len(todos) < tt.minTodos {
				t.Errorf("ParseFile(%s) returned %d todos, want at least %d",
					tt.fileName, len(todos), tt.minTodos)
			}
		})
	}
}

// TestGetSupportedExtensions 测试获取支持的扩展名列表
func TestGetSupportedExtensions(t *testing.T) {
	extensions := GetSupportedExtensions()

	if len(extensions) == 0 {
		t.Error("GetSupportedExtensions() returned empty list")
	}

	// 检查一些常见的扩展名是否在列表中
	expected := []string{".go", ".py", ".js", ".java", ".rs"}
	for _, ext := range expected {
		found := false
		for _, e := range extensions {
			if e == ext {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected extension %s not found in supported extensions", ext)
		}
	}
}

// BenchmarkParseFile 基准测试文件解析
func BenchmarkParseFile(b *testing.B) {
	parser := NewParser(nil)
	content := `package main

// TODO: 第一个TODO
// FIXME: 需要修复
// TODO!: 高优先级
// TODO(@alice): 分配的任务
// TODO(#123): 关联Issue

func main() {
	// TODO: 函数内的TODO
}
`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parser.ParseFile(content, "bench.go")
	}
}

// BenchmarkParseLine 基准测试单行解析
func BenchmarkParseLine(b *testing.B) {
	parser := NewParser(nil)
	langConfig := &LanguageConfig{
		SingleLine: []string{"//"},
	}
	line := "// TODO(@alice) #123!: 这是一个测试任务"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parser.parseLine(line, "bench.go", 1, langConfig)
	}
}
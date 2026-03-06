// Package reporter_test 测试报告生成功能
package reporter

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/mxihan/todo-tracker/pkg/types"
)

// =============================================================================
// Test Helper Functions
// =============================================================================

// createTestTODO creates a TODO for testing
func createTestTODO(overrides ...func(*types.TODO)) types.TODO {
	todo := types.TODO{
		ID:          "test-id-123",
		Type:        "TODO",
		Message:     "Test TODO message",
		File:        "test.go",
		Line:        10,
		LineEnd:     12,
		Priority:    "medium",
		Assignee:    "testuser",
		TicketRef:   "#123",
		Author:      "testauthor",
		CommitHash:  "abc123",
		CreatedAt:   time.Now().AddDate(0, -1, 0), // 1 month ago
		LastModified: time.Now().AddDate(0, 0, -5),
		Status:      "open",
		Age:         30,
		ChurnScore:  5,
		IsOrphaned:  false,
	}
	for _, override := range overrides {
		override(&todo)
	}
	return todo
}

// createTestScanResult creates a ScanResult for testing
func createTestScanResult(todos []types.TODO, warnings []types.Warning) *types.ScanResult {
	byType := make(map[string]int)
	byPriority := make(map[string]int)
	byAuthor := make(map[string]int)

	for _, todo := range todos {
		byType[todo.Type]++
		byPriority[todo.Priority]++
		if todo.Author != "" {
			byAuthor[todo.Author]++
		}
	}

	return &types.ScanResult{
		Summary: types.Summary{
			Total:        len(todos),
			FilesScanned: 10,
			Duration:     time.Millisecond * 150,
			ByType:       byType,
			ByPriority:   byPriority,
			ByAuthor:     byAuthor,
		},
		TODOs:    todos,
		Warnings: warnings,
	}
}

// =============================================================================
// JSON Reporter Tests
// =============================================================================

func TestNewJSONReporter(t *testing.T) {
	tests := []struct {
		name   string
		opts   []JSONOption
		check  func(*testing.T, *JSONReporter)
	}{
		{
			name: "default options",
			opts: nil,
			check: func(t *testing.T, r *JSONReporter) {
				if r.indent != true {
					t.Error("default indent should be true")
				}
			},
		},
		{
			name: "with indent disabled",
			opts: []JSONOption{WithIndent(false)},
			check: func(t *testing.T, r *JSONReporter) {
				if r.indent != false {
					t.Error("indent should be false")
				}
			},
		},
		{
			name: "with custom writer",
			opts: []JSONOption{WithJSONWriter(&bytes.Buffer{})},
			check: func(t *testing.T, r *JSONReporter) {
				if r.writer == nil {
					t.Error("writer should not be nil")
				}
			},
		},
		{
			name: "multiple options",
			opts: []JSONOption{
				WithIndent(false),
				WithJSONWriter(&bytes.Buffer{}),
			},
			check: func(t *testing.T, r *JSONReporter) {
				if r.indent != false {
					t.Error("indent should be false")
				}
				if r.writer == nil {
					t.Error("writer should not be nil")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reporter := NewJSONReporter(tt.opts...)
			if reporter == nil {
				t.Fatal("NewJSONReporter returned nil")
			}
			tt.check(t, reporter)
		})
	}
}

func TestJSONReporter_Report(t *testing.T) {
	tests := []struct {
		name         string
		result       *types.ScanResult
		wantContains []string
		wantErr      bool
	}{
		{
			name: "empty result",
			result: &types.ScanResult{
				Summary: types.Summary{
					Total:        0,
					FilesScanned: 0,
					Duration:     time.Millisecond * 50,
					ByType:       map[string]int{},
					ByPriority:   map[string]int{},
					ByAuthor:     map[string]int{},
				},
				TODOs:    []types.TODO{},
				Warnings: []types.Warning{},
			},
			wantContains: []string{
				`"total": 0`,
				`"todos": []`,
				`"warnings": []`,
			},
			wantErr: false,
		},
		{
			name: "single TODO",
			result: createTestScanResult(
				[]types.TODO{createTestTODO()},
				[]types.Warning{},
			),
			wantContains: []string{
				`"total": 1`,
				`"test.go"`,
				`"Test TODO message"`,
				`"medium"`,
			},
			wantErr: false,
		},
		{
			name: "multiple TODOs with different priorities",
			result: createTestScanResult(
				[]types.TODO{
					createTestTODO(func(t *types.TODO) { t.Priority = "high"; t.Type = "FIXME" }),
					createTestTODO(func(t *types.TODO) { t.Priority = "low"; t.Type = "HACK"; t.File = "another.go" }),
					createTestTODO(func(t *types.TODO) { t.Priority = "medium"; t.Author = "" }),
				},
				[]types.Warning{},
			),
			wantContains: []string{
				`"total": 3`,
				`"high"`,
				`"low"`,
				`"FIXME"`,
				`"HACK"`,
			},
			wantErr: false,
		},
		{
			name: "TODOs with warnings",
			result: createTestScanResult(
				[]types.TODO{createTestTODO()},
				[]types.Warning{
					{File: "warn.go", Line: 20, Message: "Warning message", Type: "parse_error"},
				},
			),
			wantContains: []string{
				`"warnings"`,
				`"warn.go"`,
				`"parse_error"`,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			reporter := NewJSONReporter(WithJSONWriter(&buf), WithIndent(true))

			err := reporter.Report(tt.result)

			if (err != nil) != tt.wantErr {
				t.Errorf("Report() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			output := buf.String()

			// Verify it's valid JSON
			var parsed map[string]interface{}
			if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
				t.Errorf("Output is not valid JSON: %v", err)
				return
			}

			// Check for expected content
			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("Output missing expected content: %s", want)
				}
			}
		})
	}
}

func TestJSONReporter_ReportStale(t *testing.T) {
	tests := []struct {
		name          string
		todos         []types.TODO
		thresholdDays int
		wantContains  []string
		wantErr       bool
	}{
		{
			name:          "empty stale list",
			todos:         []types.TODO{},
			thresholdDays: 90,
			wantContains: []string{
				`"count": 0`,
				`"threshold_days": 90`,
				`"todos": []`,
			},
			wantErr: false,
		},
		{
			name: "single stale TODO",
			todos: []types.TODO{
				createTestTODO(func(t *types.TODO) {
					t.Age = 100
					t.Message = "Old TODO"
				}),
			},
			thresholdDays: 90,
			wantContains: []string{
				`"count": 1`,
				`"Old TODO"`,
				`"threshold_days": 90`,
			},
			wantErr: false,
		},
		{
			name: "multiple stale TODOs",
			todos: []types.TODO{
				createTestTODO(func(t *types.TODO) { t.Age = 100; t.Author = "alice" }),
				createTestTODO(func(t *types.TODO) { t.Age = 200; t.Author = "bob" }),
				createTestTODO(func(t *types.TODO) { t.Age = 150; t.Author = "" }),
			},
			thresholdDays: 30,
			wantContains: []string{
				`"count": 3`,
				`"alice"`,
				`"bob"`,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			reporter := NewJSONReporter(WithJSONWriter(&buf))

			err := reporter.ReportStale(tt.todos, tt.thresholdDays)

			if (err != nil) != tt.wantErr {
				t.Errorf("ReportStale() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			output := buf.String()

			// Verify it's valid JSON
			var parsed map[string]interface{}
			if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
				t.Errorf("Output is not valid JSON: %v", err)
				return
			}

			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("Output missing expected content: %s", want)
				}
			}
		})
	}
}

func TestJSONReporter_ReportOrphaned(t *testing.T) {
	tests := []struct {
		name         string
		todos        []types.TODO
		inactiveDays int
		wantContains []string
		wantErr      bool
	}{
		{
			name:         "empty orphaned list",
			todos:        []types.TODO{},
			inactiveDays: 180,
			wantContains: []string{
				`"count": 0`,
				`"inactive_days": 180`,
			},
			wantErr: false,
		},
		{
			name: "single orphaned TODO",
			todos: []types.TODO{
				createTestTODO(func(t *types.TODO) {
					t.Author = "leftuser"
					t.Priority = "high"
				}),
			},
			inactiveDays: 180,
			wantContains: []string{
				`"count": 1`,
				`"leftuser"`,
				`"authors"`,
			},
			wantErr: false,
		},
		{
			name: "multiple orphaned TODOs grouped by author",
			todos: []types.TODO{
				createTestTODO(func(t *types.TODO) { t.Author = "alice"; t.Priority = "high" }),
				createTestTODO(func(t *types.TODO) { t.Author = "alice"; t.Priority = "medium" }),
				createTestTODO(func(t *types.TODO) { t.Author = "bob"; t.Priority = "low" }),
				createTestTODO(func(t *types.TODO) { t.Author = ""; t.Priority = "low" }),
			},
			inactiveDays: 90,
			wantContains: []string{
				`"count": 4`,
				`"alice"`,
				`"bob"`,
				`"未知"`,
				`"high": 1`,
				`"medium": 1`,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			reporter := NewJSONReporter(WithJSONWriter(&buf))

			err := reporter.ReportOrphaned(tt.todos, tt.inactiveDays)

			if (err != nil) != tt.wantErr {
				t.Errorf("ReportOrphaned() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			output := buf.String()

			// Verify it's valid JSON
			var parsed map[string]interface{}
			if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
				t.Errorf("Output is not valid JSON: %v", err)
				return
			}

			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("Output missing expected content: %s", want)
				}
			}
		})
	}
}

func TestJSONReporter_NoIndent(t *testing.T) {
	var buf bytes.Buffer
	reporter := NewJSONReporter(WithJSONWriter(&buf), WithIndent(false))

	result := createTestScanResult(
		[]types.TODO{createTestTODO()},
		[]types.Warning{},
	)

	err := reporter.Report(result)
	if err != nil {
		t.Fatalf("Report() error = %v", err)
	}

	output := buf.String()

	// Without indent, output should not contain leading spaces for indentation
	// (just newlines, no "  " indentation)
	if strings.Contains(output, "\n  ") {
		t.Error("Output should not contain indentation when indent is false")
	}
}

// =============================================================================
// Markdown Reporter Tests
// =============================================================================

func TestNewMarkdownReporter(t *testing.T) {
	reporter := NewMarkdownReporter()
	if reporter == nil {
		t.Fatal("NewMarkdownReporter returned nil")
	}
}

func TestMarkdownReporter_SetOutput(t *testing.T) {
	reporter := NewMarkdownReporter()
	var buf bytes.Buffer
	reporter.SetOutput(&buf)

	if reporter.output == nil {
		t.Error("output should not be nil after SetOutput")
	}
}

func TestMarkdownReporter_Report(t *testing.T) {
	tests := []struct {
		name         string
		result       *types.ScanResult
		wantContains []string
		wantErr      bool
	}{
		{
			name: "empty result",
			result: &types.ScanResult{
				Summary: types.Summary{
					Total:        0,
					FilesScanned: 0,
					Duration:     time.Millisecond * 50,
				},
				TODOs:    []types.TODO{},
				Warnings: []types.Warning{},
			},
			wantContains: []string{
				"# TODO 报告",
				"暂无 TODO",
				"- **总计**: 0 个 TODO",
			},
			wantErr: false,
		},
		{
			name: "single TODO",
			result: createTestScanResult(
				[]types.TODO{createTestTODO(func(t *types.TODO) { t.Priority = "high" })},
				[]types.Warning{},
			),
			wantContains: []string{
				"# TODO 报告",
				"- **总计**: 1 个 TODO",
				"test.go",
				"Test TODO message",
				":red_circle:", // high priority emoji
			},
			wantErr: false,
		},
		{
			name: "multiple TODOs with statistics",
			result: createTestScanResult(
				[]types.TODO{
					createTestTODO(func(t *types.TODO) { t.Priority = "high"; t.Type = "FIXME" }),
					createTestTODO(func(t *types.TODO) { t.Priority = "medium"; t.Type = "TODO" }),
					createTestTODO(func(t *types.TODO) { t.Priority = "low"; t.Type = "HACK" }),
				},
				[]types.Warning{},
			),
			wantContains: []string{
				"- **总计**: 3 个 TODO",
				"### 按类型统计",
				"### 按优先级统计",
				"| 类型 | 数量 |",
				"| 优先级 | 数量 |",
				"FIXME",
				"HACK",
			},
			wantErr: false,
		},
		{
			name: "TODOs with warnings",
			result: createTestScanResult(
				[]types.TODO{createTestTODO()},
				[]types.Warning{
					{File: "warn.go", Line: 20, Message: "Parse error", Type: "error"},
				},
			),
			wantContains: []string{
				"## 警告",
				"warn.go:20",
				"Parse error",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			reporter := NewMarkdownReporter()
			reporter.SetOutput(&buf)

			err := reporter.Report(tt.result)

			if (err != nil) != tt.wantErr {
				t.Errorf("Report() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			output := buf.String()

			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("Output missing expected content: %q\nGot output:\n%s", want, output)
				}
			}
		})
	}
}

func TestMarkdownReporter_ReportStale(t *testing.T) {
	tests := []struct {
		name         string
		todos        []types.TODO
		threshold    int
		wantContains []string
		wantErr      bool
	}{
		{
			name:      "empty stale list",
			todos:     []types.TODO{},
			threshold: 90,
			wantContains: []string{
				"# 过期 TODO 报告",
				"未发现过期 TODO",
				"过期阈值: 90 天",
			},
			wantErr: false,
		},
		{
			name: "stale TODOs",
			todos: []types.TODO{
				createTestTODO(func(t *types.TODO) {
					t.Age = 100
					t.Message = "Old TODO to fix"
					t.Author = "olddev"
				}),
				createTestTODO(func(t *types.TODO) {
					t.Age = 200
					t.Message = "Very old TODO"
					t.Author = "ancientdev"
				}),
			},
			threshold: 30,
			wantContains: []string{
				"发现 2 个过期 TODO",
				"| 年龄 | 文件 | 修改次数 | 作者 | 描述 |",
				"olddev",
				"Old TODO to fix",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			reporter := NewMarkdownReporter()
			reporter.SetOutput(&buf)

			err := reporter.ReportStale(tt.todos, tt.threshold)

			if (err != nil) != tt.wantErr {
				t.Errorf("ReportStale() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			output := buf.String()

			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("Output missing expected content: %q", want)
				}
			}
		})
	}
}

func TestMarkdownReporter_ReportOrphaned(t *testing.T) {
	tests := []struct {
		name         string
		todos        []types.TODO
		wantContains []string
		wantErr      bool
	}{
		{
			name:  "empty orphaned list",
			todos: []types.TODO{},
			wantContains: []string{
				"# 孤儿 TODO 报告",
				"未发现孤儿 TODO",
			},
			wantErr: false,
		},
		{
			name: "orphaned TODOs grouped by author",
			todos: []types.TODO{
				createTestTODO(func(t *types.TODO) { t.Author = "alice"; t.Priority = "high" }),
				createTestTODO(func(t *types.TODO) { t.Author = "alice"; t.Priority = "medium" }),
				createTestTODO(func(t *types.TODO) { t.Author = "bob"; t.Priority = "low" }),
			},
			wantContains: []string{
				"发现 3 个孤儿 TODO",
				"## @alice (2 个)",
				"## @bob (1 个)",
			},
			wantErr: false,
		},
		{
			name: "orphaned TODOs with unknown author",
			todos: []types.TODO{
				createTestTODO(func(t *types.TODO) { t.Author = "" }),
			},
			wantContains: []string{
				"发现 1 个孤儿 TODO",
				"## @未知",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			reporter := NewMarkdownReporter()
			reporter.SetOutput(&buf)

			err := reporter.ReportOrphaned(tt.todos)

			if (err != nil) != tt.wantErr {
				t.Errorf("ReportOrphaned() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			output := buf.String()

			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("Output missing expected content: %q", want)
				}
			}
		})
	}
}

func TestGetPriorityEmoji(t *testing.T) {
	tests := []struct {
		priority string
		want     string
	}{
		{"high", ":red_circle: 高"},
		{"HIGH", ":red_circle: 高"},
		{"High", ":red_circle: 高"},
		{"medium", ":yellow_circle: 中"},
		{"MEDIUM", ":yellow_circle: 中"},
		{"low", ":green_circle: 低"},
		{"LOW", ":green_circle: 低"},
		{"unknown", "unknown"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.priority, func(t *testing.T) {
			got := getPriorityEmoji(tt.priority)
			if got != tt.want {
				t.Errorf("getPriorityEmoji(%q) = %q, want %q", tt.priority, got, tt.want)
			}
		})
	}
}

func TestEscapeMarkdown(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "no special characters",
			input: "Simple message",
			want:  "Simple message",
		},
		{
			name:  "pipe character",
			input: "Message with | pipe",
			want:  "Message with \\| pipe",
		},
		{
			name:  "multiple pipes",
			input: "a | b | c",
			want:  "a \\| b \\| c",
		},
		{
			name:  "newline replaced with space",
			input: "Line1\nLine2",
			want:  "Line1 Line2",
		},
		{
			name:  "combined special chars",
			input: "A | B\nC",
			want:  "A \\| B C",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := escapeMarkdown(tt.input)
			if got != tt.want {
				t.Errorf("escapeMarkdown(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestFormatAge(t *testing.T) {
	tests := []struct {
		name string
		days int
		want string
	}{
		{
			name: "less than 30 days",
			days: 15,
			want: "15 天",
		},
		{
			name: "exactly 30 days",
			days: 30,
			want: "1.0 个月",
		},
		{
			name: "60 days",
			days: 60,
			want: "2.0 个月",
		},
		{
			name: "less than 365 days",
			days: 180,
			want: "6.0 个月",
		},
		{
			name: "exactly 365 days",
			days: 365,
			want: "1.0 年",
		},
		{
			name: "more than 365 days",
			days: 500,
			want: "1.4 年",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatAge(tt.days)
			if got != tt.want {
				t.Errorf("formatAge(%d) = %q, want %q", tt.days, got, tt.want)
			}
		})
	}
}

// =============================================================================
// Text Reporter Tests
// =============================================================================

func TestNewTextReporter(t *testing.T) {
	tests := []struct {
		name  string
		opts  []Option
		check func(*testing.T, *TextReporter)
	}{
		{
			name: "default options",
			opts: nil,
			check: func(t *testing.T, r *TextReporter) {
				if r.truncate != 80 {
					t.Errorf("default truncate should be 80, got %d", r.truncate)
				}
				if r.showColors != true {
					t.Error("default showColors should be true")
				}
			},
		},
		{
			name: "with custom truncate",
			opts: []Option{WithTruncate(50)},
			check: func(t *testing.T, r *TextReporter) {
				if r.truncate != 50 {
					t.Errorf("truncate should be 50, got %d", r.truncate)
				}
			},
		},
		{
			name: "with colors disabled",
			opts: []Option{WithColors(false)},
			check: func(t *testing.T, r *TextReporter) {
				if r.showColors != false {
					t.Error("showColors should be false")
				}
			},
		},
		{
			name: "with custom writer",
			opts: []Option{WithWriter(&bytes.Buffer{})},
			check: func(t *testing.T, r *TextReporter) {
				if r.writer == nil {
					t.Error("writer should not be nil")
				}
			},
		},
		{
			name: "multiple options",
			opts: []Option{
				WithTruncate(100),
				WithColors(false),
				WithWriter(&bytes.Buffer{}),
			},
			check: func(t *testing.T, r *TextReporter) {
				if r.truncate != 100 {
					t.Errorf("truncate should be 100, got %d", r.truncate)
				}
				if r.showColors != false {
					t.Error("showColors should be false")
				}
				if r.writer == nil {
					t.Error("writer should not be nil")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reporter := NewTextReporter(tt.opts...)
			if reporter == nil {
				t.Fatal("NewTextReporter returned nil")
			}
			tt.check(t, reporter)
		})
	}
}

func TestTextReporter_Report(t *testing.T) {
	tests := []struct {
		name         string
		result       *types.ScanResult
		wantContains []string
		wantErr      bool
	}{
		{
			name: "empty result",
			result: &types.ScanResult{
				Summary: types.Summary{
					Total:        0,
					FilesScanned: 5,
					Duration:     time.Millisecond * 100,
				},
				TODOs:    []types.TODO{},
				Warnings: []types.Warning{},
			},
			wantContains: []string{
				"TODO 扫描结果",
				"未发现 TODO",
				"扫描了 5 个文件",
			},
			wantErr: false,
		},
		{
			name: "single TODO",
			result: createTestScanResult(
				[]types.TODO{createTestTODO()},
				[]types.Warning{},
			),
			wantContains: []string{
				"TODO 扫描结果",
				"test.go",
				"Test TODO message",
				"扫描了 10 个文件",
				"发现 1 个 TODO",
			},
			wantErr: false,
		},
		{
			name: "multiple TODOs with warnings",
			result: createTestScanResult(
				[]types.TODO{
					createTestTODO(func(t *types.TODO) { t.Priority = "high" }),
					createTestTODO(func(t *types.TODO) { t.Priority = "low"; t.File = "other.go" }),
				},
				[]types.Warning{
					{File: "warn.go", Line: 5, Message: "Test warning", Type: "error"},
				},
			),
			wantContains: []string{
				"发现 2 个 TODO",
				"警告",
				"warn.go:5",
				"Test warning",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			reporter := NewTextReporter(WithWriter(&buf))

			err := reporter.Report(tt.result)

			if (err != nil) != tt.wantErr {
				t.Errorf("Report() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			output := buf.String()

			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("Output missing expected content: %q\nGot output:\n%s", want, output)
				}
			}
		})
	}
}

func TestTextReporter_ReportStale(t *testing.T) {
	tests := []struct {
		name          string
		todos         []types.TODO
		thresholdDays int
		wantContains  []string
		wantErr       bool
	}{
		{
			name:          "empty stale list",
			todos:         []types.TODO{},
			thresholdDays: 90,
			wantContains: []string{
				"过期 TODO",
				"超过 90 天",
				"未发现过期 TODO",
			},
			wantErr: false,
		},
		{
			name: "stale TODOs",
			todos: []types.TODO{
				createTestTODO(func(t *types.TODO) {
					t.Age = 100
					t.Author = "olddev"
					t.ChurnScore = 10
					t.Message = "Old TODO message"
				}),
			},
			thresholdDays: 30,
			wantContains: []string{
				"发现 1 个过期 TODO",
				"olddev",
				"Old TODO message",
			},
			wantErr: false,
		},
		{
			name: "stale TODO with unknown author",
			todos: []types.TODO{
				createTestTODO(func(t *types.TODO) {
					t.Author = ""
					t.ChurnScore = 5
				}),
			},
			thresholdDays: 30,
			wantContains: []string{
				"未知", // unknown author should show as "未知"
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			reporter := NewTextReporter(WithWriter(&buf))

			err := reporter.ReportStale(tt.todos, tt.thresholdDays)

			if (err != nil) != tt.wantErr {
				t.Errorf("ReportStale() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			output := buf.String()

			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("Output missing expected content: %q", want)
				}
			}
		})
	}
}

func TestTextReporter_ReportOrphaned(t *testing.T) {
	tests := []struct {
		name         string
		todos        []types.TODO
		inactiveDays int
		wantContains []string
		wantErr      bool
	}{
		{
			name:         "empty orphaned list",
			todos:        []types.TODO{},
			inactiveDays: 180,
			wantContains: []string{
				"孤儿 TODO",
				"未发现孤儿 TODO",
			},
			wantErr: false,
		},
		{
			name: "orphaned TODOs grouped by author",
			todos: []types.TODO{
				createTestTODO(func(t *types.TODO) { t.Author = "alice"; t.Priority = "high" }),
				createTestTODO(func(t *types.TODO) { t.Author = "alice"; t.Priority = "medium" }),
				createTestTODO(func(t *types.TODO) { t.Author = "bob"; t.Priority = "low" }),
			},
			inactiveDays: 90,
			wantContains: []string{
				"发现 3 个孤儿 TODO",
				"alice",
				"bob",
				"1高, 1中",  // priority count for alice: 1 high, 1 medium
				"详细列表",
			},
			wantErr: false,
		},
		{
			name: "orphaned TODO with unknown author",
			todos: []types.TODO{
				createTestTODO(func(t *types.TODO) { t.Author = "" }),
			},
			inactiveDays: 180,
			wantContains: []string{
				"未知",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			reporter := NewTextReporter(WithWriter(&buf))

			err := reporter.ReportOrphaned(tt.todos, tt.inactiveDays)

			if (err != nil) != tt.wantErr {
				t.Errorf("ReportOrphaned() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			output := buf.String()

			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("Output missing expected content: %q", want)
				}
			}
		})
	}
}

func TestTextReporter_TruncateString(t *testing.T) {
	reporter := &TextReporter{truncate: 80}

	tests := []struct {
		name   string
		input  string
		maxLen int
		want   string
	}{
		{
			name:   "short string unchanged",
			input:  "short",
			maxLen: 10,
			want:   "short",
		},
		{
			name:   "exact length unchanged",
			input:  "12345",
			maxLen: 5,
			want:   "12345",
		},
		{
			name:   "long string truncated with ellipsis",
			input:  "this is a very long string that needs truncation",
			maxLen: 20,
			want:   "this is a very lo...",
		},
		{
			name:   "maxLen 3",
			input:  "abcd",
			maxLen: 3,
			want:   "abc",
		},
		{
			name:   "maxLen 0",
			input:  "test",
			maxLen: 0,
			want:   "",
		},
		{
			name:   "maxLen 1",
			input:  "test",
			maxLen: 1,
			want:   "t",
		},
		{
			name:   "maxLen 2",
			input:  "test",
			maxLen: 2,
			want:   "te",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := reporter.truncateString(tt.input, tt.maxLen)
			if got != tt.want {
				t.Errorf("truncateString(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
			}
		})
	}
}

func TestTextReporter_FormatPriority(t *testing.T) {
	reporter := &TextReporter{}

	tests := []struct {
		priority string
		want     string
	}{
		{"high", "HIGH"},
		{"HIGH", "HIGH"},
		{"High", "HIGH"},
		{"medium", "MED"},
		{"MEDIUM", "MED"},
		{"Medium", "MED"},
		{"low", "LOW"},
		{"LOW", "LOW"},
		{"unknown", "LOW"},
		{"", "LOW"},
	}

	for _, tt := range tests {
		t.Run(tt.priority, func(t *testing.T) {
			got := reporter.formatPriority(tt.priority)
			if got != tt.want {
				t.Errorf("formatPriority(%q) = %q, want %q", tt.priority, got, tt.want)
			}
		})
	}
}

func TestTextReporter_FormatType(t *testing.T) {
	reporter := &TextReporter{}

	tests := []struct {
		todoType string
		want     string
	}{
		{"todo", "TODO"},
		{"TODO", "TODO"},
		{"fixme", "FIXME"},
		{"FIXME", "FIXME"},
		{"hack", "HACK"},
		{"HACK", "HACK"},
		{"bug", "BUG"},
		{"mixedcase", "MIXEDCASE"},
	}

	for _, tt := range tests {
		t.Run(tt.todoType, func(t *testing.T) {
			got := reporter.formatType(tt.todoType)
			if got != tt.want {
				t.Errorf("formatType(%q) = %q, want %q", tt.todoType, got, tt.want)
			}
		})
	}
}

func TestCountByPriority(t *testing.T) {
	tests := []struct {
		name      string
		todos     []types.TODO
		wantHigh  int
		wantMed   int
		wantLow   int
	}{
		{
			name:      "empty list",
			todos:     []types.TODO{},
			wantHigh:  0,
			wantMed:   0,
			wantLow:   0,
		},
		{
			name: "single high",
			todos: []types.TODO{
				createTestTODO(func(t *types.TODO) { t.Priority = "high" }),
			},
			wantHigh: 1,
			wantMed:  0,
			wantLow:  0,
		},
		{
			name: "mixed priorities",
			todos: []types.TODO{
				createTestTODO(func(t *types.TODO) { t.Priority = "high" }),
				createTestTODO(func(t *types.TODO) { t.Priority = "high" }),
				createTestTODO(func(t *types.TODO) { t.Priority = "medium" }),
				createTestTODO(func(t *types.TODO) { t.Priority = "low" }),
				createTestTODO(func(t *types.TODO) { t.Priority = "low" }),
				createTestTODO(func(t *types.TODO) { t.Priority = "unknown" }),
			},
			wantHigh: 2,
			wantMed:  1,
			wantLow:  3, // "unknown" and "low" both count as low
		},
		{
			name: "case insensitive",
			todos: []types.TODO{
				createTestTODO(func(t *types.TODO) { t.Priority = "HIGH" }),
				createTestTODO(func(t *types.TODO) { t.Priority = "Medium" }),
				createTestTODO(func(t *types.TODO) { t.Priority = "LOW" }),
			},
			wantHigh: 1,
			wantMed:  1,
			wantLow:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			high, med, low := countByPriority(tt.todos)
			if high != tt.wantHigh {
				t.Errorf("countByPriority() high = %d, want %d", high, tt.wantHigh)
			}
			if med != tt.wantMed {
				t.Errorf("countByPriority() medium = %d, want %d", med, tt.wantMed)
			}
			if low != tt.wantLow {
				t.Errorf("countByPriority() low = %d, want %d", low, tt.wantLow)
			}
		})
	}
}

// =============================================================================
// Integration Tests
// =============================================================================

func TestAllReporters_ConsistentOutput(t *testing.T) {
	// Create test data
	todos := []types.TODO{
		createTestTODO(func(t *types.TODO) {
			t.Priority = "high"
			t.Type = "FIXME"
			t.Message = "Fix this critical bug"
		}),
		createTestTODO(func(t *types.TODO) {
			t.Priority = "medium"
			t.Type = "TODO"
			t.File = "main.go"
			t.Line = 50
			t.Message = "Implement feature"
		}),
	}
	result := createTestScanResult(todos, []types.Warning{})

	// Test JSON Reporter
	var jsonBuf bytes.Buffer
	jsonReporter := NewJSONReporter(WithJSONWriter(&jsonBuf))
	if err := jsonReporter.Report(result); err != nil {
		t.Errorf("JSON Reporter failed: %v", err)
	}

	// Test Markdown Reporter
	var mdBuf bytes.Buffer
	mdReporter := NewMarkdownReporter()
	mdReporter.SetOutput(&mdBuf)
	if err := mdReporter.Report(result); err != nil {
		t.Errorf("Markdown Reporter failed: %v", err)
	}

	// Test Text Reporter
	var textBuf bytes.Buffer
	textReporter := NewTextReporter(WithWriter(&textBuf))
	if err := textReporter.Report(result); err != nil {
		t.Errorf("Text Reporter failed: %v", err)
	}

	// All should complete without error and produce output
	if jsonBuf.Len() == 0 {
		t.Error("JSON Reporter produced empty output")
	}
	if mdBuf.Len() == 0 {
		t.Error("Markdown Reporter produced empty output")
	}
	if textBuf.Len() == 0 {
		t.Error("Text Reporter produced empty output")
	}

	// Verify key content appears in all outputs
	jsonOutput := jsonBuf.String()
	mdOutput := mdBuf.String()
	textOutput := textBuf.String()

	// All should mention the TODO count
	if !strings.Contains(jsonOutput, `"total": 2`) {
		t.Error("JSON output missing total count")
	}
	if !strings.Contains(mdOutput, "2 个 TODO") {
		t.Error("Markdown output missing total count")
	}
	if !strings.Contains(textOutput, "2 个 TODO") {
		t.Error("Text output missing total count")
	}
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkJSONReporter_Report(b *testing.B) {
	todos := make([]types.TODO, 100)
	for i := 0; i < 100; i++ {
		todos[i] = createTestTODO(func(t *types.TODO) {
			t.File = "file.go"
			t.Line = i
			t.Message = "Benchmark TODO"
		})
	}
	result := createTestScanResult(todos, []types.Warning{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		reporter := NewJSONReporter(WithJSONWriter(&buf))
		reporter.Report(result)
	}
}

func BenchmarkMarkdownReporter_Report(b *testing.B) {
	todos := make([]types.TODO, 100)
	for i := 0; i < 100; i++ {
		todos[i] = createTestTODO(func(t *types.TODO) {
			t.File = "file.go"
			t.Line = i
			t.Message = "Benchmark TODO"
		})
	}
	result := createTestScanResult(todos, []types.Warning{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		reporter := NewMarkdownReporter()
		reporter.SetOutput(&buf)
		reporter.Report(result)
	}
}

func BenchmarkTextReporter_Report(b *testing.B) {
	todos := make([]types.TODO, 100)
	for i := 0; i < 100; i++ {
		todos[i] = createTestTODO(func(t *types.TODO) {
			t.File = "file.go"
			t.Line = i
			t.Message = "Benchmark TODO"
		})
	}
	result := createTestScanResult(todos, []types.Warning{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		reporter := NewTextReporter(WithWriter(&buf))
		reporter.Report(result)
	}
}

func BenchmarkCountByPriority(b *testing.B) {
	todos := make([]types.TODO, 1000)
	for i := 0; i < 1000; i++ {
		priority := "medium"
		if i%3 == 0 {
			priority = "high"
		} else if i%3 == 1 {
			priority = "low"
		}
		todos[i] = createTestTODO(func(t *types.TODO) {
			t.Priority = priority
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		countByPriority(todos)
	}
}
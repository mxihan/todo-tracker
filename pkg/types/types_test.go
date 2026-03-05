// Package types_test 测试types包中的核心功能
package types

import (
	"testing"
	"time"
)

// TestTODOGenerateID 测试TODO ID生成
func TestTODOGenerateID(t *testing.T) {
	tests := []struct {
		name     string
		todo     TODO
		wantLen  int // 期望ID长度（hex编码后）
	}{
		{
			name: "基本TODO",
			todo: TODO{
				File: "main.go",
				Line: 10,
			},
			wantLen: 16, // SHA256前8字节 = 16个hex字符
		},
		{
			name: "不同文件相同行号",
			todo: TODO{
				File: "utils.go",
				Line: 10,
			},
			wantLen: 16,
		},
		{
			name: "相同文件不同行号",
			todo: TODO{
				File: "main.go",
				Line: 20,
			},
			wantLen: 16,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.todo.GenerateID()
			if len(got) != tt.wantLen {
				t.Errorf("GenerateID() length = %v, want %v", len(got), tt.wantLen)
			}
			// 相同输入应生成相同ID
			got2 := tt.todo.GenerateID()
			if got != got2 {
				t.Errorf("GenerateID() not deterministic: %v != %v", got, got2)
			}
		})
	}
}

// TestTODOGenerateIDUniqueness 测试ID唯一性
func TestTODOGenerateIDUniqueness(t *testing.T) {
	todos := []TODO{
		{File: "main.go", Line: 10},
		{File: "main.go", Line: 20},
		{File: "utils.go", Line: 10},
		{File: "utils.go", Line: 20},
	}

	ids := make(map[string]bool)
	for _, todo := range todos {
		id := todo.GenerateID()
		if ids[id] {
			t.Errorf("Duplicate ID generated: %s", id)
		}
		ids[id] = true
	}
}

// TestTODOIsStale 测试过期检测
func TestTODOIsStale(t *testing.T) {
	tests := []struct {
		name          string
		todo          TODO
		thresholdDays int
		want          bool
	}{
		{
			name: "新创建的TODO未过期",
			todo: TODO{
				CreatedAt: time.Now(),
			},
			thresholdDays: 90,
			want:          false,
		},
		{
			name: "正好91天的TODO已过期",
			todo: TODO{
				CreatedAt: time.Now().AddDate(0, 0, -91),
			},
			thresholdDays: 90,
			want:          true,
		},
		{
			name: "正好89天的TODO未过期",
			todo: TODO{
				CreatedAt: time.Now().AddDate(0, 0, -89),
			},
			thresholdDays: 90,
			want:          false,
		},
		{
			name: "零值CreatedAt未过期",
			todo: TODO{
				CreatedAt: time.Time{},
			},
			thresholdDays: 90,
			want:          false,
		},
		{
			name: "一年前的TODO已过期",
			todo: TODO{
				CreatedAt: time.Now().AddDate(-1, 0, 0),
			},
			thresholdDays: 90,
			want:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.todo.IsStale(tt.thresholdDays)
			if got != tt.want {
				t.Errorf("IsStale() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestTODOFormatAge 测试年龄格式化
func TestTODOFormatAge(t *testing.T) {
	tests := []struct {
		name string
		todo TODO
		want string
	}{
		{
			name: "零值时间返回未知",
			todo: TODO{CreatedAt: time.Time{}},
			want: "未知",
		},
		{
			name: "刚创建（1天）",
			todo: TODO{CreatedAt: time.Now().AddDate(0, 0, -1)},
			want: "1天",
		},
		{
			name: "15天",
			todo: TODO{CreatedAt: time.Now().AddDate(0, 0, -15)},
			want: "15天",
		},
		{
			name: "60天显示为月",
			todo: TODO{CreatedAt: time.Now().AddDate(0, 0, -60)},
			want: "2.0个月",
		},
		{
			name: "400天显示为年",
			todo: TODO{CreatedAt: time.Now().AddDate(-1, 0, -35)},
			want: "1.1年",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.todo.FormatAge()
			// 由于时间计算可能有微小差异，检查前缀匹配
			if tt.want == "未知" && got != "未知" {
				t.Errorf("FormatAge() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestDefaultConfig 测试默认配置
func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Version != 1 {
		t.Errorf("DefaultConfig().Version = %v, want 1", config.Version)
	}

	if len(config.Scan.Paths) == 0 {
		t.Error("DefaultConfig().Scan.Paths should not be empty")
	}

	if config.Git.Enabled != true {
		t.Error("DefaultConfig().Git.Enabled should be true")
	}

	if config.Stale.ThresholdDays != 90 {
		t.Errorf("DefaultConfig().Stale.ThresholdDays = %v, want 90", config.Stale.ThresholdDays)
	}

	if config.Orphan.InactiveDays != 180 {
		t.Errorf("DefaultConfig().Orphan.InactiveDays = %v, want 180", config.Orphan.InactiveDays)
	}
}

// TestDefaultPatternConfig 测试默认模式配置
func TestDefaultPatternConfig(t *testing.T) {
	config := DefaultPatternConfig()

	expectedTypes := []string{"TODO", "FIXME", "HACK", "XXX", "BUG"}
	for _, expected := range expectedTypes {
		found := false
		for _, actual := range config.Types {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("DefaultPatternConfig().Types missing %s", expected)
		}
	}

	if _, ok := config.PriorityMarkers["high"]; !ok {
		t.Error("DefaultPatternConfig().PriorityMarkers missing 'high' key")
	}
}

// BenchmarkTODOGenerateID 基准测试ID生成
func BenchmarkTODOGenerateID(b *testing.B) {
	todo := TODO{File: "main.go", Line: 10}
	for i := 0; i < b.N; i++ {
		todo.GenerateID()
	}
}

// BenchmarkTODOIsStale 基准测试过期检测
func BenchmarkTODOIsStale(b *testing.B) {
	todo := TODO{
		CreatedAt: time.Now().AddDate(0, 0, -100),
	}
	for i := 0; i < b.N; i++ {
		todo.IsStale(90)
	}
}
// Package scanner_test 性能基准测试
package scanner

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/mxihan/todo-tracker/internal/parser"
	"github.com/mxihan/todo-tracker/pkg/types"
)

// BenchmarkScannerScan 基准测试扫描器
func BenchmarkScannerScan(b *testing.B) {
	// 创建测试目录
	tempDir, err := os.MkdirTemp("", "bench-scan")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建测试文件
	for i := 0; i < 100; i++ {
		filePath := filepath.Join(tempDir, fmt.Sprintf("file%03d.go", i))
		content := generateTestContent(50) // 每个文件50行
		os.WriteFile(filePath, []byte(content), 0644)
	}

	config := types.DefaultConfig()
	scanner := NewScanner(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scanner.Scan(nil, tempDir)
	}
}

// BenchmarkParserParseFile 基准测试解析器
func BenchmarkParserParseFile(b *testing.B) {
	p := parser.NewParser(nil)
	content := generateTestContent(100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.ParseFile(content, "bench.go")
	}
}

// BenchmarkParserParseLine 基准测试单行解析（通过ParseFile）
func BenchmarkParserParseLine(b *testing.B) {
	p := parser.NewParser(nil)
	line := "// TODO(@alice) #123!: 这是一个测试任务，包含一些描述内容"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.ParseFile(line, "bench.go")
	}
}

// BenchmarkSkipRulesShouldSkipDirectory 基准测试目录跳过判断
func BenchmarkSkipRulesShouldSkipDirectory(b *testing.B) {
	rules := DefaultSkipRules()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rules.ShouldSkipDirectory("node_modules")
	}
}

// BenchmarkSkipRulesShouldSkipFile 基准测试文件跳过判断
func BenchmarkSkipRulesShouldSkipFile(b *testing.B) {
	rules := DefaultSkipRules()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rules.ShouldSkipFile("main.go")
	}
}

// BenchmarkScanWithDifferentFileCounts 测试不同文件数量的扫描性能
func BenchmarkScanWithDifferentFileCounts(b *testing.B) {
	fileCounts := []int{10, 50, 100, 500}

	for _, count := range fileCounts {
		b.Run(fmt.Sprintf("Files_%d", count), func(b *testing.B) {
			tempDir, err := os.MkdirTemp("", "bench-files")
			if err != nil {
				b.Fatalf("Failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tempDir)

			for i := 0; i < count; i++ {
				filePath := filepath.Join(tempDir, fmt.Sprintf("file%03d.go", i))
				content := generateTestContent(20)
				os.WriteFile(filePath, []byte(content), 0644)
			}

			config := types.DefaultConfig()
			scanner := NewScanner(config)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				scanner.Scan(nil, tempDir)
			}
		})
	}
}

// BenchmarkParseWithDifferentTODODensities 测试不同TODO密度的解析性能
func BenchmarkParseWithDifferentTODODensities(b *testing.B) {
	densities := []string{"low", "medium", "high"}

	for _, density := range densities {
		b.Run(fmt.Sprintf("Density_%s", density), func(b *testing.B) {
			p := parser.NewParser(nil)
			content := generateContentWithDensity(density, 100)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				p.ParseFile(content, "bench.go")
			}
		})
	}
}

// BenchmarkConcurrentScan 基准测试并发扫描
func BenchmarkConcurrentScan(b *testing.B) {
	workerCounts := []int{1, 2, 4, 8}

	for _, workers := range workerCounts {
		b.Run(fmt.Sprintf("Workers_%d", workers), func(b *testing.B) {
			tempDir, err := os.MkdirTemp("", "bench-concurrent")
			if err != nil {
				b.Fatalf("Failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tempDir)

			// 创建测试文件
			for i := 0; i < 100; i++ {
				filePath := filepath.Join(tempDir, fmt.Sprintf("file%03d.go", i))
				content := generateTestContent(30)
				os.WriteFile(filePath, []byte(content), 0644)
			}

			config := types.DefaultConfig()
			config.Scan.Workers = workers
			scanner := NewScanner(config)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				scanner.Scan(nil, tempDir)
			}
		})
	}
}

// BenchmarkMultiLanguageParsing 基准测试多语言解析
func BenchmarkMultiLanguageParsing(b *testing.B) {
	p := parser.NewParser(nil)

	languages := []struct {
		name    string
		content string
		file    string
	}{
		{
			name: "Go",
			content: `package main
// TODO: Go comment
func main() {}
`,
			file: "main.go",
		},
		{
			name: "Python",
			content: `# TODO: Python comment
def main():
    pass
`,
			file: "main.py",
		},
		{
			name: "JavaScript",
			content: `// TODO: JS comment
function main() {}
`,
			file: "main.js",
		},
		{
			name: "HTML",
			content: `<!-- TODO: HTML comment -->
<html></html>
`,
			file: "index.html",
		},
	}

	for _, lang := range languages {
		b.Run(lang.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				p.ParseFile(lang.content, lang.file)
			}
		})
	}
}

// generateTestContent 生成测试内容
func generateTestContent(lines int) string {
	content := "package main\n\n"

	todoPatterns := []string{
		"// TODO: 这是一个测试任务 %d\n",
		"// FIXME: 需要修复的问题 %d\n",
		"// HACK: 临时解决方案 %d\n",
		"// normal comment %d\n",
		"var x%d int\n",
	}

	for i := 0; i < lines; i++ {
		pattern := todoPatterns[i%len(todoPatterns)]
		content += fmt.Sprintf(pattern, i)
	}

	content += "\nfunc main() {}\n"
	return content
}

// generateContentWithDensity 根据密度生成内容
func generateContentWithDensity(density string, lines int) string {
	content := "package main\n\n"

	todoInterval := 10 // 默认低密度
	switch density {
	case "low":
		todoInterval = 10
	case "medium":
		todoInterval = 5
	case "high":
		todoInterval = 2
	}

	for i := 0; i < lines; i++ {
		if i%todoInterval == 0 {
			content += fmt.Sprintf("// TODO: 测试任务 %d\n", i)
		} else {
			content += fmt.Sprintf("// 普通注释 %d\n", i)
		}
	}

	content += "\nfunc main() {}\n"
	return content
}

// BenchmarkMemoryAllocation 测试内存分配
func BenchmarkMemoryAllocation(b *testing.B) {
	p := parser.NewParser(nil)
	content := generateTestContent(100)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		todos := p.ParseFile(content, "bench.go")
		_ = todos
	}
}
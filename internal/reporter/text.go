// Package reporter 提供报告生成功能
package reporter

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/todo-tracker/todo-tracker/pkg/types"
)

// TextReporter 文本报告生成器
type TextReporter struct {
	writer     io.Writer
	truncate   int
	showColors bool
}

// NewTextReporter 创建新的文本报告生成器
func NewTextReporter(opts ...Option) *TextReporter {
	r := &TextReporter{
		writer:     os.Stdout,
		truncate:   80,
		showColors: true,
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

// Option 报告选项
type Option func(*TextReporter)

// WithWriter 设置输出写入器
func WithWriter(w io.Writer) Option {
	return func(r *TextReporter) {
		r.writer = w
	}
}

// WithTruncate 设置截断长度
func WithTruncate(length int) Option {
	return func(r *TextReporter) {
		r.truncate = length
	}
}

// WithColors 设置是否显示颜色
func WithColors(show bool) Option {
	return func(r *TextReporter) {
		r.showColors = show
	}
}

// Report 生成报告
func (r *TextReporter) Report(result *types.ScanResult) error {
	r.printHeader()
	r.printSummary(&result.Summary)

	if len(result.TODOs) == 0 {
		fmt.Fprintln(r.writer, "  未发现 TODO")
	} else {
		r.printTODOs(result.TODOs)
	}

	if len(result.Warnings) > 0 {
		r.printWarnings(result.Warnings)
	}

	r.printFooter(&result.Summary)
	return nil
}

// ReportStale 生成过期TODO报告
func (r *TextReporter) ReportStale(todos []types.TODO, thresholdDays int) error {
	fmt.Fprintf(r.writer, "\n 过期 TODO（超过 %d 天）\n", thresholdDays)
	fmt.Fprintln(r.writer, "════════════════════════════════════════════════════════════")
	fmt.Fprintln(r.writer)

	if len(todos) == 0 {
		fmt.Fprintln(r.writer, " 未发现过期 TODO")
	} else {
		fmt.Fprintln(r.writer, " 年龄      文件                      修改次数   作者         描述")
		fmt.Fprintln(r.writer, "════════════════════════════════════════════════════════════")

		for _, todo := range todos {
			age := todo.FormatAge()
			file := r.truncateString(fmt.Sprintf("%s:%d", todo.File, todo.Line), 30)
			author := r.truncateString(todo.Author, 12)
			if author == "" {
				author = "未知"
			}
			message := r.truncateString(todo.Message, 25)

			fmt.Fprintf(r.writer, " %-8s %-30s %-8d  %-12s %s\n",
				age, file, todo.ChurnScore, author, message)
		}
	}

	fmt.Fprintln(r.writer)
	if len(todos) > 0 {
		fmt.Fprintf(r.writer, " ⚠ 发现 %d 个过期 TODO，可能已过时或需要紧急处理\n", len(todos))
	}

	return nil
}

// ReportOrphaned 生成孤儿TODO报告
func (r *TextReporter) ReportOrphaned(todos []types.TODO, inactiveDays int) error {
	fmt.Fprintf(r.writer, "\n 孤儿 TODO（作者不活跃超过 %d 天）\n", inactiveDays)
	fmt.Fprintln(r.writer, "════════════════════════════════════════════════════════════")
	fmt.Fprintln(r.writer)

	if len(todos) == 0 {
		fmt.Fprintln(r.writer, " 未发现孤儿 TODO")
	} else {
		// 按作者分组
		byAuthor := make(map[string][]types.TODO)
		for _, todo := range todos {
			author := todo.Author
			if author == "" {
				author = "未知"
			}
			byAuthor[author] = append(byAuthor[author], todo)
		}

		fmt.Fprintln(r.writer, " 作者          TODO数量   优先级分布")
		fmt.Fprintln(r.writer, "════════════════════════════════════════════════════════════")

		for author, authorTodos := range byAuthor {
			high, medium, low := countByPriority(authorTodos)
			fmt.Fprintf(r.writer, " %-14s %-8d   %d高, %d中, %d低\n",
				r.truncateString(author, 14), len(authorTodos), high, medium, low)
		}

		fmt.Fprintln(r.writer)
		fmt.Fprintln(r.writer, " 详细列表：")
		for author, authorTodos := range byAuthor {
			fmt.Fprintf(r.writer, "   %s:\n", author)
			for _, todo := range authorTodos {
				file := r.truncateString(fmt.Sprintf("%s:%d", todo.File, todo.Line), 40)
				message := r.truncateString(todo.Message, 30)
				fmt.Fprintf(r.writer, "     [%s] %s - %s\n",
					strings.ToUpper(todo.Priority)[:1], file, message)
			}
		}
	}

	fmt.Fprintln(r.writer)
	if len(todos) > 0 {
		fmt.Fprintf(r.writer, " ⚠ 发现 %d 个孤儿 TODO，需要重新分配\n", len(todos))
	}

	return nil
}

// printHeader 打印报告头
func (r *TextReporter) printHeader() {
	fmt.Fprintln(r.writer)
	fmt.Fprintln(r.writer, " TODO 扫描结果")
	fmt.Fprintln(r.writer, "════════════════════════════════════════════════════════════")
	fmt.Fprintln(r.writer)
}

// printSummary 打印摘要
func (r *TextReporter) printSummary(summary *types.Summary) {
	fmt.Fprintf(r.writer, " 优先级   类型     位置                      描述\n")
	fmt.Fprintln(r.writer, "════════════════════════════════════════════════════════════")
}

// printTODOs 打印TODO列表
func (r *TextReporter) printTODOs(todos []types.TODO) {
	for _, todo := range todos {
		priority := r.formatPriority(todo.Priority)
		todoType := r.formatType(todo.Type)
		location := r.truncateString(fmt.Sprintf("%s:%d", todo.File, todo.Line), 25)
		message := r.truncateString(todo.Message, 30)

		fmt.Fprintf(r.writer, " %-8s %-8s %-25s %s\n", priority, todoType, location, message)
	}
	fmt.Fprintln(r.writer, "════════════════════════════════════════════════════════════")
}

// printWarnings 打印警告
func (r *TextReporter) printWarnings(warnings []types.Warning) {
	fmt.Fprintln(r.writer)
	fmt.Fprintln(r.writer, " 警告：")
	for _, w := range warnings {
		fmt.Fprintf(r.writer, "  [%s] %s:%d - %s\n", w.Type, w.File, w.Line, w.Message)
	}
}

// printFooter 打印报告尾
func (r *TextReporter) printFooter(summary *types.Summary) {
	fmt.Fprintln(r.writer)
	fmt.Fprintf(r.writer, " 扫描了 %d 个文件，发现 %d 个 TODO（耗时 %s）\n",
		summary.FilesScanned, summary.Total, summary.Duration.Round(time.Millisecond))

	fmt.Fprintln(r.writer)
	fmt.Fprintln(r.writer, " 运行 `todo stale` 查看过期 TODO")
	fmt.Fprintln(r.writer, " 运行 `todo report -f md -o TODO.md` 导出报告")
}

// formatPriority 格式化优先级
func (r *TextReporter) formatPriority(priority string) string {
	switch strings.ToLower(priority) {
	case "high":
		return "HIGH"
	case "medium":
		return "MED"
	default:
		return "LOW"
	}
}

// formatType 格式化TODO类型
func (r *TextReporter) formatType(todoType string) string {
	return strings.ToUpper(todoType)
}

// truncateString 截断字符串
func (r *TextReporter) truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// countByPriority 按优先级统计
func countByPriority(todos []types.TODO) (high, medium, low int) {
	for _, todo := range todos {
		switch strings.ToLower(todo.Priority) {
		case "high":
			high++
		case "medium":
			medium++
		default:
			low++
		}
	}
	return
}
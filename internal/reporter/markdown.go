// Package reporter 提供 Markdown 格式报告生成功能
package reporter

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/todo-tracker/todo-tracker/pkg/types"
)

// MarkdownReporter Markdown 格式报告生成器
type MarkdownReporter struct {
	output io.Writer
}

// NewMarkdownReporter 创建 Markdown 报告生成器
func NewMarkdownReporter() *MarkdownReporter {
	return &MarkdownReporter{
		output: os.Stdout,
	}
}

// SetOutput 设置输出目标
func (r *MarkdownReporter) SetOutput(w io.Writer) {
	r.output = w
}

// Report 生成 Markdown 报告
func (r *MarkdownReporter) Report(result *types.ScanResult) error {
	// 标题
	fmt.Fprintln(r.output, "# TODO 报告")
	fmt.Fprintln(r.output)

	// 元信息
	fmt.Fprintf(r.output, "> 生成时间: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Fprintln(r.output)

	// 概述
	fmt.Fprintln(r.output, "## 概述")
	fmt.Fprintln(r.output)
	fmt.Fprintf(r.output, "- **总计**: %d 个 TODO\n", result.Summary.Total)
	fmt.Fprintf(r.output, "- **扫描文件**: %d 个\n", result.Summary.FilesScanned)
	fmt.Fprintf(r.output, "- **扫描耗时**: %s\n", result.Summary.Duration)
	fmt.Fprintln(r.output)

	// 按类型统计
	if len(result.Summary.ByType) > 0 {
		fmt.Fprintln(r.output, "### 按类型统计")
		fmt.Fprintln(r.output)
		fmt.Fprintln(r.output, "| 类型 | 数量 |")
		fmt.Fprintln(r.output, "|------|------|")
		for t, count := range result.Summary.ByType {
			fmt.Fprintf(r.output, "| %s | %d |\n", t, count)
		}
		fmt.Fprintln(r.output)
	}

	// 按优先级统计
	if len(result.Summary.ByPriority) > 0 {
		fmt.Fprintln(r.output, "### 按优先级统计")
		fmt.Fprintln(r.output)
		fmt.Fprintln(r.output, "| 优先级 | 数量 |")
		fmt.Fprintln(r.output, "|--------|------|")
		for p, count := range result.Summary.ByPriority {
			fmt.Fprintf(r.output, "| %s | %d |\n", p, count)
		}
		fmt.Fprintln(r.output)
	}

	// TODO 列表
	if len(result.TODOs) > 0 {
		fmt.Fprintln(r.output, "## TODO 列表")
		fmt.Fprintln(r.output)

		// 表头
		fmt.Fprintln(r.output, "| 优先级 | 类型 | 位置 | 描述 |")
		fmt.Fprintln(r.output, "|--------|------|------|------|")

		// 内容
		for _, todo := range result.TODOs {
			priority := getPriorityEmoji(todo.Priority)
			location := fmt.Sprintf("[%s:%d](%s#L%d)", todo.File, todo.Line, todo.File, todo.Line)
			message := escapeMarkdown(todo.Message)

			fmt.Fprintf(r.output, "| %s | %s | %s | %s |\n",
				priority, todo.Type, location, message)
		}
		fmt.Fprintln(r.output)
	} else {
		fmt.Fprintln(r.output, "## TODO 列表")
		fmt.Fprintln(r.output)
		fmt.Fprintln(r.output, "暂无 TODO")
		fmt.Fprintln(r.output)
	}

	// 警告
	if len(result.Warnings) > 0 {
		fmt.Fprintln(r.output, "## 警告")
		fmt.Fprintln(r.output)
		for _, w := range result.Warnings {
			fmt.Fprintf(r.output, "- **%s:%d**: %s\n", w.File, w.Line, w.Message)
		}
		fmt.Fprintln(r.output)
	}

	// 页脚
	fmt.Fprintln(r.output, "---")
	fmt.Fprintln(r.output, "*由 [TODO Tracker](https://github.com/todo-tracker/todo-tracker) 生成*")

	return nil
}

// ReportStale 生成过期 TODO 报告
func (r *MarkdownReporter) ReportStale(todos []types.TODO, threshold int) error {
	fmt.Fprintln(r.output, "# 过期 TODO 报告")
	fmt.Fprintln(r.output)
	fmt.Fprintf(r.output, "> 过期阈值: %d 天\n", threshold)
	fmt.Fprintf(r.output, "> 生成时间: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Fprintln(r.output)

	if len(todos) == 0 {
		fmt.Fprintln(r.output, "未发现过期 TODO")
		return nil
	}

	fmt.Fprintf(r.output, "发现 %d 个过期 TODO:\n\n", len(todos))

	fmt.Fprintln(r.output, "| 年龄 | 文件 | 修改次数 | 作者 | 描述 |")
	fmt.Fprintln(r.output, "|------|------|----------|------|------|")

	for _, todo := range todos {
		age := formatAge(todo.Age)
		location := fmt.Sprintf("%s:%d", todo.File, todo.Line)

		fmt.Fprintf(r.output, "| %s | %s | %d | %s | %s |\n",
			age, location, todo.ChurnScore, todo.Author, escapeMarkdown(todo.Message))
	}

	return nil
}

// ReportOrphaned 生成孤儿 TODO 报告
func (r *MarkdownReporter) ReportOrphaned(todos []types.TODO) error {
	fmt.Fprintln(r.output, "# 孤儿 TODO 报告")
	fmt.Fprintln(r.output)
	fmt.Fprintf(r.output, "> 生成时间: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Fprintln(r.output)

	if len(todos) == 0 {
		fmt.Fprintln(r.output, "未发现孤儿 TODO")
		return nil
	}

	fmt.Fprintf(r.output, "发现 %d 个孤儿 TODO:\n\n", len(todos))

	// 按作者分组
	byAuthor := make(map[string][]types.TODO)
	for _, todo := range todos {
		author := todo.Author
		if author == "" {
			author = "未知"
		}
		byAuthor[author] = append(byAuthor[author], todo)
	}

	for author, authorTodos := range byAuthor {
		fmt.Fprintf(r.output, "## @%s (%d 个)\n\n", author, len(authorTodos))

		for _, todo := range authorTodos {
			priority := getPriorityEmoji(todo.Priority)
			fmt.Fprintf(r.output, "- %s [%s:%d] %s\n",
				priority, todo.File, todo.Line, escapeMarkdown(todo.Message))
		}
		fmt.Fprintln(r.output)
	}

	return nil
}

// getPriorityEmoji 获取优先级对应的 emoji
func getPriorityEmoji(priority string) string {
	switch strings.ToLower(priority) {
	case "high":
		return ":red_circle: 高"
	case "medium":
		return ":yellow_circle: 中"
	case "low":
		return ":green_circle: 低"
	default:
		return priority
	}
}

// escapeMarkdown 转义 Markdown 特殊字符
func escapeMarkdown(s string) string {
	replacer := strings.NewReplacer(
		"|", "\\|",
		"\n", " ",
	)
	return replacer.Replace(s)
}

// formatAge 格式化年龄
func formatAge(days int) string {
	if days < 30 {
		return fmt.Sprintf("%d 天", days)
	} else if days < 365 {
		return fmt.Sprintf("%.1f 个月", float64(days)/30)
	}
	return fmt.Sprintf("%.1f 年", float64(days)/365)
}
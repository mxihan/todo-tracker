// Package reporter 提供 JSON 格式报告生成功能
package reporter

import (
	"encoding/json"
	"io"
	"os"
	"time"

	"github.com/mxihan/todo-tracker/pkg/types"
)

// JSONReporter JSON报告生成器
type JSONReporter struct {
	writer io.Writer
	indent bool
}

// NewJSONReporter 创建新的JSON报告生成器
func NewJSONReporter(opts ...JSONOption) *JSONReporter {
	r := &JSONReporter{
		writer: os.Stdout,
		indent: true,
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

// JSONOption JSON报告选项
type JSONOption func(*JSONReporter)

// WithJSONWriter 设置输出写入器
func WithJSONWriter(w io.Writer) JSONOption {
	return func(r *JSONReporter) {
		r.writer = w
	}
}

// WithIndent 设置是否缩进
func WithIndent(indent bool) JSONOption {
	return func(r *JSONReporter) {
		r.indent = indent
	}
}

// Report 生成JSON报告
func (r *JSONReporter) Report(result *types.ScanResult) error {
	data := &JSONReport{
		GeneratedAt: time.Now().Format(time.RFC3339),
		Summary:     jsonSummaryFromSummary(&result.Summary),
		TODOs:       result.TODOs,
		Warnings:    result.Warnings,
	}

	return r.encode(data)
}

// ReportStale 生成过期TODO的JSON报告
func (r *JSONReporter) ReportStale(todos []types.TODO, thresholdDays int) error {
	data := &StaleReport{
		GeneratedAt:   time.Now().Format(time.RFC3339),
		ThresholdDays: thresholdDays,
		Count:         len(todos),
		TODOs:         todos,
	}

	return r.encode(data)
}

// ReportOrphaned 生成孤儿TODO的JSON报告
func (r *JSONReporter) ReportOrphaned(todos []types.TODO, inactiveDays int) error {
	// 按作者分组
	byAuthor := make(map[string][]types.TODO)
	for _, todo := range todos {
		author := todo.Author
		if author == "" {
			author = "未知"
		}
		byAuthor[author] = append(byAuthor[author], todo)
	}

	authors := make([]AuthorSummary, 0, len(byAuthor))
	for author, authorTodos := range byAuthor {
		high, medium, low := countByPriority(authorTodos)
		authors = append(authors, AuthorSummary{
			Name:     author,
			Count:    len(authorTodos),
			High:     high,
			Medium:   medium,
			Low:      low,
			TODOs:    authorTodos,
		})
	}

	data := &OrphanedReport{
		GeneratedAt:  time.Now().Format(time.RFC3339),
		InactiveDays: inactiveDays,
		Count:        len(todos),
		Authors:      authors,
	}

	return r.encode(data)
}

// encode 编码并输出JSON
func (r *JSONReporter) encode(data interface{}) error {
	encoder := json.NewEncoder(r.writer)
	if r.indent {
		encoder.SetIndent("", "  ")
	}
	return encoder.Encode(data)
}

// JSONReport JSON报告结构
type JSONReport struct {
	GeneratedAt string        `json:"generated_at"`
	Summary     jsonSummary   `json:"summary"`
	TODOs       []types.TODO  `json:"todos"`
	Warnings    []types.Warning `json:"warnings"`
}

// jsonSummary JSON摘要结构
type jsonSummary struct {
	Total        int            `json:"total"`
	FilesScanned int            `json:"files_scanned"`
	Duration     string         `json:"duration"`
	ByType       map[string]int `json:"by_type"`
	ByPriority   map[string]int `json:"by_priority"`
	ByAuthor     map[string]int `json:"by_author"`
}

// jsonSummaryFromSummary 从Summary创建jsonSummary
func jsonSummaryFromSummary(s *types.Summary) jsonSummary {
	return jsonSummary{
		Total:        s.Total,
		FilesScanned: s.FilesScanned,
		Duration:     s.Duration.String(),
		ByType:       s.ByType,
		ByPriority:   s.ByPriority,
		ByAuthor:     s.ByAuthor,
	}
}

// StaleReport 过期TODO报告结构
type StaleReport struct {
	GeneratedAt   string       `json:"generated_at"`
	ThresholdDays int          `json:"threshold_days"`
	Count         int          `json:"count"`
	TODOs         []types.TODO `json:"todos"`
}

// OrphanedReport 孤儿TODO报告结构
type OrphanedReport struct {
	GeneratedAt  string          `json:"generated_at"`
	InactiveDays int             `json:"inactive_days"`
	Count        int             `json:"count"`
	Authors      []AuthorSummary `json:"authors"`
}

// AuthorSummary 作者摘要
type AuthorSummary struct {
	Name   string       `json:"name"`
	Count  int          `json:"count"`
	High   int          `json:"high"`
	Medium int          `json:"medium"`
	Low    int          `json:"low"`
	TODOs  []types.TODO `json:"todos"`
}
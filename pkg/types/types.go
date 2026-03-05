// Package types 定义TODO Tracker的核心数据类型
package types

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

// TODO 表示一个待办事项
type TODO struct {
	// 基本信息
	ID        string `json:"id"`          // 唯一标识符
	Type      string `json:"type"`        // 类型: TODO, FIXME, HACK, BUG, XXX
	Message   string `json:"message"`     // TODO描述内容
	File      string `json:"file"`        // 文件路径
	Line      int    `json:"line"`        // 起始行号
	LineEnd   int    `json:"line_end"`    // 结束行号（多行TODO）
	Priority  string `json:"priority"`    // 优先级: high, medium, low
	Assignee  string `json:"assignee"`    // 负责人
	TicketRef string `json:"ticket_ref"`  // 关联的工单号，如 #123, JIRA-456

	// Git元数据
	Author       string     `json:"author"`        // git blame获取的作者
	CommitHash   string     `json:"commit_hash"`  // 提交哈希
	CreatedAt    time.Time  `json:"created_at"`   // TODO创建时间
	LastModified time.Time  `json:"last_modified"` // 最后修改时间

	// 状态
	Status string `json:"status"` // open, resolved, wontfix

	// 计算字段
	Age         int  `json:"age_days"`    // TODO年龄（天）
	ChurnScore  int  `json:"churn_score"` // 文件修改次数
	IsOrphaned  bool `json:"is_orphaned"` // 作者是否离开
}

// GenerateID 根据文件路径和行号生成唯一ID
func (t *TODO) GenerateID() string {
	data := fmt.Sprintf("%s:%d", t.File, t.Line)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:8])
}

// IsStale 检查TODO是否过期
func (t *TODO) IsStale(thresholdDays int) bool {
	if t.CreatedAt.IsZero() {
		return false
	}
	age := int(time.Since(t.CreatedAt).Hours() / 24)
	return age > thresholdDays
}

// FormatAge 返回人类可读的年龄描述
func (t *TODO) FormatAge() string {
	if t.CreatedAt.IsZero() {
		return "未知"
	}

	age := int(time.Since(t.CreatedAt).Hours() / 24)
	switch {
	case age < 30:
		return fmt.Sprintf("%d天", age)
	case age < 365:
		return fmt.Sprintf("%.1f个月", float64(age)/30)
	default:
		return fmt.Sprintf("%.1f年", float64(age)/365)
	}
}

// Summary 扫描结果摘要
type Summary struct {
	Total        int            `json:"total"`         // TODO总数
	FilesScanned int            `json:"files_scanned"` // 扫描文件数
	Duration     time.Duration  `json:"duration"`      // 扫描耗时
	ByType       map[string]int `json:"by_type"`       // 按类型统计
	ByPriority   map[string]int `json:"by_priority"`   // 按优先级统计
	ByAuthor     map[string]int `json:"by_author"`     // 按作者统计
}

// ScanResult 扫描结果
type ScanResult struct {
	Summary  Summary    `json:"summary"`  // 摘要
	TODOs    []TODO     `json:"todos"`    // TODO列表
	Warnings []Warning  `json:"warnings"` // 警告列表
}

// Warning 警告信息
type Warning struct {
	File    string `json:"file"`    // 文件路径
	Line    int    `json:"line"`    // 行号
	Message string `json:"message"` // 警告消息
	Type    string `json:"type"`    // 警告类型
}

// FileRecord 文件扫描记录（用于缓存）
type FileRecord struct {
	Path        string    `json:"path"`         // 文件路径
	Hash        string    `json:"hash"`         // 文件内容哈希
	LastScanned time.Time `json:"last_scanned"` // 最后扫描时间
	SizeBytes   int64     `json:"size_bytes"`   // 文件大小
	ChurnCount  int       `json:"churn_count"`  // 文件修改次数
}

// Author 作者信息
type Author struct {
	Name        string    `json:"name"`         // 作者名
	LastCommit  time.Time `json:"last_commit"`  // 最后提交时间
	CommitCount int       `json:"commit_count"` // 提交次数
	IsActive    bool      `json:"is_active"`    // 是否活跃
}

// Config 配置结构
type Config struct {
	Version int         `json:"version" yaml:"version"`
	Scan    ScanConfig  `json:"scan" yaml:"scan"`
	Git     GitConfig   `json:"git" yaml:"git"`
	Stale   StaleConfig `json:"stale" yaml:"stale"`
	Orphan  OrphanConfig `json:"orphan" yaml:"orphan"`
	Output  OutputConfig `json:"output" yaml:"output"`
}

// ScanConfig 扫描配置
type ScanConfig struct {
	Paths   []string `json:"paths" yaml:"paths"`
	Exclude []string `json:"exclude" yaml:"exclude"`
	Workers int      `json:"workers" yaml:"workers"`
}

// GitConfig Git集成配置
type GitConfig struct {
	Enabled    bool   `json:"enabled" yaml:"enabled"`
	Blame      bool   `json:"blame" yaml:"blame"`
	DefaultBase string `json:"default_base" yaml:"default_base"`
}

// StaleConfig 过期检测配置
type StaleConfig struct {
	ThresholdDays int `json:"threshold_days" yaml:"threshold_days"`
	ChurnThreshold int `json:"churn_threshold" yaml:"churn_threshold"`
}

// OrphanConfig 孤儿检测配置
type OrphanConfig struct {
	InactiveDays int `json:"inactive_days" yaml:"inactive_days"`
}

// OutputConfig 输出配置
type OutputConfig struct {
	Format   string `json:"format" yaml:"format"`
	Color    string `json:"color" yaml:"color"`
	Truncate int    `json:"truncate" yaml:"truncate"`
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Version: 1,
		Scan: ScanConfig{
			Paths:   []string{"."},
			Exclude: []string{"**/node_modules/**", "**/vendor/**", "**/.git/**", "**/dist/**"},
			Workers: 0, // 0表示自动检测CPU核心数
		},
		Git: GitConfig{
			Enabled:     true,
			Blame:       true,
			DefaultBase: "main",
		},
		Stale: StaleConfig{
			ThresholdDays: 90,
			ChurnThreshold: 10,
		},
		Orphan: OrphanConfig{
			InactiveDays: 180,
		},
		Output: OutputConfig{
			Format:   "table",
			Color:    "auto",
			Truncate: 80,
		},
	}
}

// PatternConfig TODO模式配置
type PatternConfig struct {
	Types            []string          `json:"types" yaml:"types"`
	PriorityMarkers  map[string]string `json:"priority_markers" yaml:"priority_markers"`
	AssigneePattern  string            `json:"assignee_pattern" yaml:"assignee_pattern"`
	TicketPattern    string            `json:"ticket_pattern" yaml:"ticket_pattern"`
}

// DefaultPatternConfig 返回默认模式配置
func DefaultPatternConfig() *PatternConfig {
	return &PatternConfig{
		Types: []string{"TODO", "FIXME", "HACK", "XXX", "BUG"},
		PriorityMarkers: map[string]string{
			"high":   "!",
			"medium": ">",
		},
		AssigneePattern: `\(([^)]+)\)|@(\w+)`,
		TicketPattern:   `#(\d+)|([A-Z]+-\d+)`,
	}
}
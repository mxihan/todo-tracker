// Package parser 定义TODO模式
package parser

import (
	"regexp"
	"strings"

	"github.com/todo-tracker/todo-tracker/pkg/types"
)

// Pattern TODO模式定义
type Pattern struct {
	Name        string         // 模式名称
	Regex       *regexp.Regexp // 正则表达式
	Priority    string         // 默认优先级
	Description string         // 描述
}

// PatternSet 模式集合
type PatternSet struct {
	patterns []Pattern
	config   *types.PatternConfig
}

// NewPatternSet 创建新的模式集合
func NewPatternSet(config *types.PatternConfig) *PatternSet {
	if config == nil {
		config = types.DefaultPatternConfig()
	}

	ps := &PatternSet{
		config: config,
	}

	// 初始化模式
	ps.initPatterns()

	return ps
}

// initPatterns 初始化所有模式
func (ps *PatternSet) initPatterns() {
	// 基础TODO模式
	for _, todoType := range ps.config.Types {
		// 标准格式: TODO: 描述
		ps.patterns = append(ps.patterns, Pattern{
			Name:        todoType + "_standard",
			Regex:       regexp.MustCompile(`(?i)\b` + todoType + `\b\s*:\s*(.*)`),
			Priority:    ps.getDefaultPriority(todoType),
			Description: "标准" + todoType + "格式",
		})

		// 高优先级格式: TODO!: 描述
		ps.patterns = append(ps.patterns, Pattern{
			Name:        todoType + "_high",
			Regex:       regexp.MustCompile(`(?i)\b` + todoType + `\s*!\s*:\s*(.*)`),
			Priority:    "high",
			Description: "高优先级" + todoType + "格式",
		})

		// 带负责人格式: TODO(@user): 描述
		ps.patterns = append(ps.patterns, Pattern{
			Name:        todoType + "_assignee",
			Regex:       regexp.MustCompile(`(?i)\b` + todoType + `\s*\(\s*(@?\w+)\s*\)\s*:\s*(.*)`),
			Priority:    ps.getDefaultPriority(todoType),
			Description: "带负责人的" + todoType + "格式",
		})

		// 带工单格式: TODO #123: 描述
		ps.patterns = append(ps.patterns, Pattern{
			Name:        todoType + "_ticket",
			Regex:       regexp.MustCompile(`(?i)\b` + todoType + `\s+(#\d+|[A-Z]+-\d+)\s*:\s*(.*)`),
			Priority:    ps.getDefaultPriority(todoType),
			Description: "带工单的" + todoType + "格式",
		})

		// 组合格式: TODO(@user) #123!: 描述
		ps.patterns = append(ps.patterns, Pattern{
			Name:        todoType + "_combined",
			Regex:       regexp.MustCompile(`(?i)\b` + todoType + `\s*\(\s*(@?\w+)\s*\)\s+(#\d+|[A-Z]+-\d+)\s*(!?)\s*:\s*(.*)`),
			Priority:    "medium",
			Description: "组合格式" + todoType,
		})
	}
}

// getDefaultPriority 根据TODO类型获取默认优先级
func (ps *PatternSet) getDefaultPriority(todoType string) string {
	switch strings.ToUpper(todoType) {
	case "BUG", "FIXME":
		return "high"
	case "HACK", "XXX":
		return "medium"
	default:
		return "low"
	}
}

// Match 匹配文本并返回匹配的模式
func (ps *PatternSet) Match(text string) *PatternMatch {
	for _, pattern := range ps.patterns {
		if matches := pattern.Regex.FindStringSubmatch(text); matches != nil {
			return &PatternMatch{
				Pattern: pattern,
				Matches: matches,
			}
		}
	}
	return nil
}

// PatternMatch 模式匹配结果
type PatternMatch struct {
	Pattern Pattern
	Matches []string
}

// ExtractMetadata 从文本中提取元数据
func (ps *PatternSet) ExtractMetadata(text string) *TODOMetadata {
	metadata := &TODOMetadata{}

	// 提取负责人
	assigneePattern := regexp.MustCompile(`(?i)@(\w+)|\(([^)]+)\)`)
	if matches := assigneePattern.FindStringSubmatch(text); matches != nil {
		if matches[1] != "" {
			metadata.Assignee = matches[1]
		} else if matches[2] != "" {
			metadata.Assignee = strings.TrimSpace(matches[2])
		}
	}

	// 提取工单号
	ticketPattern := regexp.MustCompile(`#(\d+)|([A-Z]+-\d+)`)
	if matches := ticketPattern.FindStringSubmatch(text); matches != nil {
		if matches[1] != "" {
			metadata.TicketRef = "#" + matches[1]
		} else if matches[2] != "" {
			metadata.TicketRef = matches[2]
		}
	}

	// 提取日期
	datePattern := regexp.MustCompile(`\[(\d{4}-\d{2}-\d{2})\]`)
	if matches := datePattern.FindStringSubmatch(text); matches != nil {
		metadata.DueDate = matches[1]
	}

	// 提取优先级标记
	if strings.Contains(text, "!") || strings.Contains(strings.ToUpper(text), "URGENT") || strings.Contains(strings.ToUpper(text), "CRITICAL") {
		metadata.Priority = "high"
	} else if strings.Contains(text, ">") || strings.Contains(strings.ToUpper(text), "MEDIUM") {
		metadata.Priority = "medium"
	}

	return metadata
}

// TODOMetadata TODO元数据
type TODOMetadata struct {
	Assignee  string // 负责人
	TicketRef string // 工单引用
	DueDate   string // 截止日期
	Priority  string // 优先级
}

// GetTypes 返回支持的TODO类型
func (ps *PatternSet) GetTypes() []string {
	return ps.config.Types
}

// GetPriorityMarkers 返回优先级标记
func (ps *PatternSet) GetPriorityMarkers() map[string]string {
	return ps.config.PriorityMarkers
}
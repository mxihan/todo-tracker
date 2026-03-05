// Package parser 提供TODO注释解析功能
package parser

import (
	"regexp"
	"strings"

	"github.com/mxihan/todo-tracker/pkg/types"
)

// Parser TODO解析器
type Parser struct {
	config    *types.PatternConfig
	patterns  map[string]*regexp.Regexp
	languages map[string]*LanguageConfig
}

// NewParser 创建新的解析器
func NewParser(config *types.PatternConfig) *Parser {
	if config == nil {
		config = types.DefaultPatternConfig()
	}

	p := &Parser{
		config:    config,
		patterns:  make(map[string]*regexp.Regexp),
		languages: DefaultLanguages(),
	}

	// 编译正则表达式
	p.compilePatterns()

	return p
}

// compilePatterns 编译所有正则表达式模式
func (p *Parser) compilePatterns() {
	// TODO类型匹配模式
	typesPattern := strings.Join(p.config.Types, "|")
	p.patterns["todo"] = regexp.MustCompile(`(?i)\b(` + typesPattern + `)\b\s*(!?)\s*(?:\(([^)]+)\))?\s*(?:#(\d+)|([A-Z]+-\d+))?\s*:?\s*(.*)`)

	// 负责人匹配模式
	if p.config.AssigneePattern != "" {
		p.patterns["assignee"] = regexp.MustCompile(p.config.AssigneePattern)
	}

	// 工单引用匹配模式
	if p.config.TicketPattern != "" {
		p.patterns["ticket"] = regexp.MustCompile(p.config.TicketPattern)
	}

	// 日期匹配模式
	p.patterns["date"] = regexp.MustCompile(`\[(\d{4}-\d{2}-\d{2})\]`)

	// 优先级标记
	p.patterns["priority_high"] = regexp.MustCompile(`!|URGENT|CRITICAL`)
	p.patterns["priority_medium"] = regexp.MustCompile(`>|MEDIUM`)
}

// ParseFile 解析单个文件
func (p *Parser) ParseFile(content string, filePath string) []types.TODO {
	var todos []types.TODO
	lines := strings.Split(content, "\n")

	// 获取文件语言配置
	lang := p.getLanguageConfig(filePath)

	// 单行注释处理
	for lineNum, line := range lines {
		// 跳过空行
		if strings.TrimSpace(line) == "" {
			continue
		}

		// 检查是否包含TODO类型关键字
		if p.containsTODOType(line) {
			todo := p.parseLine(line, filePath, lineNum+1, lang)
			if todo != nil {
				todos = append(todos, *todo)
			}
		}
	}

	// 多行注释处理
	if lang != nil && lang.MultiLineStart != "" {
		multiLineTodos := p.parseMultiLineComments(content, filePath, lang)
		todos = append(todos, multiLineTodos...)
	}

	return todos
}

// parseLine 解析单行TODO
func (p *Parser) parseLine(line string, filePath string, lineNum int, lang *LanguageConfig) *types.TODO {
	// 提取注释内容
	commentContent := p.extractCommentContent(line, lang)
	if commentContent == "" {
		return nil
	}

	// 匹配TODO模式
	match := p.patterns["todo"].FindStringSubmatch(commentContent)
	if match == nil {
		return nil
	}

	todo := &types.TODO{
		File:     filePath,
		Line:     lineNum,
		LineEnd:  lineNum,
		Type:     strings.ToUpper(match[1]),
		Message:  strings.TrimSpace(match[6]),
		Priority: "low",
		Status:   "open",
	}

	// 解析优先级
	if match[2] == "!" || p.patterns["priority_high"].MatchString(commentContent) {
		todo.Priority = "high"
	} else if p.patterns["priority_medium"].MatchString(commentContent) {
		todo.Priority = "medium"
	}

	// 解析负责人
	if match[3] != "" {
		todo.Assignee = match[3]
	} else if p.patterns["assignee"] != nil {
		if assigneeMatch := p.patterns["assignee"].FindStringSubmatch(commentContent); assigneeMatch != nil {
			if assigneeMatch[1] != "" {
				todo.Assignee = assigneeMatch[1]
			} else if assigneeMatch[2] != "" {
				todo.Assignee = assigneeMatch[2]
			}
		}
	}

	// 解析工单引用
	if match[4] != "" {
		todo.TicketRef = "#" + match[4]
	} else if match[5] != "" {
		todo.TicketRef = match[5]
	} else if p.patterns["ticket"] != nil {
		if ticketMatch := p.patterns["ticket"].FindStringSubmatch(commentContent); ticketMatch != nil {
			if ticketMatch[1] != "" {
				todo.TicketRef = "#" + ticketMatch[1]
			} else if ticketMatch[2] != "" {
				todo.TicketRef = ticketMatch[2]
			}
		}
	}

	// 生成ID
	todo.ID = todo.GenerateID()

	return todo
}

// parseMultiLineComments 解析多行注释中的TODO
func (p *Parser) parseMultiLineComments(content string, filePath string, lang *LanguageConfig) []types.TODO {
	var todos []types.TODO

	// 查找所有多行注释块
	startPattern := regexp.MustCompile(regexp.QuoteMeta(lang.MultiLineStart))
	endPattern := regexp.MustCompile(regexp.QuoteMeta(lang.MultiLineEnd))

	startMatches := startPattern.FindAllStringIndex(content, -1)
	endMatches := endPattern.FindAllStringIndex(content, -1)

	if len(startMatches) == 0 || len(endMatches) == 0 {
		return todos
	}

	// 匹配注释块
	for i, start := range startMatches {
		if i >= len(endMatches) {
			break
		}

		end := endMatches[i]
		if end[0] < start[0] {
			continue
		}

		// 提取注释块内容
		commentBlock := content[start[1]:end[0]]
		lines := strings.Split(commentBlock, "\n")

		// 计算起始行号
		lineNum := strings.Count(content[:start[0]], "\n") + 2

		for j, line := range lines {
			if p.containsTODOType(line) {
				todo := p.parseLine(line, filePath, lineNum+j, lang)
				if todo != nil {
					todo.LineEnd = lineNum + j
					todos = append(todos, *todo)
				}
			}
		}
	}

	return todos
}

// extractCommentContent 提取注释内容
func (p *Parser) extractCommentContent(line string, lang *LanguageConfig) string {
	if lang == nil {
		return line
	}

	// 尝试匹配单行注释
	for _, marker := range lang.SingleLine {
		if strings.Contains(line, marker) {
			idx := strings.Index(line, marker)
			if idx >= 0 {
				return strings.TrimSpace(line[idx+len(marker):])
			}
		}
	}

	return line
}

// containsTODOType 检查行是否包含TODO类型关键字
func (p *Parser) containsTODOType(line string) bool {
	for _, todoType := range p.config.Types {
		if strings.Contains(strings.ToUpper(line), todoType) {
			return true
		}
	}
	return false
}

// getLanguageConfig 根据文件扩展名获取语言配置
func (p *Parser) getLanguageConfig(filePath string) *LanguageConfig {
	ext := ""
	if idx := strings.LastIndex(filePath, "."); idx >= 0 {
		ext = strings.ToLower(filePath[idx:])
	}

	if lang, ok := p.languages[ext]; ok {
		return lang
	}

	return nil
}

// Result 解析结果
type Result struct {
	TODOs []types.TODO
}
// Package scanner 提供 TODO 扫描功能
package scanner

import (
	"path/filepath"
	"strings"
)

// SkipRules 跳过规则定义
type SkipRules struct {
	directoryPatterns []string
	filePatterns      []string
	extensions        []string
}

// DefaultSkipRules 返回默认的跳过规则
func DefaultSkipRules() *SkipRules {
	return &SkipRules{
		directoryPatterns: []string{
			// 版本控制
			".git",
			".svn",
			".hg",
			".bzr",

			// 依赖目录
			"node_modules",
			"vendor",
			"venv",
			".venv",
			"env",
			".env",

			// 构建输出
			"dist",
			"build",
			"target",
			"out",
			"bin",
			"pkg",

			// 缓存目录
			".cache",
			".tmp",
			"__pycache__",
			".pytest_cache",
			".mypy_cache",

			// IDE 配置
			".idea",
			".vscode",
			".vs",
			".eclipse",

			// 其他
			"coverage",
			".nyc_output",
			" Pods",
		},
		filePatterns: []string{
			// 压缩文件
			"*.zip",
			"*.tar",
			"*.gz",
			"*.rar",
			"*.7z",

			// 二进制文件
			"*.exe",
			"*.dll",
			"*.so",
			"*.dylib",
			"*.bin",

			// 图像文件
			"*.png",
			"*.jpg",
			"*.jpeg",
			"*.gif",
			"*.ico",
			"*.svg",

			// 压缩资源
			"*.min.js",
			"*.min.css",

			// 锁文件
			"*.lock",
			"*.sum",
			"package-lock.json",
			"yarn.lock",
			"pnpm-lock.yaml",
			"go.sum",
			"Cargo.lock",

			// 数据文件
			"*.db",
			"*.sqlite",
			"*.sqlite3",
		},
		extensions: []string{
			// 二进制扩展名
			".exe", ".dll", ".so", ".dylib",
			".png", ".jpg", ".jpeg", ".gif", ".ico", ".svg",
			".pdf", ".doc", ".docx", ".xls", ".xlsx",
			".zip", ".tar", ".gz", ".rar", ".7z",
			".mp3", ".mp4", ".avi", ".mov", ".wav",
			".ttf", ".otf", ".woff", ".woff2", ".eot",
		},
	}
}

// ShouldSkipDirectory 检查是否应该跳过目录
func (r *SkipRules) ShouldSkipDirectory(path string) bool {
	base := filepath.Base(path)

	for _, pattern := range r.directoryPatterns {
		// 精确匹配
		if base == pattern {
			return true
		}
		// 通配符匹配
		if matched, _ := filepath.Match(pattern, base); matched {
			return true
		}
	}

	return false
}

// ShouldSkipFile 检查是否应该跳过文件
func (r *SkipRules) ShouldSkipFile(path string) bool {
	base := filepath.Base(path)
	ext := strings.ToLower(filepath.Ext(path))

	// 检查文件模式
	for _, pattern := range r.filePatterns {
		if matched, _ := filepath.Match(pattern, base); matched {
			return true
		}
		// 检查压缩资源模式（如 *.min.js）
		if strings.HasSuffix(base, pattern[1:]) {
			return true
		}
	}

	// 检查扩展名
	for _, skipExt := range r.extensions {
		if ext == skipExt {
			return true
		}
	}

	return false
}

// AddDirectoryPattern 添加目录跳过模式
func (r *SkipRules) AddDirectoryPattern(pattern string) {
	r.directoryPatterns = append(r.directoryPatterns, pattern)
}

// AddFilePattern 添加文件跳过模式
func (r *SkipRules) AddFilePattern(pattern string) {
	r.filePatterns = append(r.filePatterns, pattern)
}

// Merge 合并另一个跳过规则
func (r *SkipRules) Merge(other *SkipRules) {
	if other == nil {
		return
	}

	r.directoryPatterns = append(r.directoryPatterns, other.directoryPatterns...)
	r.filePatterns = append(r.filePatterns, other.filePatterns...)
	r.extensions = append(r.extensions, other.extensions...)
}

// FromConfig 从配置创建跳过规则
func FromConfig(excludePatterns []string) *SkipRules {
	rules := DefaultSkipRules()

	for _, pattern := range excludePatterns {
		// 解析 glob 模式
		cleanPattern := strings.TrimPrefix(pattern, "**/")

		if strings.Contains(cleanPattern, "/") {
			// 包含路径分隔符，提取所有目录名
			parts := strings.Split(cleanPattern, "/")
			for _, part := range parts {
				// 跳过空字符串、通配符和文件模式
				if part != "" && part != "**" && part != "*" && !strings.HasPrefix(part, "*.") {
					rules.AddDirectoryPattern(part)
				}
			}
			// 检查最后部分是否是文件模式
			if len(parts) > 0 && strings.HasPrefix(parts[len(parts)-1], "*.") {
				rules.AddFilePattern(parts[len(parts)-1])
			}
		} else if strings.HasPrefix(cleanPattern, "*.") {
			// 文件扩展名模式
			rules.AddFilePattern(cleanPattern)
		} else {
			// 可能是目录名
			rules.AddDirectoryPattern(cleanPattern)
		}
	}

	return rules
}
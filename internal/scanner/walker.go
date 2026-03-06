// Package scanner 提供 TODO 扫描功能
package scanner

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/mxihan/todo-tracker/pkg/types"
)

// Walker 目录遍历器
type Walker struct {
	config    *types.Config
	skipDirs  []string
	skipFiles []string
}

// NewWalker 创建新的目录遍历器
func NewWalker(config *types.Config) *Walker {
	// 默认跳过的目录
	defaultSkipDirs := []string{
		".git",
		"node_modules",
		"vendor",
		"dist",
		"build",
		"target",
		".idea",
		".vscode",
		"__pycache__",
		".cache",
		"coverage",
	}

	// 默认跳过的文件（使用通配符模式）
	defaultSkipFiles := []string{
		// 压缩资源
		"*.min.js",
		"*.min.css",
		// 锁文件（通配符模式）
		"*.lock",
		"*.sum",
		// 特定锁文件
		"package-lock.json",
		"yarn.lock",
		"pnpm-lock.yaml",
		"go.sum",
		"Cargo.lock",
	}

	// 合并配置中的排除规则
	skipDirs := append(defaultSkipDirs, extractDirPatterns(config.Scan.Exclude)...)
	skipFiles := append(defaultSkipFiles, extractFilePatterns(config.Scan.Exclude)...)

	return &Walker{
		config:    config,
		skipDirs:  skipDirs,
		skipFiles: skipFiles,
	}
}

// Walk 遍历目录，返回文件通道
func (w *Walker) Walk(root string) (<-chan string, <-chan error) {
	fileCh := make(chan string, 100)
	errCh := make(chan error, 1)

	go func() {
		defer close(fileCh)
		defer close(errCh)

		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// 跳过目录
			if info.IsDir() {
				if w.shouldSkipDir(path) {
					return filepath.SkipDir
				}
				return nil
			}

			// 跳过文件
			if w.shouldSkipFile(path) {
				return nil
			}

			// 发送文件路径
			fileCh <- path
			return nil
		})

		if err != nil {
			errCh <- err
		}
	}()

	return fileCh, errCh
}

// shouldSkipDir 检查是否应该跳过目录
func (w *Walker) shouldSkipDir(path string) bool {
	base := filepath.Base(path)

	for _, pattern := range w.skipDirs {
		if matched, _ := filepath.Match(pattern, base); matched {
			return true
		}
		if strings.Contains(path, pattern) {
			return true
		}
	}

	return false
}

// shouldSkipFile 检查是否应该跳过文件
func (w *Walker) shouldSkipFile(path string) bool {
	base := filepath.Base(path)

	for _, pattern := range w.skipFiles {
		// 尝试通配符匹配 (如 *.lock, *.min.js)
		if matched, _ := filepath.Match(pattern, base); matched {
			return true
		}

		// 尝试精确匹配 (如 package-lock.json)
		if base == pattern {
			return true
		}

		// 尝试后缀匹配 (如 .min.js 匹配 app.min.js)
		if strings.HasPrefix(pattern, "*.") {
			suffix := pattern[1:] // 将 "*.lock" 转换为 ".lock"
			if strings.HasSuffix(base, suffix) {
				return true
			}
		}
	}

	// 检查是否是二进制文件
	if isBinary, _ := isBinaryFile(path); isBinary {
		return true
	}

	return false
}

// extractDirPatterns 从排除规则中提取目录模式
func extractDirPatterns(excludes []string) []string {
	var dirs []string
	for _, exclude := range excludes {
		if strings.Contains(exclude, "/") {
			parts := strings.Split(exclude, "/")
			for _, part := range parts {
				if part != "" && part != "**" && part != "*" {
					dirs = append(dirs, part)
				}
			}
		}
	}
	return dirs
}

// extractFilePatterns 从排除规则中提取文件模式
func extractFilePatterns(excludes []string) []string {
	var files []string
	for _, exclude := range excludes {
		if strings.Contains(exclude, "*.") {
			ext := strings.TrimPrefix(exclude, "**/*")
			files = append(files, ext)
		}
	}
	return files
}

// isBinaryFile 检查是否是二进制文件
func isBinaryFile(path string) (bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer file.Close()

	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil {
		return false, err
	}

	// 检查是否包含空字节（二进制文件的典型特征）
	for i := 0; i < n; i++ {
		if buf[i] == 0 {
			return true, nil
		}
	}

	return false, nil
}
// Package scanner 提供 TODO 扫描功能
// 负责遍历目录、解析文件并提取 TODO 注释
package scanner

import (
	"context"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/mxihan/todo-tracker/internal/git"
	"github.com/mxihan/todo-tracker/internal/parser"
	"github.com/mxihan/todo-tracker/pkg/types"
)

// Scanner 主扫描器结构
type Scanner struct {
	config    *types.Config
	workers   int
	parser    *parser.Parser
	gitClient *git.Client
}

// NewScanner 创建新的扫描器实例
func NewScanner(config *types.Config) *Scanner {
	workers := config.Scan.Workers
	if workers <= 0 {
		workers = runtime.NumCPU()
	}

	// Determine repo path for git operations
	repoPath := "."
	if len(config.Scan.Paths) > 0 {
		repoPath = config.Scan.Paths[0]
	}

	return &Scanner{
		config:    config,
		workers:   workers,
		parser:    parser.NewParser(types.DefaultPatternConfig()),
		gitClient: git.NewClient(repoPath),
	}
}

// Scan 执行扫描
func (s *Scanner) Scan(ctx context.Context, path string) (*types.ScanResult, error) {
	startTime := time.Now()

	result := &types.ScanResult{
		Summary: types.Summary{
			Total:        0,
			FilesScanned: 0,
			ByType:       make(map[string]int),
			ByPriority:   make(map[string]int),
			ByAuthor:     make(map[string]int),
		},
		TODOs:    make([]types.TODO, 0),
		Warnings: make([]types.Warning, 0),
	}

	// 1. 创建Walker遍历目录
	walker := NewWalker(s.config)

	// 2. 获取文件通道
	fileCh, errCh := walker.Walk(path)

	// 3. 使用worker pool并行处理文件
	todoCh := make(chan []types.TODO, s.workers*2)
	var processWg sync.WaitGroup

	// 启动worker
	for i := 0; i < s.workers; i++ {
		processWg.Add(1)
		go func() {
			defer processWg.Done()
			for filePath := range fileCh {
				select {
				case <-ctx.Done():
					return
				default:
				}

				todos, err := s.processFile(filePath)
				if err != nil {
					// 记录警告但不中断扫描
					result.Warnings = append(result.Warnings, types.Warning{
						File:    filePath,
						Message: err.Error(),
						Type:    "scan_error",
					})
					continue
				}

				if len(todos) > 0 {
					todoCh <- todos
				}
			}
		}()
	}

	// 等待所有worker完成
	go func() {
		processWg.Wait()
		close(todoCh)
	}()

	// 收集结果
	filesScanned := 0
	for todos := range todoCh {
		for _, todo := range todos {
			result.TODOs = append(result.TODOs, todo)
			result.Summary.Total++
			result.Summary.ByType[todo.Type]++
			result.Summary.ByPriority[todo.Priority]++
			if todo.Author != "" {
				result.Summary.ByAuthor[todo.Author]++
			}
		}
		filesScanned++
	}

	// 检查错误
	select {
	case err := <-errCh:
		if err != nil {
			return result, err
		}
	default:
	}

	result.Summary.FilesScanned = filesScanned
	result.Summary.Duration = time.Since(startTime)

	return result, nil
}

// processFile 处理单个文件
func (s *Scanner) processFile(filePath string) ([]types.TODO, error) {
	// 读取文件内容
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// 使用parser解析TODO
	todos := s.parser.ParseFile(string(content), filePath)

	return todos, nil
}

// ScanStaged 扫描暂存文件
func (s *Scanner) ScanStaged(ctx context.Context) (*types.ScanResult, error) {
	startTime := time.Now()

	result := &types.ScanResult{
		Summary: types.Summary{
			Total:        0,
			FilesScanned: 0,
			ByType:       make(map[string]int),
			ByPriority:   make(map[string]int),
			ByAuthor:     make(map[string]int),
		},
		TODOs:    make([]types.TODO, 0),
		Warnings: make([]types.Warning, 0),
	}

	// Check if git is enabled
	if !s.config.Git.Enabled {
		return result, nil
	}

	// Get staged files from git
	files, err := s.gitClient.GetStagedFiles()
	if err != nil {
		return result, err
	}

	if len(files) == 0 {
		return result, nil
	}

	// Process each staged file
	for _, filePath := range files {
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		default:
		}

		// Check if file exists (may be deleted in staging)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			continue
		}

		todos, err := s.processFile(filePath)
		if err != nil {
			result.Warnings = append(result.Warnings, types.Warning{
				File:    filePath,
				Message: err.Error(),
				Type:    "scan_error",
			})
			continue
		}

		for _, todo := range todos {
			result.TODOs = append(result.TODOs, todo)
			result.Summary.Total++
			result.Summary.ByType[todo.Type]++
			result.Summary.ByPriority[todo.Priority]++
			if todo.Author != "" {
				result.Summary.ByAuthor[todo.Author]++
			}
		}
		result.Summary.FilesScanned++
	}

	result.Summary.Duration = time.Since(startTime)
	return result, nil
}

// ScanSince 扫描指定 commit 后的变更
func (s *Scanner) ScanSince(ctx context.Context, since string) (*types.ScanResult, error) {
	startTime := time.Now()

	result := &types.ScanResult{
		Summary: types.Summary{
			Total:        0,
			FilesScanned: 0,
			ByType:       make(map[string]int),
			ByPriority:   make(map[string]int),
			ByAuthor:     make(map[string]int),
		},
		TODOs:    make([]types.TODO, 0),
		Warnings: make([]types.Warning, 0),
	}

	// Check if git is enabled
	if !s.config.Git.Enabled {
		return result, nil
	}

	// Get changed files since the specified commit
	files, err := s.gitClient.GetChangedFiles(since)
	if err != nil {
		return result, err
	}

	if len(files) == 0 {
		return result, nil
	}

	// Process each changed file
	for _, filePath := range files {
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		default:
		}

		// Check if file exists (may be deleted)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			continue
		}

		todos, err := s.processFile(filePath)
		if err != nil {
			result.Warnings = append(result.Warnings, types.Warning{
				File:    filePath,
				Message: err.Error(),
				Type:    "scan_error",
			})
			continue
		}

		for _, todo := range todos {
			result.TODOs = append(result.TODOs, todo)
			result.Summary.Total++
			result.Summary.ByType[todo.Type]++
			result.Summary.ByPriority[todo.Priority]++
			if todo.Author != "" {
				result.Summary.ByAuthor[todo.Author]++
			}
		}
		result.Summary.FilesScanned++
	}

	result.Summary.Duration = time.Since(startTime)
	return result, nil
}

// ScanFile 扫描单个文件
func (s *Scanner) ScanFile(ctx context.Context, filePath string) ([]types.TODO, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	return s.processFile(filePath)
}

// ResultChan 返回结果通道，用于流式处理
func (s *Scanner) ResultChan(ctx context.Context, path string) <-chan ScanEvent {
	ch := make(chan ScanEvent, 100)

	go func() {
		defer close(ch)

		// 创建Walker
		walker := NewWalker(s.config)
		fileCh, errCh := walker.Walk(path)

		for filePath := range fileCh {
			select {
			case <-ctx.Done():
				ch <- ScanEvent{Type: EventError, Error: ctx.Err()}
				return
			default:
			}

			// 发送文件开始事件
			ch <- ScanEvent{Type: EventFileStart, File: filePath}

			// 处理文件
			todos, err := s.processFile(filePath)
			if err != nil {
				ch <- ScanEvent{Type: EventError, File: filePath, Error: err}
				continue
			}

			// 发送TODO发现事件
			for i := range todos {
				ch <- ScanEvent{Type: EventTODOFound, File: filePath, TODO: &todos[i]}
			}

			// 发送文件完成事件
			ch <- ScanEvent{Type: EventFileDone, File: filePath}
		}

		// 检查错误
		select {
		case err := <-errCh:
			if err != nil {
				ch <- ScanEvent{Type: EventError, Error: err}
			}
		default:
		}

		// 发送完成事件
		ch <- ScanEvent{Type: EventEventComplete}
	}()

	return ch
}

// ScanEvent 扫描事件
type ScanEvent struct {
	Type    EventType
	File    string
	TODO    *types.TODO
	Error   error
	Summary *types.Summary
}

// EventType 事件类型
type EventType int

const (
	// EventFileStart 文件开始扫描
	EventFileStart EventType = iota
	// EventTODOFound 发现 TODO
	EventTODOFound
	// EventFileDone 文件扫描完成
	EventFileDone
	// EventError 发生错误
	EventError
	// EventComplete 扫描完成
	EventEventComplete
)

// workerPool 工作池
type workerPool struct {
	workers int
	taskCh  chan scanTask
	resultCh chan scanResult
	wg      sync.WaitGroup
}

type scanTask struct {
	filePath string
	content  []byte
}

type scanResult struct {
	todos []types.TODO
	err   error
}
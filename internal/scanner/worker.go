// Package scanner 提供 TODO 扫描功能
package scanner

import (
	"context"
	"sync"
)

// WorkerPool 并行处理工作池
type WorkerPool struct {
	workerCount int
	taskCh      chan Task
	resultCh    chan Result
	wg          sync.WaitGroup
}

// Task 扫描任务
type Task struct {
	ID       int
	FilePath string
	Content  []byte
}

// Result 扫描结果
type Result struct {
	ID       int
	FilePath string
	TODOs    []TODOItem
	Error    error
}

// TODOItem 临时的 TODO 结构，用于内部传递
type TODOItem struct {
	Type     string
	Message  string
	Line     int
	Priority string
	Assignee string
}

// NewWorkerPool 创建新的工作池
func NewWorkerPool(workerCount int) *WorkerPool {
	if workerCount <= 0 {
		workerCount = 4 // 默认 4 个工作协程
	}

	return &WorkerPool{
		workerCount: workerCount,
		taskCh:      make(chan Task, 100),
		resultCh:    make(chan Result, 100),
	}
}

// Start 启动工作池
func (p *WorkerPool) Start(ctx context.Context, handler func(Task) Result) {
	for i := 0; i < p.workerCount; i++ {
		p.wg.Add(1)
		go p.worker(ctx, handler)
	}
}

// worker 工作协程
func (p *WorkerPool) worker(ctx context.Context, handler func(Task) Result) {
	defer p.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case task, ok := <-p.taskCh:
			if !ok {
				return
			}
			result := handler(task)
			p.resultCh <- result
		}
	}
}

// Submit 提交任务
func (p *WorkerPool) Submit(task Task) {
	p.taskCh <- task
}

// Results 返回结果通道
func (p *WorkerPool) Results() <-chan Result {
	return p.resultCh
}

// Stop 停止工作池
func (p *WorkerPool) Stop() {
	close(p.taskCh)
	p.wg.Wait()
	close(p.resultCh)
}

// ProcessFiles 并行处理文件
func (p *WorkerPool) ProcessFiles(ctx context.Context, files []string, processor func(string) ([]TODOItem, error)) []Result {
	results := make([]Result, 0, len(files))

	// 启动结果收集协程
	var mu sync.Mutex
	done := make(chan struct{})

	go func() {
		for result := range p.resultCh {
			mu.Lock()
			results = append(results, result)
			mu.Unlock()
		}
		close(done)
	}()

	// 提交任务
	for i, file := range files {
		p.Submit(Task{
			ID:       i,
			FilePath: file,
		})
	}

	// 等待所有任务完成
	p.Stop()
	<-done

	return results
}
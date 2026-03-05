// Package cache 提供缓存功能，用于增量扫描
package cache

import (
	"time"

	"github.com/mxihan/todo-tracker/pkg/types"
)

// Cache 缓存接口
type Cache interface {
	// 文件操作
	GetFileHash(path string) (string, bool)
	SetFileHash(path, hash string) error
	GetFileRecord(path string) (*types.FileRecord, bool)
	SetFileRecord(record *types.FileRecord) error

	// TODO操作
	GetTODOs(filePath string) ([]types.TODO, bool)
	SetTODOs(filePath string, todos []types.TODO) error
	GetTODO(id string) (*types.TODO, bool)
	UpdateTODO(todo *types.TODO) error
	DeleteTODO(id string) error

	// 作者操作
	GetAuthor(name string) (*types.Author, bool)
	SetAuthor(author *types.Author) error
	GetAllAuthors() ([]types.Author, error)

	// 扫描历史
	AddScanHistory(filesScanned, todosFound int, duration time.Duration) error
	GetScanHistory(limit int) ([]ScanHistory, error)

	// 通用操作
	Close() error
	Clear() error
}

// ScanHistory 扫描历史记录
type ScanHistory struct {
	ID          int
	Timestamp   time.Time
	FilesScanned int
	TODOsFound  int
	Duration    time.Duration
}

// Options 缓存选项
type Options struct {
	// 缓存文件路径
	Path string
	// 是否启用缓存
	Enabled bool
	// 缓存过期时间（秒）
	TTL int
}

// DefaultOptions 返回默认缓存选项
func DefaultOptions() *Options {
	return &Options{
		Path:    ".todo-cache.db",
		Enabled: true,
		TTL:     86400, // 24小时
	}
}
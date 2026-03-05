// Package cache 提供 SQLite 缓存实现
package cache

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/mxihan/todo-tracker/pkg/types"
	_ "github.com/mattn/go-sqlite3"
)

// SQLiteCache SQLite缓存实现
type SQLiteCache struct {
	db   *sql.DB
	path string
	mu   sync.RWMutex
}

// NewSQLiteCache 创建新的SQLite缓存
func NewSQLiteCache(opts *Options) (*SQLiteCache, error) {
	if opts == nil {
		opts = DefaultOptions()
	}

	// 确保目录存在
	dir := filepath.Dir(opts.Path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("创建缓存目录失败: %w", err)
		}
	}

	// 打开数据库
	db, err := sql.Open("sqlite3", opts.Path)
	if err != nil {
		return nil, fmt.Errorf("打开数据库失败: %w", err)
	}

	cache := &SQLiteCache{
		db:   db,
		path: opts.Path,
	}

	// 初始化表结构
	if err := cache.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("初始化数据库失败: %w", err)
	}

	return cache, nil
}

// initSchema 初始化数据库表结构
func (c *SQLiteCache) initSchema() error {
	// 文件表
	_, err := c.db.Exec(`
		CREATE TABLE IF NOT EXISTS files (
			path TEXT PRIMARY KEY,
			hash TEXT NOT NULL,
			last_scanned TIMESTAMP,
			size_bytes INTEGER,
			churn_count INTEGER DEFAULT 0
		)
	`)
	if err != nil {
		return err
	}

	// TODO表
	_, err = c.db.Exec(`
		CREATE TABLE IF NOT EXISTS todos (
			id TEXT PRIMARY KEY,
			file_path TEXT,
			line_start INTEGER,
			line_end INTEGER,
			type TEXT,
			message TEXT,
			priority TEXT DEFAULT 'low',
			assignee TEXT,
			ticket_ref TEXT,
			status TEXT DEFAULT 'open',
			git_author TEXT,
			git_commit TEXT,
			git_date TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(file_path, line_start)
		)
	`)
	if err != nil {
		return err
	}

	// 索引
	_, err = c.db.Exec(`CREATE INDEX IF NOT EXISTS idx_todos_status ON todos(status)`)
	if err != nil {
		return err
	}

	_, err = c.db.Exec(`CREATE INDEX IF NOT EXISTS idx_todos_git_date ON todos(git_date)`)
	if err != nil {
		return err
	}

	_, err = c.db.Exec(`CREATE INDEX IF NOT EXISTS idx_todos_git_author ON todos(git_author)`)
	if err != nil {
		return err
	}

	// 扫描历史表
	_, err = c.db.Exec(`
		CREATE TABLE IF NOT EXISTS scan_history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			files_scanned INTEGER,
			todos_found INTEGER,
			duration_ms INTEGER
		)
	`)
	if err != nil {
		return err
	}

	// 作者表
	_, err = c.db.Exec(`
		CREATE TABLE IF NOT EXISTS authors (
			name TEXT PRIMARY KEY,
			last_commit TIMESTAMP,
			commit_count INTEGER,
			is_active BOOLEAN DEFAULT true
		)
	`)

	return err
}

// GetFileHash 获取文件哈希
func (c *SQLiteCache) GetFileHash(path string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var hash string
	err := c.db.QueryRow("SELECT hash FROM files WHERE path = ?", path).Scan(&hash)
	if err != nil {
		return "", false
	}
	return hash, true
}

// SetFileHash 设置文件哈希
func (c *SQLiteCache) SetFileHash(path, hash string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, err := c.db.Exec(`
		INSERT OR REPLACE INTO files (path, hash, last_scanned)
		VALUES (?, ?, ?)
	`, path, hash, time.Now())

	return err
}

// GetFileRecord 获取文件记录
func (c *SQLiteCache) GetFileRecord(path string) (*types.FileRecord, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	record := &types.FileRecord{}
	err := c.db.QueryRow(`
		SELECT path, hash, last_scanned, size_bytes, churn_count
		FROM files WHERE path = ?
	`, path).Scan(&record.Path, &record.Hash, &record.LastScanned, &record.SizeBytes, &record.ChurnCount)

	if err != nil {
		return nil, false
	}
	return record, true
}

// SetFileRecord 设置文件记录
func (c *SQLiteCache) SetFileRecord(record *types.FileRecord) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, err := c.db.Exec(`
		INSERT OR REPLACE INTO files (path, hash, last_scanned, size_bytes, churn_count)
		VALUES (?, ?, ?, ?, ?)
	`, record.Path, record.Hash, record.LastScanned, record.SizeBytes, record.ChurnCount)

	return err
}

// GetTODOs 获取文件的所有TODO
func (c *SQLiteCache) GetTODOs(filePath string) ([]types.TODO, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	rows, err := c.db.Query(`
		SELECT id, file_path, line_start, line_end, type, message, priority,
			   assignee, ticket_ref, status, git_author, git_commit, git_date
		FROM todos WHERE file_path = ?
	`, filePath)
	if err != nil {
		return nil, false
	}
	defer rows.Close()

	var todos []types.TODO
	for rows.Next() {
		var todo types.TODO
		var gitDate sql.NullTime
		err := rows.Scan(
			&todo.ID, &todo.File, &todo.Line, &todo.LineEnd, &todo.Type,
			&todo.Message, &todo.Priority, &todo.Assignee, &todo.TicketRef,
			&todo.Status, &todo.Author, &todo.CommitHash, &gitDate,
		)
		if err != nil {
			continue
		}
		if gitDate.Valid {
			todo.CreatedAt = gitDate.Time
		}
		todos = append(todos, todo)
	}

	return todos, len(todos) > 0
}

// SetTODOs 设置文件的所有TODO
func (c *SQLiteCache) SetTODOs(filePath string, todos []types.TODO) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 开始事务
	tx, err := c.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 删除旧的TODO
	_, err = tx.Exec("DELETE FROM todos WHERE file_path = ?", filePath)
	if err != nil {
		return err
	}

	// 插入新的TODO
	stmt, err := tx.Prepare(`
		INSERT INTO todos (id, file_path, line_start, line_end, type, message,
						   priority, assignee, ticket_ref, status, git_author,
						   git_commit, git_date)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, todo := range todos {
		_, err = stmt.Exec(
			todo.ID, todo.File, todo.Line, todo.LineEnd, todo.Type,
			todo.Message, todo.Priority, todo.Assignee, todo.TicketRef,
			todo.Status, todo.Author, todo.CommitHash, todo.CreatedAt,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetTODO 获取单个TODO
func (c *SQLiteCache) GetTODO(id string) (*types.TODO, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	todo := &types.TODO{}
	var gitDate sql.NullTime
	err := c.db.QueryRow(`
		SELECT id, file_path, line_start, line_end, type, message, priority,
			   assignee, ticket_ref, status, git_author, git_commit, git_date
		FROM todos WHERE id = ?
	`, id).Scan(
		&todo.ID, &todo.File, &todo.Line, &todo.LineEnd, &todo.Type,
		&todo.Message, &todo.Priority, &todo.Assignee, &todo.TicketRef,
		&todo.Status, &todo.Author, &todo.CommitHash, &gitDate,
	)

	if err != nil {
		return nil, false
	}
	if gitDate.Valid {
		todo.CreatedAt = gitDate.Time
	}
	return todo, true
}

// UpdateTODO 更新TODO
func (c *SQLiteCache) UpdateTODO(todo *types.TODO) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, err := c.db.Exec(`
		UPDATE todos SET
			line_end = ?, type = ?, message = ?, priority = ?,
			assignee = ?, ticket_ref = ?, status = ?, git_author = ?,
			git_commit = ?, git_date = ?, updated_at = ?
		WHERE id = ?
	`, todo.LineEnd, todo.Type, todo.Message, todo.Priority,
		todo.Assignee, todo.TicketRef, todo.Status, todo.Author,
		todo.CommitHash, todo.CreatedAt, time.Now(), todo.ID)

	return err
}

// DeleteTODO 删除TODO
func (c *SQLiteCache) DeleteTODO(id string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, err := c.db.Exec("DELETE FROM todos WHERE id = ?", id)
	return err
}

// GetAuthor 获取作者信息
func (c *SQLiteCache) GetAuthor(name string) (*types.Author, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	author := &types.Author{}
	var lastCommit sql.NullTime
	err := c.db.QueryRow(`
		SELECT name, last_commit, commit_count, is_active
		FROM authors WHERE name = ?
	`, name).Scan(&author.Name, &lastCommit, &author.CommitCount, &author.IsActive)

	if err != nil {
		return nil, false
	}
	if lastCommit.Valid {
		author.LastCommit = lastCommit.Time
	}
	return author, true
}

// SetAuthor 设置作者信息
func (c *SQLiteCache) SetAuthor(author *types.Author) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, err := c.db.Exec(`
		INSERT OR REPLACE INTO authors (name, last_commit, commit_count, is_active)
		VALUES (?, ?, ?, ?)
	`, author.Name, author.LastCommit, author.CommitCount, author.IsActive)

	return err
}

// GetAllAuthors 获取所有作者
func (c *SQLiteCache) GetAllAuthors() ([]types.Author, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	rows, err := c.db.Query("SELECT name, last_commit, commit_count, is_active FROM authors")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var authors []types.Author
	for rows.Next() {
		var author types.Author
		var lastCommit sql.NullTime
		err := rows.Scan(&author.Name, &lastCommit, &author.CommitCount, &author.IsActive)
		if err != nil {
			continue
		}
		if lastCommit.Valid {
			author.LastCommit = lastCommit.Time
		}
		authors = append(authors, author)
	}

	return authors, nil
}

// AddScanHistory 添加扫描历史
func (c *SQLiteCache) AddScanHistory(filesScanned, todosFound int, duration time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, err := c.db.Exec(`
		INSERT INTO scan_history (timestamp, files_scanned, todos_found, duration_ms)
		VALUES (?, ?, ?, ?)
	`, time.Now(), filesScanned, todosFound, duration.Milliseconds())

	return err
}

// GetScanHistory 获取扫描历史
func (c *SQLiteCache) GetScanHistory(limit int) ([]ScanHistory, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	rows, err := c.db.Query(`
		SELECT id, timestamp, files_scanned, todos_found, duration_ms
		FROM scan_history ORDER BY timestamp DESC LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []ScanHistory
	for rows.Next() {
		var h ScanHistory
		var durationMs int64
		err := rows.Scan(&h.ID, &h.Timestamp, &h.FilesScanned, &h.TODOsFound, &durationMs)
		if err != nil {
			continue
		}
		h.Duration = time.Duration(durationMs) * time.Millisecond
		history = append(history, h)
	}

	return history, nil
}

// Close 关闭缓存
func (c *SQLiteCache) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.db.Close()
}

// Clear 清空缓存
func (c *SQLiteCache) Clear() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, err := c.db.Exec("DELETE FROM files")
	if err != nil {
		return err
	}

	_, err = c.db.Exec("DELETE FROM todos")
	if err != nil {
		return err
	}

	_, err = c.db.Exec("DELETE FROM scan_history")
	if err != nil {
		return err
	}

	_, err = c.db.Exec("DELETE FROM authors")
	return err
}

// GetStats 获取缓存统计信息
func (c *SQLiteCache) GetStats() (map[string]int64, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := make(map[string]int64)
	var count int64

	// 文件数量
	err := c.db.QueryRow("SELECT COUNT(*) FROM files").Scan(&count)
	if err != nil {
		return nil, err
	}
	stats["files"] = count

	// TODO数量
	err = c.db.QueryRow("SELECT COUNT(*) FROM todos").Scan(&count)
	if err != nil {
		return nil, err
	}
	stats["todos"] = count

	// 作者数量
	err = c.db.QueryRow("SELECT COUNT(*) FROM authors").Scan(&count)
	if err != nil {
		return nil, err
	}
	stats["authors"] = count

	// 扫描次数
	err = c.db.QueryRow("SELECT COUNT(*) FROM scan_history").Scan(&count)
	if err != nil {
		return nil, err
	}
	stats["scans"] = count

	return stats, nil
}
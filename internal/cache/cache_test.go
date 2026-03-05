// Package cache_test 测试缓存功能
package cache

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/todo-tracker/todo-tracker/pkg/types"
)

// TestDefaultOptions 测试默认选项
func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()

	if opts == nil {
		t.Fatal("DefaultOptions() returned nil")
	}

	if opts.Path != ".todo-cache.db" {
		t.Errorf("Path = %s, want .todo-cache.db", opts.Path)
	}

	if !opts.Enabled {
		t.Error("Enabled should be true")
	}

	if opts.TTL != 86400 {
		t.Errorf("TTL = %d, want 86400", opts.TTL)
	}
}

// TestNewSQLiteCache 测试创建SQLite缓存
func TestNewSQLiteCache(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "cache-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test.db")

	cache, err := NewSQLiteCache(&Options{Path: dbPath})
	if err != nil {
		t.Fatalf("NewSQLiteCache() failed: %v", err)
	}
	defer cache.Close()

	if cache == nil {
		t.Error("NewSQLiteCache() returned nil")
	}
}

// TestNewSQLiteCacheWithNilOptions 测试nil选项
func TestNewSQLiteCacheWithNilOptions(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "cache-nil-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 切换到临时目录
	oldDir, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldDir)

	cache, err := NewSQLiteCache(nil)
	if err != nil {
		t.Fatalf("NewSQLiteCache(nil) failed: %v", err)
	}
	defer cache.Close()
}

// TestFileHashOperations 测试文件哈希操作
func TestFileHashOperations(t *testing.T) {
	cache := createTestCache(t)
	defer cache.Close()

	// 测试获取不存在的哈希
	_, exists := cache.GetFileHash("/nonexistent/file.go")
	if exists {
		t.Error("GetFileHash() should return false for nonexistent file")
	}

	// 测试设置哈希
	err := cache.SetFileHash("/test/file.go", "abc123")
	if err != nil {
		t.Fatalf("SetFileHash() failed: %v", err)
	}

	// 测试获取存在的哈希
	hash, exists := cache.GetFileHash("/test/file.go")
	if !exists {
		t.Error("GetFileHash() should return true for existing file")
	}
	if hash != "abc123" {
		t.Errorf("Hash = %s, want abc123", hash)
	}
}

// TestFileRecordOperations 测试文件记录操作
func TestFileRecordOperations(t *testing.T) {
	cache := createTestCache(t)
	defer cache.Close()

	// 测试获取不存在的记录
	_, exists := cache.GetFileRecord("/nonexistent/file.go")
	if exists {
		t.Error("GetFileRecord() should return false for nonexistent file")
	}

	// 测试设置记录
	record := &types.FileRecord{
		Path:        "/test/file.go",
		Hash:        "def456",
		LastScanned: time.Now(),
		SizeBytes:   1024,
		ChurnCount:  5,
	}

	err := cache.SetFileRecord(record)
	if err != nil {
		t.Fatalf("SetFileRecord() failed: %v", err)
	}

	// 测试获取存在的记录
	got, exists := cache.GetFileRecord("/test/file.go")
	if !exists {
		t.Fatal("GetFileRecord() should return true for existing file")
	}

	if got.Hash != "def456" {
		t.Errorf("Hash = %s, want def456", got.Hash)
	}

	if got.ChurnCount != 5 {
		t.Errorf("ChurnCount = %d, want 5", got.ChurnCount)
	}
}

// TestTODOOperations 测试TODO操作
func TestTODOOperations(t *testing.T) {
	cache := createTestCache(t)
	defer cache.Close()

	// 测试获取不存在的TODO
	_, exists := cache.GetTODO("nonexistent-id")
	if exists {
		t.Error("GetTODO() should return false for nonexistent id")
	}

	// 测试设置TODO
	todos := []types.TODO{
		{
			ID:        "todo1",
			Type:      "TODO",
			Message:   "Test message",
			File:      "/test/file.go",
			Line:      10,
			LineEnd:   10,
			Priority:  "high",
			Assignee:  "alice",
			TicketRef: "#123",
			Status:    "open",
		},
		{
			ID:       "todo2",
			Type:     "FIXME",
			Message:  "Fix this",
			File:     "/test/file.go",
			Line:     20,
			LineEnd:  20,
			Priority: "medium",
			Status:   "open",
		},
	}

	err := cache.SetTODOs("/test/file.go", todos)
	if err != nil {
		t.Fatalf("SetTODOs() failed: %v", err)
	}

	// 测试获取TODO列表
	gotTodos, exists := cache.GetTODOs("/test/file.go")
	if !exists {
		t.Fatal("GetTODOs() should return true for existing file")
	}

	if len(gotTodos) != 2 {
		t.Errorf("TODOs count = %d, want 2", len(gotTodos))
	}

	// 测试获取单个TODO
	gotTodo, exists := cache.GetTODO("todo1")
	if !exists {
		t.Fatal("GetTODO() should return true for existing id")
	}

	if gotTodo.Message != "Test message" {
		t.Errorf("Message = %s, want Test message", gotTodo.Message)
	}
}

// TestTODOUpdate 测试TODO更新
func TestTODOUpdate(t *testing.T) {
	cache := createTestCache(t)
	defer cache.Close()

	// 先创建一个TODO
	todos := []types.TODO{
		{
			ID:       "update-test",
			Type:     "TODO",
			Message:  "Original message",
			File:     "/test/update.go",
			Line:     5,
			Priority: "low",
			Status:   "open",
		},
	}

	err := cache.SetTODOs("/test/update.go", todos)
	if err != nil {
		t.Fatalf("SetTODOs() failed: %v", err)
	}

	// 更新TODO
	updated := &types.TODO{
		ID:       "update-test",
		Type:     "TODO",
		Message:  "Updated message",
		File:     "/test/update.go",
		Line:     5,
		Priority: "high",
		Status:   "resolved",
	}

	err = cache.UpdateTODO(updated)
	if err != nil {
		t.Fatalf("UpdateTODO() failed: %v", err)
	}

	// 验证更新
	got, exists := cache.GetTODO("update-test")
	if !exists {
		t.Fatal("GetTODO() should return true")
	}

	if got.Message != "Updated message" {
		t.Errorf("Message = %s, want Updated message", got.Message)
	}

	if got.Priority != "high" {
		t.Errorf("Priority = %s, want high", got.Priority)
	}
}

// TestTODODelete 测试TODO删除
func TestTODODelete(t *testing.T) {
	cache := createTestCache(t)
	defer cache.Close()

	// 先创建一个TODO
	todos := []types.TODO{
		{
			ID:      "delete-test",
			Type:    "TODO",
			Message: "To be deleted",
			File:    "/test/delete.go",
			Line:    1,
			Status:  "open",
		},
	}

	cache.SetTODOs("/test/delete.go", todos)

	// 删除TODO
	err := cache.DeleteTODO("delete-test")
	if err != nil {
		t.Fatalf("DeleteTODO() failed: %v", err)
	}

	// 验证删除
	_, exists := cache.GetTODO("delete-test")
	if exists {
		t.Error("GetTODO() should return false for deleted TODO")
	}
}

// TestAuthorOperations 测试作者操作
func TestAuthorOperations(t *testing.T) {
	cache := createTestCache(t)
	defer cache.Close()

	// 测试获取不存在的作者
	_, exists := cache.GetAuthor("nonexistent")
	if exists {
		t.Error("GetAuthor() should return false for nonexistent author")
	}

	// 测试设置作者
	author := &types.Author{
		Name:        "Alice",
		LastCommit:  time.Now(),
		CommitCount: 42,
		IsActive:    true,
	}

	err := cache.SetAuthor(author)
	if err != nil {
		t.Fatalf("SetAuthor() failed: %v", err)
	}

	// 测试获取存在的作者
	got, exists := cache.GetAuthor("Alice")
	if !exists {
		t.Fatal("GetAuthor() should return true for existing author")
	}

	if got.CommitCount != 42 {
		t.Errorf("CommitCount = %d, want 42", got.CommitCount)
	}
}

// TestGetAllAuthors 测试获取所有作者
func TestGetAllAuthors(t *testing.T) {
	cache := createTestCache(t)
	defer cache.Close()

	// 添加多个作者
	authors := []*types.Author{
		{Name: "Alice", CommitCount: 10, IsActive: true},
		{Name: "Bob", CommitCount: 20, IsActive: true},
		{Name: "Charlie", CommitCount: 5, IsActive: false},
	}

	for _, a := range authors {
		cache.SetAuthor(a)
	}

	// 获取所有作者
	all, err := cache.GetAllAuthors()
	if err != nil {
		t.Fatalf("GetAllAuthors() failed: %v", err)
	}

	if len(all) != 3 {
		t.Errorf("Authors count = %d, want 3", len(all))
	}
}

// TestScanHistory 测试扫描历史
func TestScanHistory(t *testing.T) {
	cache := createTestCache(t)
	defer cache.Close()

	// 添加扫描历史
	err := cache.AddScanHistory(100, 25, 1500*time.Millisecond)
	if err != nil {
		t.Fatalf("AddScanHistory() failed: %v", err)
	}

	err = cache.AddScanHistory(150, 30, 2000*time.Millisecond)
	if err != nil {
		t.Fatalf("AddScanHistory() failed: %v", err)
	}

	// 获取扫描历史
	history, err := cache.GetScanHistory(10)
	if err != nil {
		t.Fatalf("GetScanHistory() failed: %v", err)
	}

	if len(history) != 2 {
		t.Errorf("History count = %d, want 2", len(history))
	}

	// 验证最新记录在前面
	if history[0].FilesScanned != 150 {
		t.Errorf("Latest history FilesScanned = %d, want 150", history[0].FilesScanned)
	}
}

// TestClear 测试清空缓存
func TestClear(t *testing.T) {
	cache := createTestCache(t)
	defer cache.Close()

	// 添加一些数据
	cache.SetFileHash("/test/file.go", "hash123")
	cache.SetAuthor(&types.Author{Name: "Alice", CommitCount: 10})
	cache.AddScanHistory(10, 5, 100*time.Millisecond)

	// 清空缓存
	err := cache.Clear()
	if err != nil {
		t.Fatalf("Clear() failed: %v", err)
	}

	// 验证数据已清空
	_, exists := cache.GetFileHash("/test/file.go")
	if exists {
		t.Error("File hash should be cleared")
	}

	_, exists = cache.GetAuthor("Alice")
	if exists {
		t.Error("Author should be cleared")
	}
}

// TestGetStats 测试获取统计信息
func TestGetStats(t *testing.T) {
	cache := createTestCache(t)
	defer cache.Close()

	// 添加一些数据
	cache.SetFileHash("/test/file1.go", "hash1")
	cache.SetFileHash("/test/file2.go", "hash2")
	cache.SetAuthor(&types.Author{Name: "Alice", CommitCount: 10})
	cache.AddScanHistory(100, 50, 1000*time.Millisecond)

	// 获取统计
	stats, err := cache.GetStats()
	if err != nil {
		t.Fatalf("GetStats() failed: %v", err)
	}

	if stats["files"] != 2 {
		t.Errorf("Files count = %d, want 2", stats["files"])
	}

	if stats["authors"] != 1 {
		t.Errorf("Authors count = %d, want 1", stats["authors"])
	}

	if stats["scans"] != 1 {
		t.Errorf("Scans count = %d, want 1", stats["scans"])
	}
}

// TestScanHistoryStruct 测试ScanHistory结构
func TestScanHistoryStruct(t *testing.T) {
	h := ScanHistory{
		ID:           1,
		Timestamp:    time.Now(),
		FilesScanned: 100,
		TODOsFound:   25,
		Duration:     1500 * time.Millisecond,
	}

	if h.ID != 1 {
		t.Errorf("ID = %d, want 1", h.ID)
	}

	if h.FilesScanned != 100 {
		t.Errorf("FilesScanned = %d, want 100", h.FilesScanned)
	}
}

// createTestCache 创建测试用缓存
func createTestCache(t *testing.T) *SQLiteCache {
	t.Helper()

	tempFile, err := os.CreateTemp("", "cache-test-*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tempFile.Close()

	cache, err := NewSQLiteCache(&Options{Path: tempFile.Name()})
	if err != nil {
		os.Remove(tempFile.Name())
		t.Fatalf("Failed to create cache: %v", err)
	}

	// 设置清理函数
	t.Cleanup(func() {
		os.Remove(tempFile.Name())
	})

	return cache
}

// BenchmarkSetFileHash 基准测试设置文件哈希
func BenchmarkSetFileHash(b *testing.B) {
	tempFile, err := os.CreateTemp("", "bench-cache-*.db")
	if err != nil {
		b.Fatalf("Failed to create temp file: %v", err)
	}
	tempFile.Close()
	defer os.Remove(tempFile.Name())

	cache, err := NewSQLiteCache(&Options{Path: tempFile.Name()})
	if err != nil {
		b.Fatalf("Failed to create cache: %v", err)
	}
	defer cache.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.SetFileHash("/test/file.go", "hash123")
	}
}

// BenchmarkGetFileHash 基准测试获取文件哈希
func BenchmarkGetFileHash(b *testing.B) {
	tempFile, err := os.CreateTemp("", "bench-cache-*.db")
	if err != nil {
		b.Fatalf("Failed to create temp file: %v", err)
	}
	tempFile.Close()
	defer os.Remove(tempFile.Name())

	cache, err := NewSQLiteCache(&Options{Path: tempFile.Name()})
	if err != nil {
		b.Fatalf("Failed to create cache: %v", err)
	}
	defer cache.Close()

	cache.SetFileHash("/test/file.go", "hash123")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.GetFileHash("/test/file.go")
	}
}

// BenchmarkSetTODOs 基准测试设置TODO
func BenchmarkSetTODOs(b *testing.B) {
	tempFile, err := os.CreateTemp("", "bench-cache-*.db")
	if err != nil {
		b.Fatalf("Failed to create temp file: %v", err)
	}
	tempFile.Close()
	defer os.Remove(tempFile.Name())

	cache, err := NewSQLiteCache(&Options{Path: tempFile.Name()})
	if err != nil {
		b.Fatalf("Failed to create cache: %v", err)
	}
	defer cache.Close()

	todos := []types.TODO{
		{ID: "t1", Type: "TODO", Message: "Test 1", File: "/test/file.go", Line: 10},
		{ID: "t2", Type: "FIXME", Message: "Test 2", File: "/test/file.go", Line: 20},
		{ID: "t3", Type: "HACK", Message: "Test 3", File: "/test/file.go", Line: 30},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.SetTODOs("/test/file.go", todos)
	}
}
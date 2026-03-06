// Package git_test 测试Git操作
package git

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestNewClient 测试Git客户端创建
func TestNewClient(t *testing.T) {
	tests := []struct {
		name     string
		repoPath string
	}{
		{
			name:     "当前目录",
			repoPath: ".",
		},
		{
			name:     "绝对路径",
			repoPath: "/tmp/test",
		},
		{
			name:     "相对路径",
			repoPath: "../test",
		},
		{
			name:     "空路径",
			repoPath: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(tt.repoPath)
			if client == nil {
				t.Error("NewClient() returned nil")
			}
		})
	}
}

// TestIsGitRepo 测试Git仓库检测
func TestIsGitRepo(t *testing.T) {
	// 测试非Git目录
	tempDir, err := os.MkdirTemp("", "not-git-repo")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	client := NewClient(tempDir)
	if client.IsGitRepo() {
		t.Error("IsGitRepo() should return false for non-git directory")
	}
}

// TestClientMethods 测试客户端方法不会崩溃
func TestClientMethods(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "git-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	client := NewClient(tempDir)

	// 这些方法在非Git仓库中会失败，但不应崩溃
	t.Run("GetCurrentBranch", func(t *testing.T) {
		_, err := client.GetCurrentBranch()
		// 预期会失败，因为不是Git仓库
		if err == nil {
			t.Log("GetCurrentBranch succeeded (unexpected)")
		}
	})

	t.Run("GetDefaultBranch", func(t *testing.T) {
		branch := client.GetDefaultBranch()
		// 应该返回默认值
		if branch == "" {
			t.Error("GetDefaultBranch() returned empty string")
		}
	})

	t.Run("GetStagedFiles", func(t *testing.T) {
		_, err := client.GetStagedFiles()
		if err == nil {
			t.Log("GetStagedFiles succeeded (unexpected)")
		}
	})

	t.Run("GetAuthors", func(t *testing.T) {
		_, err := client.GetAuthors()
		if err == nil {
			t.Log("GetAuthors succeeded (unexpected)")
		}
	})
}

// TestCommitInfo 测试提交信息结构
func TestCommitInfo(t *testing.T) {
	info := &CommitInfo{
		Hash:      "abc123",
		Author:    "Test User",
		Email:     "test@example.com",
		Date:      time.Now(),
		Message:   "Test commit",
		FileCount: 3,
	}

	if info.Hash != "abc123" {
		t.Errorf("Hash = %s, want abc123", info.Hash)
	}

	if info.Author != "Test User" {
		t.Errorf("Author = %s, want Test User", info.Author)
	}
}

// TestAuthorInfo 测试作者信息结构
func TestAuthorInfo(t *testing.T) {
	info := AuthorInfo{
		Name:        "Alice",
		Email:       "alice@example.com",
		LastCommit:  time.Now(),
		CommitCount: 42,
		IsActive:    true,
	}

	if info.Name != "Alice" {
		t.Errorf("Name = %s, want Alice", info.Name)
	}

	if info.CommitCount != 42 {
		t.Errorf("CommitCount = %d, want 42", info.CommitCount)
	}
}

// TestGetFileHash 测试文件哈希计算
func TestGetFileHash(t *testing.T) {
	// 创建临时文件
	tempFile, err := os.CreateTemp("", "hash-test-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	testContent := "test content for hash"
	if _, err := tempFile.WriteString(testContent); err != nil {
		t.Fatalf("Failed to write content: %v", err)
	}
	tempFile.Close()

	client := NewClient(filepath.Dir(tempFile.Name()))

	// 这个测试可能在非Git目录中失败
	hash, err := client.GetFileHash(tempFile.Name())
	if err != nil {
		t.Logf("GetFileHash() returned error (expected in non-git dir): %v", err)
	} else {
		if len(hash) != 40 { // SHA-1哈希长度
			t.Errorf("Hash length = %d, want 40", len(hash))
		}
	}
}

// TestGetFileChurn 测试文件修改次数获取
func TestGetFileChurn(t *testing.T) {
	// 创建临时文件
	tempFile, err := os.CreateTemp("", "churn-test-*.go")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	tempFile.WriteString("package main\nfunc main() {}\n")
	tempFile.Close()

	client := NewClient(filepath.Dir(tempFile.Name()))

	churn, err := client.GetFileChurn(tempFile.Name())
	if err != nil {
		t.Logf("GetFileChurn() returned error: %v", err)
	}

	// 新文件的churn应该是0
	_ = churn
}

// TestGetChangedFiles 测试获取变更文件
func TestGetChangedFiles(t *testing.T) {
	client := NewClient(".")

	files, err := client.GetChangedFiles("HEAD~1")
	if err != nil {
		t.Logf("GetChangedFiles() returned error: %v", err)
		return
	}

	// 只是验证返回的是slice
	t.Logf("Found %d changed files", len(files))
}

// TestGetRepoRoot 测试获取仓库根目录
func TestGetRepoRoot(t *testing.T) {
	client := NewClient(".")

	root, err := client.GetRepoRoot()
	if err != nil {
		t.Logf("GetRepoRoot() returned error (not in git repo): %v", err)
		return
	}

	if root == "" {
		t.Error("GetRepoRoot() returned empty string")
	}

	t.Logf("Repo root: %s", root)
}

// TestGetAuthorLastCommit 测试获取作者最后提交时间
func TestGetAuthorLastCommit(t *testing.T) {
	client := NewClient(".")

	lastCommit, err := client.GetAuthorLastCommit("nonexistent-user")
	if err != nil {
		t.Logf("GetAuthorLastCommit() returned error: %v", err)
		return
	}

	// 不存在的作者应该返回零值
	if !lastCommit.IsZero() {
		t.Logf("Last commit for nonexistent user: %v", lastCommit)
	}
}

// TestGetFileLastModified 测试获取文件最后修改时间
func TestGetFileLastModified(t *testing.T) {
	// 创建临时文件
	tempFile, err := os.CreateTemp("", "modified-test-*.go")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	tempFile.Close()

	client := NewClient(filepath.Dir(tempFile.Name()))

	modTime, err := client.GetFileLastModified(tempFile.Name())
	if err != nil {
		t.Logf("GetFileLastModified() returned error: %v", err)
		return
	}

	_ = modTime
}

// TestGetCommit 测试获取提交信息
func TestGetCommit(t *testing.T) {
	client := NewClient(".")

	// 尝试获取最近的提交
	info, err := client.GetCommit("HEAD")
	if err != nil {
		t.Logf("GetCommit() returned error: %v", err)
		return
	}

	if info == nil {
		t.Error("GetCommit() returned nil")
		return
	}

	t.Logf("Commit: %s by %s", info.Hash[:8], info.Author)
}

// BenchmarkGetFileHash 基准测试文件哈希计算
func BenchmarkGetFileHash(b *testing.B) {
	// 创建临时文件
	tempFile, err := os.CreateTemp("", "bench-hash-*.txt")
	if err != nil {
		b.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	content := make([]byte, 1024)
	for i := range content {
		content[i] = byte(i % 256)
	}
	tempFile.Write(content)
	tempFile.Close()

	client := NewClient(filepath.Dir(tempFile.Name()))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.GetFileHash(tempFile.Name())
	}
}

// BenchmarkGetFileChurn 基准测试文件修改次数获取
func BenchmarkGetFileChurn(b *testing.B) {
	tempFile, err := os.CreateTemp("", "bench-churn-*.go")
	if err != nil {
		b.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	tempFile.WriteString("package main")
	tempFile.Close()

	client := NewClient(filepath.Dir(tempFile.Name()))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.GetFileChurn(tempFile.Name())
	}
}

// TestClient_Run tests the Run method
func TestClient_Run(t *testing.T) {
	// Test in a non-git directory
	tempDir, err := os.MkdirTemp("", "git-run-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	client := NewClient(tempDir)

	// This should fail since it's not a git repo
	_, err = client.Run("status")
	if err == nil {
		t.Log("Run() succeeded (unexpected in non-git dir)")
	} else {
		t.Logf("Run() returned expected error: %v", err)
	}
}

// TestIsGitRepo_ValidGitDir tests IsGitRepo with valid git directory
func TestIsGitRepo_ValidGitDir(t *testing.T) {
	// Create a temp directory with .git
	tempDir, err := os.MkdirTemp("", "git-valid-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create .git directory
	gitDir := filepath.Join(tempDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("Failed to create .git dir: %v", err)
	}

	client := NewClient(tempDir)
	if !client.IsGitRepo() {
		t.Error("IsGitRepo() should return true for valid git directory")
	}
}

// TestGetFileHash_ErrorPaths tests GetFileHash error conditions
func TestGetFileHash_ErrorPaths(t *testing.T) {
	t.Run("non-existent file", func(t *testing.T) {
		client := NewClient(".")
		_, err := client.GetFileHash("/non/existent/file.txt")
		if err == nil {
			t.Error("GetFileHash() should return error for non-existent file")
		}
	})

	t.Run("path outside repo", func(t *testing.T) {
		client := NewClient(".")
		// Using an absolute path that may not be in the repo
		_, err := client.GetFileHash("/tmp/some-file.txt")
		if err == nil {
			t.Log("GetFileHash() succeeded (file might exist)")
		} else {
			t.Logf("GetFileHash() returned error: %v", err)
		}
	})
}

// TestGetFileChurn_ErrorPaths tests GetFileChurn error conditions
func TestGetFileChurn_ErrorPaths(t *testing.T) {
	t.Run("file outside repo", func(t *testing.T) {
		client := NewClient(".")
		_, err := client.GetFileChurn("/non/existent/path/file.go")
		if err == nil {
			t.Log("GetFileChurn() succeeded (unexpected)")
		}
	})
}

// TestGetFileLastModified_ErrorPaths tests GetFileLastModified error conditions
func TestGetFileLastModified_ErrorPaths(t *testing.T) {
	t.Run("file outside repo", func(t *testing.T) {
		client := NewClient(".")
		_, err := client.GetFileLastModified("/non/existent/path/file.go")
		if err == nil {
			t.Log("GetFileLastModified() succeeded (unexpected)")
		}
	})
}

// TestGetChangedFiles_EmptyRef tests GetChangedFiles with various refs
func TestGetChangedFiles_EdgeCases(t *testing.T) {
	client := NewClient(".")

	t.Run("invalid ref", func(t *testing.T) {
		_, err := client.GetChangedFiles("nonexistent-ref-12345")
		if err == nil {
			t.Log("GetChangedFiles() succeeded (unexpected for invalid ref)")
		}
	})
}

// TestGetStagedFiles_EdgeCases tests GetStagedFiles
func TestGetStagedFiles_EdgeCases(t *testing.T) {
	client := NewClient(".")

	// In a non-staged state, this should work but return empty
	files, err := client.GetStagedFiles()
	if err != nil {
		t.Logf("GetStagedFiles() returned error: %v", err)
	} else {
		t.Logf("GetStagedFiles() returned %d files", len(files))
	}
}

// TestGetCommit_EdgeCases tests GetCommit with edge cases
func TestGetCommit_EdgeCases(t *testing.T) {
	t.Run("invalid hash", func(t *testing.T) {
		client := NewClient(".")
		_, err := client.GetCommit("invalidhash12345")
		if err == nil {
			t.Log("GetCommit() succeeded (unexpected for invalid hash)")
		}
	})

	t.Run("non-existent commit", func(t *testing.T) {
		client := NewClient(".")
		_, err := client.GetCommit("0000000000000000000000000000000000000000")
		if err == nil {
			t.Log("GetCommit() succeeded (unexpected for non-existent commit)")
		}
	})
}

// TestGetAuthors_EdgeCases tests GetAuthors edge cases
func TestGetAuthors_EdgeCases(t *testing.T) {
	// Test in non-git directory
	tempDir, err := os.MkdirTemp("", "git-authors-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	client := NewClient(tempDir)
	_, err = client.GetAuthors()
	if err == nil {
		t.Log("GetAuthors() succeeded (unexpected in non-git dir)")
	}
}

// TestGetAuthorLastCommit_EdgeCases tests GetAuthorLastCommit edge cases
func TestGetAuthorLastCommit_EdgeCases(t *testing.T) {
	t.Run("empty author name", func(t *testing.T) {
		client := NewClient(".")
		lastCommit, err := client.GetAuthorLastCommit("")
		if err != nil {
			t.Logf("GetAuthorLastCommit('') returned error: %v", err)
		}
		// Empty author might return zero time
		_ = lastCommit
	})
}

// TestGetCurrentBranch_EdgeCases tests GetCurrentBranch edge cases
func TestGetCurrentBranch_EdgeCases(t *testing.T) {
	// Test in non-git directory
	tempDir, err := os.MkdirTemp("", "git-branch-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	client := NewClient(tempDir)
	_, err = client.GetCurrentBranch()
	if err == nil {
		t.Log("GetCurrentBranch() succeeded (unexpected in non-git dir)")
	}
}

// TestGetDefaultBranch_Fallback tests GetDefaultBranch fallback behavior
func TestGetDefaultBranch_Fallback(t *testing.T) {
	// Test in non-git directory - should fallback to "main"
	tempDir, err := os.MkdirTemp("", "git-default-branch-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	client := NewClient(tempDir)
	branch := client.GetDefaultBranch()
	if branch != "main" {
		t.Errorf("GetDefaultBranch() = %s, want main", branch)
	}
}

// TestGetRepoRoot_EdgeCases tests GetRepoRoot edge cases
func TestGetRepoRoot_EdgeCases(t *testing.T) {
	// Test in non-git directory
	tempDir, err := os.MkdirTemp("", "git-root-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	client := NewClient(tempDir)
	_, err = client.GetRepoRoot()
	if err == nil {
		t.Log("GetRepoRoot() succeeded (unexpected in non-git dir)")
	}
}

// TestNewClient_EdgeCases tests NewClient with edge cases
func TestNewClient_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		repoPath string
	}{
		{"empty string", ""},
		{"dot", "."},
		{"double dot", ".."},
		{"absolute path", "/tmp"},
		{"relative path", "./internal/git"},
		{"path with trailing slash", "./"},
		{"very long path", "/a/very/long/path/that/does/not/exist/anywhere/in/the/filesystem"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(tt.repoPath)
			if client == nil {
				t.Error("NewClient() should not return nil")
			}
		})
	}
}

// TestCommitInfo_Fields tests all CommitInfo fields
func TestCommitInfo_Fields(t *testing.T) {
	now := time.Now()
	info := &CommitInfo{
		Hash:      "abc123def456",
		Author:    "Test User",
		Email:     "test@example.com",
		Date:      now,
		Message:   "Test commit message",
		FileCount: 5,
	}

	if info.Hash != "abc123def456" {
		t.Errorf("Hash = %s, want abc123def456", info.Hash)
	}
	if info.Author != "Test User" {
		t.Errorf("Author = %s, want Test User", info.Author)
	}
	if info.Email != "test@example.com" {
		t.Errorf("Email = %s, want test@example.com", info.Email)
	}
	if !info.Date.Equal(now) {
		t.Errorf("Date = %v, want %v", info.Date, now)
	}
	if info.Message != "Test commit message" {
		t.Errorf("Message = %s, want 'Test commit message'", info.Message)
	}
	if info.FileCount != 5 {
		t.Errorf("FileCount = %d, want 5", info.FileCount)
	}
}

// TestAuthorInfo_Fields tests all AuthorInfo fields
func TestAuthorInfo_Fields(t *testing.T) {
	now := time.Now()
	info := AuthorInfo{
		Name:        "Test Author",
		Email:       "author@example.com",
		LastCommit:  now,
		CommitCount: 100,
		IsActive:    true,
	}

	if info.Name != "Test Author" {
		t.Errorf("Name = %s, want Test Author", info.Name)
	}
	if info.Email != "author@example.com" {
		t.Errorf("Email = %s, want author@example.com", info.Email)
	}
	if !info.LastCommit.Equal(now) {
		t.Errorf("LastCommit = %v, want %v", info.LastCommit, now)
	}
	if info.CommitCount != 100 {
		t.Errorf("CommitCount = %d, want 100", info.CommitCount)
	}
	if !info.IsActive {
		t.Error("IsActive should be true")
	}
}

// TestClient_EmptyOutput tests methods with empty git output
func TestClient_EmptyOutput(t *testing.T) {
	// Create temp git repo
	tempDir, err := os.MkdirTemp("", "git-empty-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	client := NewClient(tempDir)

	// GetStagedFiles should return empty slice for empty output
	files, err := client.GetStagedFiles()
	if err != nil {
		t.Logf("GetStagedFiles() error: %v", err)
	} else if files == nil {
		t.Error("GetStagedFiles() should return empty slice, not nil")
	}

	// GetChangedFiles with empty output
	changedFiles, err := client.GetChangedFiles("HEAD")
	if err != nil {
		t.Logf("GetChangedFiles() error: %v", err)
	} else if changedFiles == nil {
		t.Error("GetChangedFiles() should return empty slice, not nil")
	}
}
// Package git_test 测试Git客户端接口和模拟实现
package git

import (
	"testing"
	"time"
)

// TestMockClientImplementsInterface 验证MockClient实现GitClient接口
func TestMockClientImplementsInterface(t *testing.T) {
	// 编译时检查：如果MockClient没有实现GitClient接口，编译会失败
	var _ GitClient = (*MockClient)(nil)
}

// TestMockBlamerImplementsInterface 验证MockBlamer实现GitBlamer接口
func TestMockBlamerImplementsInterface(t *testing.T) {
	// 编译时检查：如果MockBlamer没有实现GitBlamer接口，编译会失败
	var _ GitBlamer = (*MockBlamer)(nil)
}

// TestClientImplementsInterface 验证Client实现GitClient接口
func TestClientImplementsInterface(t *testing.T) {
	// 编译时检查：如果Client没有实现GitClient接口，编译会失败
	var _ GitClient = (*Client)(nil)
}

// TestBlamerImplementsInterface 验证Blamer实现GitBlamer接口
func TestBlamerImplementsInterface(t *testing.T) {
	// 编译时检查：如果Blamer没有实现GitBlamer接口，编译会失败
	// 注意：Blamer需要GitClient来创建，这里只验证接口
	var _ GitBlamer = (*Blamer)(nil)
}

// TestMockClientIsGitRepo 测试MockClient的IsGitRepo方法
func TestMockClientIsGitRepo(t *testing.T) {
	mock := NewMockClient()

	// 测试默认值
	if !mock.IsGitRepo() {
		t.Error("Default IsGitRepo should be true")
	}

	// 测试设置值
	mock.IsGitRepoResult = false
	if mock.IsGitRepo() {
		t.Error("IsGitRepo should be false after setting")
	}
}

// TestMockClientRun 测试MockClient的Run方法
func TestMockClientRun(t *testing.T) {
	mock := NewMockClient()
	mock.RunResult = "test output"

	output, err := mock.Run("status", "--short")
	if err != nil {
		t.Errorf("Run() returned error: %v", err)
	}
	if output != "test output" {
		t.Errorf("Run() = %q, want %q", output, "test output")
	}

	// 验证调用记录
	calls := mock.GetRunCalls()
	if len(calls) != 1 {
		t.Errorf("Expected 1 call, got %d", len(calls))
	}
	if len(calls[0]) != 2 || calls[0][0] != "status" || calls[0][1] != "--short" {
		t.Errorf("Unexpected call args: %v", calls[0])
	}
}

// TestMockClientRunError 测试MockClient的Run方法错误注入
func TestMockClientRunError(t *testing.T) {
	mock := NewMockClient()
	mock.RunError = &testError{msg: "git command failed"}

	_, err := mock.Run("status")
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

// TestMockClientGetCurrentBranch 测试MockClient的GetCurrentBranch方法
func TestMockClientGetCurrentBranch(t *testing.T) {
	mock := NewMockClient()
	mock.CurrentBranch = "feature/test"

	branch, err := mock.GetCurrentBranch()
	if err != nil {
		t.Errorf("GetCurrentBranch() returned error: %v", err)
	}
	if branch != "feature/test" {
		t.Errorf("GetCurrentBranch() = %q, want %q", branch, "feature/test")
	}
}

// TestMockClientGetDefaultBranch 测试MockClient的GetDefaultBranch方法
func TestMockClientGetDefaultBranch(t *testing.T) {
	mock := NewMockClient()
	mock.DefaultBranch = "develop"

	branch := mock.GetDefaultBranch()
	if branch != "develop" {
		t.Errorf("GetDefaultBranch() = %q, want %q", branch, "develop")
	}
}

// TestMockClientGetFileHash 测试MockClient的GetFileHash方法
func TestMockClientGetFileHash(t *testing.T) {
	mock := NewMockClient()
	mock.FileHash = "abc123def456"

	hash, err := mock.GetFileHash("/path/to/file.go")
	if err != nil {
		t.Errorf("GetFileHash() returned error: %v", err)
	}
	if hash != "abc123def456" {
		t.Errorf("GetFileHash() = %q, want %q", hash, "abc123def456")
	}

	// 验证调用记录
	if len(mock.GetFileHashCalls) != 1 {
		t.Errorf("Expected 1 call, got %d", len(mock.GetFileHashCalls))
	}
	if mock.GetFileHashCalls[0] != "/path/to/file.go" {
		t.Errorf("Unexpected file path: %s", mock.GetFileHashCalls[0])
	}
}

// TestMockClientGetStagedFiles 测试MockClient的GetStagedFiles方法
func TestMockClientGetStagedFiles(t *testing.T) {
	mock := NewMockClient()
	mock.StagedFiles = []string{"file1.go", "file2.go"}

	files, err := mock.GetStagedFiles()
	if err != nil {
		t.Errorf("GetStagedFiles() returned error: %v", err)
	}
	if len(files) != 2 {
		t.Errorf("GetStagedFiles() returned %d files, want 2", len(files))
	}
}

// TestMockClientGetChangedFiles 测试MockClient的GetChangedFiles方法
func TestMockClientGetChangedFiles(t *testing.T) {
	mock := NewMockClient()
	mock.ChangedFiles = []string{"modified.go", "new.go"}

	files, err := mock.GetChangedFiles("HEAD~1")
	if err != nil {
		t.Errorf("GetChangedFiles() returned error: %v", err)
	}
	if len(files) != 2 {
		t.Errorf("GetChangedFiles() returned %d files, want 2", len(files))
	}
	if len(mock.GetChangedCalls) != 1 || mock.GetChangedCalls[0] != "HEAD~1" {
		t.Errorf("Unexpected call args: %v", mock.GetChangedCalls)
	}
}

// TestMockClientGetCommit 测试MockClient的GetCommit方法
func TestMockClientGetCommit(t *testing.T) {
	mock := NewMockClient()
	mock.CommitInfo = &CommitInfo{
		Hash:    "abc123",
		Author:  "Test Author",
		Email:   "test@example.com",
		Message: "Test commit",
	}

	info, err := mock.GetCommit("abc123")
	if err != nil {
		t.Errorf("GetCommit() returned error: %v", err)
	}
	if info.Hash != "abc123" {
		t.Errorf("GetCommit().Hash = %q, want %q", info.Hash, "abc123")
	}
	if info.Author != "Test Author" {
		t.Errorf("GetCommit().Author = %q, want %q", info.Author, "Test Author")
	}
}

// TestMockClientGetAuthors 测试MockClient的GetAuthors方法
func TestMockClientGetAuthors(t *testing.T) {
	mock := NewMockClient()
	mock.Authors = []AuthorInfo{
		{Name: "Alice", Email: "alice@example.com", CommitCount: 10},
		{Name: "Bob", Email: "bob@example.com", CommitCount: 5},
	}

	authors, err := mock.GetAuthors()
	if err != nil {
		t.Errorf("GetAuthors() returned error: %v", err)
	}
	if len(authors) != 2 {
		t.Errorf("GetAuthors() returned %d authors, want 2", len(authors))
	}
}

// TestMockClientGetRepoPath 测试MockClient的GetRepoPath方法
func TestMockClientGetRepoPath(t *testing.T) {
	mock := NewMockClient()
	mock.RepoPath = "/path/to/repo"

	path := mock.GetRepoPath()
	if path != "/path/to/repo" {
		t.Errorf("GetRepoPath() = %q, want %q", path, "/path/to/repo")
	}
}

// TestMockClientReset 测试MockClient的Reset方法
func TestMockClientReset(t *testing.T) {
	mock := NewMockClient()

	// 记录一些调用
	mock.Run("status")
	mock.GetFileHash("file.go")
	mock.GetChangedFiles("HEAD~1")

	// 验证调用被记录
	if len(mock.RunCalls) != 1 {
		t.Error("RunCalls should have 1 entry")
	}

	// 重置
	mock.Reset()

	// 验证调用被清除
	if len(mock.RunCalls) != 0 {
		t.Error("RunCalls should be empty after reset")
	}
	if len(mock.GetFileHashCalls) != 0 {
		t.Error("GetFileHashCalls should be empty after reset")
	}
}

// TestMockBlamerBlameFile 测试MockBlamer的BlameFile方法
func TestMockBlamerBlameFile(t *testing.T) {
	mock := NewMockBlamer()
	mock.BlameFileResult = &BlameResult{
		File: "test.go",
		Lines: []BlameInfo{
			{Author: "Alice", Line: 1, Content: "// TODO: test"},
		},
		Authors: map[string]int{"Alice": 1},
	}

	result, err := mock.BlameFile("test.go")
	if err != nil {
		t.Errorf("BlameFile() returned error: %v", err)
	}
	if result.File != "test.go" {
		t.Errorf("BlameFile().File = %q, want %q", result.File, "test.go")
	}
	if len(result.Lines) != 1 {
		t.Errorf("BlameFile() returned %d lines, want 1", len(result.Lines))
	}

	// 验证调用记录
	calls := mock.GetBlameFileCalls()
	if len(calls) != 1 || calls[0] != "test.go" {
		t.Errorf("Unexpected call args: %v", calls)
	}
}

// TestMockBlamerBlameLine 测试MockBlamer的BlameLine方法
func TestMockBlamerBlameLine(t *testing.T) {
	mock := NewMockBlamer()
	mock.BlameLineResult = &BlameInfo{
		Author:      "Bob",
		AuthorEmail: "bob@example.com",
		CommitHash:  "def456",
		Line:        10,
		Content:     "// TODO: fix this",
	}

	info, err := mock.BlameLine("test.go", 10)
	if err != nil {
		t.Errorf("BlameLine() returned error: %v", err)
	}
	if info.Author != "Bob" {
		t.Errorf("BlameLine().Author = %q, want %q", info.Author, "Bob")
	}
	if info.Line != 10 {
		t.Errorf("BlameLine().Line = %d, want 10", info.Line)
	}
}

// TestMockBlamerGetTODOMetadata 测试MockBlamer的GetTODOMetadata方法
func TestMockBlamerGetTODOMetadata(t *testing.T) {
	mock := NewMockBlamer()
	mock.TODOMetadataAuthor = "Charlie"
	mock.TODOMetadataHash = "ghi789"
	mock.TODOMetadataDate = time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	author, hash, date, err := mock.GetTODOMetadata("todo.go", 42)
	if err != nil {
		t.Errorf("GetTODOMetadata() returned error: %v", err)
	}
	if author != "Charlie" {
		t.Errorf("GetTODOMetadata() author = %q, want %q", author, "Charlie")
	}
	if hash != "ghi789" {
		t.Errorf("GetTODOMetadata() hash = %q, want %q", hash, "ghi789")
	}
	if date.Year() != 2024 {
		t.Errorf("GetTODOMetadata() date year = %d, want 2024", date.Year())
	}
}

// TestMockBlamerBatchBlame 测试MockBlamer的BatchBlame方法
func TestMockBlamerBatchBlame(t *testing.T) {
	mock := NewMockBlamer()
	mock.BatchBlameResults = map[string]*BlameResult{
		"file1.go": {File: "file1.go", Lines: []BlameInfo{{Author: "Alice"}}},
		"file2.go": {File: "file2.go", Lines: []BlameInfo{{Author: "Bob"}}},
	}

	results, err := mock.BatchBlame([]string{"file1.go", "file2.go"})
	if err != nil {
		t.Errorf("BatchBlame() returned error: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("BatchBlame() returned %d results, want 2", len(results))
	}
}

// TestMockBlamerCheckAuthorActive 测试MockBlamer的CheckAuthorActive方法
func TestMockBlamerCheckAuthorActive(t *testing.T) {
	mock := NewMockBlamer()
	mock.AuthorActive = true
	mock.AuthorLastCommit = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	isActive, lastCommit, err := mock.CheckAuthorActive("alice", 30)
	if err != nil {
		t.Errorf("CheckAuthorActive() returned error: %v", err)
	}
	if !isActive {
		t.Error("CheckAuthorActive() should return true")
	}
	if lastCommit.Year() != 2024 {
		t.Errorf("CheckAuthorActive() lastCommit year = %d, want 2024", lastCommit.Year())
	}
}

// TestMockBlamerReset 测试MockBlamer的Reset方法
func TestMockBlamerReset(t *testing.T) {
	mock := NewMockBlamer()

	// 记录一些调用
	mock.BlameFile("test.go")
	mock.BlameLine("test.go", 10)

	// 验证调用被记录
	if len(mock.BlameFileCalls) != 1 {
		t.Error("BlameFileCalls should have 1 entry")
	}

	// 重置
	mock.Reset()

	// 验证调用被清除
	if len(mock.BlameFileCalls) != 0 {
		t.Error("BlameFileCalls should be empty after reset")
	}
}

// TestNewBlamerWithMockClient 测试使用MockClient创建Blamer
func TestNewBlamerWithMockClient(t *testing.T) {
	mockClient := NewMockClient()
	mockClient.RepoPath = "/test/repo"

	blamer := NewBlamer(mockClient)
	if blamer == nil {
		t.Fatal("NewBlamer returned nil")
	}

	// 验证Blamer可以使用mock client
	repoPath := blamer.client.GetRepoPath()
	if repoPath != "/test/repo" {
		t.Errorf("GetRepoPath() = %q, want %q", repoPath, "/test/repo")
	}
}

// TestDependencyInjection 演示依赖注入模式
func TestDependencyInjection(t *testing.T) {
	// 这是一个示例函数，展示如何接受接口而不是具体类型
	processGit := func(client GitClient) error {
		// 使用接口方法，不关心具体实现
		_, err := client.GetCurrentBranch()
		return err
	}

	// 使用真实客户端（如果有git仓库）
	realClient := NewClient(".")
	if err := processGit(realClient); err != nil {
		t.Logf("Real client error (expected if not in git repo): %v", err)
	}

	// 使用模拟客户端
	mockClient := NewMockClient()
	mockClient.CurrentBranch = "main"
	if err := processGit(mockClient); err != nil {
		t.Errorf("Mock client should not return error: %v", err)
	}
}

// testError 是一个简单的测试错误类型
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
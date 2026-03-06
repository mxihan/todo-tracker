// Package git 提供Git操作封装
package git

import (
	"sync"
	"time"
)

// MockClient 是GitClient接口的模拟实现，用于测试
type MockClient struct {
	mu sync.RWMutex

	// 可配置的返回值
	RepoPath            string
	IsGitRepoResult    bool
	IsGitRepoError     error
	RunResult          string
	RunError           error
	CurrentBranch      string
	CurrentBranchError  error
	DefaultBranch      string
	FileHash           string
	FileHashError      error
	StagedFiles        []string
	StagedFilesError   error
	ChangedFiles       []string
	ChangedFilesError  error
	FileChurn          int
	FileChurnError     error
	FileLastModified   time.Time
	FileLastModError   error
	CommitInfo         *CommitInfo
	CommitError        error
	Authors            []AuthorInfo
	AuthorsError       error
	AuthorLastCommit   time.Time
	AuthorLastCommitErr error
	RepoRoot           string
	RepoRootError      error

	// 记录调用以便验证
	RunCalls          [][]string
	GetFileHashCalls  []string
	GetFileChurnCalls []string
	GetChangedCalls   []string
	GetCommitCalls    []string
	GetAuthorCalls    []string
}

// NewMockClient 创建新的MockClient实例，带有默认值
func NewMockClient() *MockClient {
	return &MockClient{
		IsGitRepoResult: true,
		DefaultBranch:   "main",
		StagedFiles:     []string{},
		ChangedFiles:    []string{},
		CommitInfo:      &CommitInfo{},
		Authors:         []AuthorInfo{},
	}
}

// IsGitRepo 实现 GitClient 接口
func (m *MockClient) IsGitRepo() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.IsGitRepoResult
}

// Run 实现 GitClient 接口
func (m *MockClient) Run(args ...string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.RunCalls = append(m.RunCalls, args)
	return m.RunResult, m.RunError
}

// GetCurrentBranch 实现 GitClient 接口
func (m *MockClient) GetCurrentBranch() (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.CurrentBranch, m.CurrentBranchError
}

// GetDefaultBranch 实现 GitClient 接口
func (m *MockClient) GetDefaultBranch() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.DefaultBranch
}

// GetFileHash 实现 GitClient 接口
func (m *MockClient) GetFileHash(filePath string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.GetFileHashCalls = append(m.GetFileHashCalls, filePath)
	return m.FileHash, m.FileHashError
}

// GetStagedFiles 实现 GitClient 接口
func (m *MockClient) GetStagedFiles() ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.StagedFiles, m.StagedFilesError
}

// GetChangedFiles 实现 GitClient 接口
func (m *MockClient) GetChangedFiles(sinceRef string) ([]string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.GetChangedCalls = append(m.GetChangedCalls, sinceRef)
	return m.ChangedFiles, m.ChangedFilesError
}

// GetFileChurn 实现 GitClient 接口
func (m *MockClient) GetFileChurn(filePath string) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.GetFileChurnCalls = append(m.GetFileChurnCalls, filePath)
	return m.FileChurn, m.FileChurnError
}

// GetFileLastModified 实现 GitClient 接口
func (m *MockClient) GetFileLastModified(filePath string) (time.Time, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.FileLastModified, m.FileLastModError
}

// GetCommit 实现 GitClient 接口
func (m *MockClient) GetCommit(hash string) (*CommitInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.GetCommitCalls = append(m.GetCommitCalls, hash)
	return m.CommitInfo, m.CommitError
}

// GetAuthors 实现 GitClient 接口
func (m *MockClient) GetAuthors() ([]AuthorInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.Authors, m.AuthorsError
}

// GetAuthorLastCommit 实现 GitClient 接口
func (m *MockClient) GetAuthorLastCommit(authorName string) (time.Time, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.GetAuthorCalls = append(m.GetAuthorCalls, authorName)
	return m.AuthorLastCommit, m.AuthorLastCommitErr
}

// GetRepoRoot 实现 GitClient 接口
func (m *MockClient) GetRepoRoot() (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.RepoRoot, m.RepoRootError
}

// GetRepoPath 实现 GitClient 接口
func (m *MockClient) GetRepoPath() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.RepoPath
}

// MockBlamer 是GitBlamer接口的模拟实现，用于测试
type MockBlamer struct {
	mu sync.RWMutex

	// 可配置的返回值
	BlameFileResult   *BlameResult
	BlameFileError    error
	BlameLineResult   *BlameInfo
	BlameLineError    error
	TODOMetadataAuthor    string
	TODOMetadataHash      string
	TODOMetadataDate      time.Time
	TODOMetadataError     error
	BatchBlameResults map[string]*BlameResult
	BatchBlameError   error
	AuthorActive      bool
	AuthorLastCommit  time.Time
	AuthorActiveError error

	// 记录调用以便验证
	BlameFileCalls   []string
	BlameLineCalls   []blameLineCall
	TODOMetadataCalls []todoMetadataCall
	BatchBlameCalls  [][]string
	AuthorActiveCalls []authorActiveCall
}

type blameLineCall struct {
	filePath string
	lineNum  int
}

type todoMetadataCall struct {
	filePath string
	lineNum  int
}

type authorActiveCall struct {
	author       string
	inactiveDays int
}

// NewMockBlamer 创建新的MockBlamer实例
func NewMockBlamer() *MockBlamer {
	return &MockBlamer{
		BlameFileResult:   &BlameResult{Lines: []BlameInfo{}, Authors: make(map[string]int)},
		BlameLineResult:   &BlameInfo{},
		BatchBlameResults: make(map[string]*BlameResult),
	}
}

// BlameFile 实现 GitBlamer 接口
func (m *MockBlamer) BlameFile(filePath string) (*BlameResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.BlameFileCalls = append(m.BlameFileCalls, filePath)
	return m.BlameFileResult, m.BlameFileError
}

// BlameLine 实现 GitBlamer 接口
func (m *MockBlamer) BlameLine(filePath string, lineNum int) (*BlameInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.BlameLineCalls = append(m.BlameLineCalls, blameLineCall{filePath, lineNum})
	return m.BlameLineResult, m.BlameLineError
}

// GetTODOMetadata 实现 GitBlamer 接口
func (m *MockBlamer) GetTODOMetadata(filePath string, lineNum int) (string, string, time.Time, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.TODOMetadataCalls = append(m.TODOMetadataCalls, todoMetadataCall{filePath, lineNum})
	return m.TODOMetadataAuthor, m.TODOMetadataHash, m.TODOMetadataDate, m.TODOMetadataError
}

// BatchBlame 实现 GitBlamer 接口
func (m *MockBlamer) BatchBlame(filePaths []string) (map[string]*BlameResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.BatchBlameCalls = append(m.BatchBlameCalls, filePaths)
	return m.BatchBlameResults, m.BatchBlameError
}

// CheckAuthorActive 实现 GitBlamer 接口
func (m *MockBlamer) CheckAuthorActive(author string, inactiveDays int) (bool, time.Time, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.AuthorActiveCalls = append(m.AuthorActiveCalls, authorActiveCall{author, inactiveDays})
	return m.AuthorActive, m.AuthorLastCommit, m.AuthorActiveError
}

// GetRunCalls 返回所有Run调用的参数
func (m *MockClient) GetRunCalls() [][]string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	calls := make([][]string, len(m.RunCalls))
	copy(calls, m.RunCalls)
	return calls
}

// GetBlameFileCalls 返回所有BlameFile调用的参数
func (m *MockBlamer) GetBlameFileCalls() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	calls := make([]string, len(m.BlameFileCalls))
	copy(calls, m.BlameFileCalls)
	return calls
}

// Reset 重置MockClient的所有调用记录
func (m *MockClient) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.RunCalls = nil
	m.GetFileHashCalls = nil
	m.GetFileChurnCalls = nil
	m.GetChangedCalls = nil
	m.GetCommitCalls = nil
	m.GetAuthorCalls = nil
}

// Reset 重置MockBlamer的所有调用记录
func (m *MockBlamer) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.BlameFileCalls = nil
	m.BlameLineCalls = nil
	m.TODOMetadataCalls = nil
	m.BatchBlameCalls = nil
	m.AuthorActiveCalls = nil
}
// Package git 提供Git操作封装
package git

import "time"

// GitClient 定义Git客户端接口，用于依赖注入和测试
// 所有方法都是线程安全的，可以在多个goroutine中并发调用
type GitClient interface {
	// IsGitRepo 检查当前路径是否是Git仓库
	IsGitRepo() bool

	// Run 执行Git命令并返回输出
	// args 是传递给git命令的参数
	Run(args ...string) (string, error)

	// GetCurrentBranch 获取当前分支名
	GetCurrentBranch() (string, error)

	// GetDefaultBranch 获取默认分支名
	// 如果无法确定，返回 "main"
	GetDefaultBranch() string

	// GetFileHash 获取文件内容的Git哈希
	GetFileHash(filePath string) (string, error)

	// GetStagedFiles 获取暂存区文件列表
	GetStagedFiles() ([]string, error)

	// GetChangedFiles 获取指定commit后的变更文件
	GetChangedFiles(sinceRef string) ([]string, error)

	// GetFileChurn 获取文件的修改次数
	GetFileChurn(filePath string) (int, error)

	// GetFileLastModified 获取文件最后修改时间
	GetFileLastModified(filePath string) (time.Time, error)

	// GetCommit 获取指定提交的信息
	GetCommit(hash string) (*CommitInfo, error)

	// GetAuthors 获取所有作者列表
	GetAuthors() ([]AuthorInfo, error)

	// GetAuthorLastCommit 获取作者最后提交时间
	GetAuthorLastCommit(authorName string) (time.Time, error)

	// GetRepoRoot 获取仓库根目录
	GetRepoRoot() (string, error)

	// GetRepoPath 获取客户端配置的仓库路径
	// 用于需要直接访问路径的场景（如Blame操作）
	GetRepoPath() string
}

// GitBlamer 定义Git Blame操作接口
type GitBlamer interface {
	// BlameFile 对整个文件执行Git Blame
	BlameFile(filePath string) (*BlameResult, error)

	// BlameLine 对特定行执行Git Blame
	BlameLine(filePath string, lineNum int) (*BlameInfo, error)

	// GetTODOMetadata 获取TODO的Git元数据
	GetTODOMetadata(filePath string, lineNum int) (author string, commitHash string, commitDate time.Time, err error)

	// BatchBlame 批量获取多个文件的Blame信息
	BatchBlame(filePaths []string) (map[string]*BlameResult, error)

	// CheckAuthorActive 检查作者是否活跃
	CheckAuthorActive(author string, inactiveDays int) (bool, time.Time, error)
}
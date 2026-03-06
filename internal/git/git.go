// Package git 提供Git操作封装
package git

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Client Git客户端
type Client struct {
	repoPath string
}

// NewClient 创建新的Git客户端
func NewClient(repoPath string) *Client {
	return &Client{
		repoPath: repoPath,
	}
}

// IsGitRepo 检查是否是Git仓库
func (c *Client) IsGitRepo() bool {
	gitDir := filepath.Join(c.repoPath, ".git")
	_, err := os.Stat(gitDir)
	return err == nil
}

// Run 执行Git命令
func (c *Client) Run(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = c.repoPath

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("git %s 失败: %s", args[0], stderr.String())
	}

	return strings.TrimSpace(stdout.String()), nil
}

// GetCurrentBranch 获取当前分支名
func (c *Client) GetCurrentBranch() (string, error) {
	return c.Run("rev-parse", "--abbrev-ref", "HEAD")
}

// GetDefaultBranch 获取默认分支名
func (c *Client) GetDefaultBranch() string {
	// 尝试获取远程默认分支
	branches := []string{"main", "master", "develop"}
	for _, branch := range branches {
		if _, err := c.Run("rev-parse", "--verify", branch); err == nil {
			return branch
		}
	}
	return "main"
}

// GetFileHash 获取文件内容的Git哈希
func (c *Client) GetFileHash(filePath string) (string, error) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return "", err
	}

	_, err = filepath.Rel(c.repoPath, absPath)
	if err != nil {
		return "", err
	}

	// 使用git hash-object计算文件哈希
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	cmd := exec.Command("git", "hash-object", "--stdin")
	cmd.Dir = c.repoPath
	cmd.Stdin = bytes.NewReader(content)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		return "", fmt.Errorf("计算文件哈希失败: %s", stderr.String())
	}

	return strings.TrimSpace(stdout.String()), nil
}

// GetStagedFiles 获取暂存区文件列表
func (c *Client) GetStagedFiles() ([]string, error) {
	output, err := c.Run("diff", "--cached", "--name-only")
	if err != nil {
		return nil, err
	}

	if output == "" {
		return []string{}, nil
	}

	return strings.Split(output, "\n"), nil
}

// GetChangedFiles 获取指定commit后的变更文件
func (c *Client) GetChangedFiles(sinceRef string) ([]string, error) {
	output, err := c.Run("diff", "--name-only", sinceRef+"...HEAD")
	if err != nil {
		return nil, err
	}

	if output == "" {
		return []string{}, nil
	}

	return strings.Split(output, "\n"), nil
}

// GetFileChurn 获取文件的修改次数
func (c *Client) GetFileChurn(filePath string) (int, error) {
	relPath, err := filepath.Rel(c.repoPath, filePath)
	if err != nil {
		return 0, err
	}

	output, err := c.Run("log", "--oneline", "--follow", "--", relPath)
	if err != nil {
		return 0, err
	}

	if output == "" {
		return 0, nil
	}

	// 计算提交数量
	lines := strings.Split(output, "\n")
	return len(lines), nil
}

// GetFileLastModified 获取文件最后修改时间
func (c *Client) GetFileLastModified(filePath string) (time.Time, error) {
	relPath, err := filepath.Rel(c.repoPath, filePath)
	if err != nil {
		return time.Time{}, err
	}

	output, err := c.Run("log", "-1", "--format=%ct", "--", relPath)
	if err != nil {
		return time.Time{}, err
	}

	if output == "" {
		return time.Time{}, nil
	}

	var timestamp int64
	fmt.Sscanf(output, "%d", &timestamp)

	return time.Unix(timestamp, 0), nil
}

// CommitInfo 提交信息
type CommitInfo struct {
	Hash      string
	Author    string
	Email     string
	Date      time.Time
	Message   string
	FileCount int
}

// GetCommit 获取指定提交的信息
func (c *Client) GetCommit(hash string) (*CommitInfo, error) {
	output, err := c.Run("show", "--format=%H%n%an%n%ae%n%ct%n%s", "--stat", hash)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(output, "\n")
	if len(lines) < 4 {
		return nil, fmt.Errorf("无效的提交信息格式")
	}

	info := &CommitInfo{
		Hash:    lines[0],
		Author:  lines[1],
		Email:   lines[2],
		Message: lines[4],
	}

	var timestamp int64
	fmt.Sscanf(lines[3], "%d", &timestamp)
	info.Date = time.Unix(timestamp, 0)

	return info, nil
}

// AuthorInfo 作者信息
type AuthorInfo struct {
	Name        string
	Email       string
	LastCommit  time.Time
	CommitCount int
	IsActive    bool
}

// GetAuthors 获取所有作者列表
func (c *Client) GetAuthors() ([]AuthorInfo, error) {
	output, err := c.Run("shortlog", "-sne", "--all")
	if err != nil {
		return nil, err
	}

	var authors []AuthorInfo
	scanner := bufio.NewScanner(strings.NewReader(output))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// 格式: "  123\tJohn Doe <john@example.com>"
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) != 2 {
			continue
		}

		var count int
		fmt.Sscanf(parts[0], "%d", &count)

		// 解析名称和邮箱
		nameEmail := parts[1]
		emailStart := strings.LastIndex(nameEmail, "<")
		emailEnd := strings.LastIndex(nameEmail, ">")

		var name, email string
		if emailStart != -1 && emailEnd != -1 {
			name = strings.TrimSpace(nameEmail[:emailStart])
			email = nameEmail[emailStart+1 : emailEnd]
		} else {
			name = nameEmail
		}

		authors = append(authors, AuthorInfo{
			Name:        name,
			Email:       email,
			CommitCount: count,
			IsActive:    true, // 默认活跃
		})
	}

	return authors, nil
}

// GetAuthorLastCommit 获取作者最后提交时间
func (c *Client) GetAuthorLastCommit(authorName string) (time.Time, error) {
	output, err := c.Run("log", "-1", "--format=%ct", "--author="+authorName)
	if err != nil {
		return time.Time{}, err
	}

	if output == "" {
		return time.Time{}, nil
	}

	var timestamp int64
	fmt.Sscanf(output, "%d", &timestamp)

	return time.Unix(timestamp, 0), nil
}

// GetRepoRoot 获取仓库根目录
func (c *Client) GetRepoRoot() (string, error) {
	return c.Run("rev-parse", "--show-toplevel")
}

// GetRepoPath 获取客户端配置的仓库路径
func (c *Client) GetRepoPath() string {
	return c.repoPath
}
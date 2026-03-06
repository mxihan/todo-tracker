// Package git 提供Git Blame功能
package git

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// BlameInfo Git Blame信息
type BlameInfo struct {
	Author      string    // 作者名
	AuthorEmail string    // 作者邮箱
	CommitHash  string    // 提交哈希
	CommitDate  time.Time // 提交日期
	Line        int       // 行号
	Content     string    // 行内容
}

// BlameResult Git Blame结果
type BlameResult struct {
	File    string       // 文件路径
	Lines   []BlameInfo  // 每行的Blame信息
	Authors map[string]int // 作者统计
}

// Blamer Git Blame处理器
type Blamer struct {
	client GitClient
}

// NewBlamer 创建新的Blamer
// 参数 client 可以是 *Client 或 MockClient 等实现了 GitClient 接口的类型
func NewBlamer(client GitClient) *Blamer {
	return &Blamer{
		client: client,
	}
}

// BlameFile 对整个文件执行Git Blame
func (b *Blamer) BlameFile(filePath string) (*BlameResult, error) {
	repoPath := b.client.GetRepoPath()
	relPath, err := filepath.Rel(repoPath, filePath)
	if err != nil {
		return nil, err
	}

	// 执行git blame命令
	cmd := exec.Command("git", "blame", "--line-porcelain", relPath)
	cmd.Dir = repoPath

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("git blame 失败: %s", stderr.String())
	}

	result := &BlameResult{
		File:    filePath,
		Lines:   make([]BlameInfo, 0),
		Authors: make(map[string]int),
	}

	b.parseBlameOutput(stdout.String(), result)

	return result, nil
}

// parseBlameOutput 解析git blame输出
func (b *Blamer) parseBlameOutput(output string, result *BlameResult) {
	scanner := bufio.NewScanner(strings.NewReader(output))

	var currentInfo BlameInfo
	var commitHash string

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "author ") {
			currentInfo.Author = strings.TrimPrefix(line, "author ")
		} else if strings.HasPrefix(line, "author-mail ") {
			email := strings.TrimPrefix(line, "author-mail ")
			email = strings.Trim(email, "<>")
			currentInfo.AuthorEmail = email
		} else if strings.HasPrefix(line, "author-time ") {
			timeStr := strings.TrimPrefix(line, "author-time ")
			var timestamp int64
			fmt.Sscanf(timeStr, "%d", &timestamp)
			currentInfo.CommitDate = time.Unix(timestamp, 0)
		} else if len(line) >= 40 && isHexString(line[:40]) {
			// 提交哈希行
			parts := strings.SplitN(line, " ", 3)
			if len(parts) >= 2 {
				commitHash = parts[0]
				currentInfo = BlameInfo{
					CommitHash: commitHash,
				}
			}
		} else if strings.HasPrefix(line, "\t") {
			// 实际代码行
			currentInfo.Content = strings.TrimPrefix(line, "\t")

			// 统计作者
			if currentInfo.Author != "" {
				result.Authors[currentInfo.Author]++
			}

			result.Lines = append(result.Lines, currentInfo)

			// 重置当前信息
			currentInfo = BlameInfo{}
		}
	}
}

// BlameLine 对特定行执行Git Blame
func (b *Blamer) BlameLine(filePath string, lineNum int) (*BlameInfo, error) {
	repoPath := b.client.GetRepoPath()
	relPath, err := filepath.Rel(repoPath, filePath)
	if err != nil {
		return nil, err
	}

	// 执行git blame命令，只获取指定行
	cmd := exec.Command("git", "blame", "-L", fmt.Sprintf("%d,%d", lineNum, lineNum), "--line-porcelain", relPath)
	cmd.Dir = repoPath

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("git blame 失败: %s", stderr.String())
	}

	info := &BlameInfo{
		Line: lineNum,
	}

	b.parseSingleBlameOutput(stdout.String(), info)

	return info, nil
}

// parseSingleBlameOutput 解析单行blame输出
func (b *Blamer) parseSingleBlameOutput(output string, info *BlameInfo) {
	scanner := bufio.NewScanner(strings.NewReader(output))

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "author ") {
			info.Author = strings.TrimPrefix(line, "author ")
		} else if strings.HasPrefix(line, "author-mail ") {
			email := strings.TrimPrefix(line, "author-mail ")
			email = strings.Trim(email, "<>")
			info.AuthorEmail = email
		} else if strings.HasPrefix(line, "author-time ") {
			timeStr := strings.TrimPrefix(line, "author-time ")
			var timestamp int64
			fmt.Sscanf(timeStr, "%d", &timestamp)
			info.CommitDate = time.Unix(timestamp, 0)
		} else if strings.HasPrefix(line, "\t") {
			info.Content = strings.TrimPrefix(line, "\t")
		} else if len(line) >= 40 && isHexString(line[:40]) {
			parts := strings.SplitN(line, " ", 3)
			if len(parts) >= 1 {
				info.CommitHash = parts[0]
			}
		}
	}
}

// GetTODOMetadata 获取TODO的Git元数据
func (b *Blamer) GetTODOMetadata(filePath string, lineNum int) (author string, commitHash string, commitDate time.Time, err error) {
	info, err := b.BlameLine(filePath, lineNum)
	if err != nil {
		return "", "", time.Time{}, err
	}

	return info.Author, info.CommitHash, info.CommitDate, nil
}

// BatchBlame 批量获取多个文件的Blame信息
func (b *Blamer) BatchBlame(filePaths []string) (map[string]*BlameResult, error) {
	results := make(map[string]*BlameResult)

	for _, filePath := range filePaths {
		result, err := b.BlameFile(filePath)
		if err != nil {
			// 记录错误但继续处理其他文件
			results[filePath] = nil
			continue
		}
		results[filePath] = result
	}

	return results, nil
}

// isHexString 检查字符串是否是有效的十六进制字符串
func isHexString(s string) bool {
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

// CheckAuthorActive 检查作者是否活跃（在指定天数内有提交）
func (b *Blamer) CheckAuthorActive(author string, inactiveDays int) (bool, time.Time, error) {
	lastCommit, err := b.client.GetAuthorLastCommit(author)
	if err != nil {
		return false, time.Time{}, err
	}

	if lastCommit.IsZero() {
		return false, time.Time{}, nil
	}

	threshold := time.Now().AddDate(0, 0, -inactiveDays)
	isActive := lastCommit.After(threshold)

	return isActive, lastCommit, nil
}
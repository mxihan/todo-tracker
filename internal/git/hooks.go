// Package git 提供 Git Hook 管理功能
package git

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// HookManager Git Hook 管理器
type HookManager struct {
	repoRoot string
	hooksDir string
}

// HookType Hook 类型
type HookType string

const (
	// HookPreCommit pre-commit hook
	HookPreCommit HookType = "pre-commit"
	// HookPrePush pre-push hook
	HookPrePush HookType = "pre-push"
	// HookCommitMsg commit-msg hook
	HookCommitMsg HookType = "commit-msg"
)

// NewHookManager 创建 Hook 管理器
func NewHookManager(repoRoot string) (*HookManager, error) {
	// 查找 .git 目录
	gitDir := findGitDir(repoRoot)
	if gitDir == "" {
		return nil, fmt.Errorf("未找到 Git 仓库: %s", repoRoot)
	}

	hooksDir := filepath.Join(gitDir, "hooks")

	return &HookManager{
		repoRoot: repoRoot,
		hooksDir: hooksDir,
	}, nil
}

// Install 安装 Git Hook
func (m *HookManager) Install(hookType HookType, content string) error {
	// 确保 hooks 目录存在
	if err := os.MkdirAll(m.hooksDir, 0755); err != nil {
		return fmt.Errorf("创建 hooks 目录失败: %w", err)
	}

	hookPath := filepath.Join(m.hooksDir, string(hookType))

	// 检查是否已存在
	if _, err := os.Stat(hookPath); err == nil {
		// 备份现有 hook
		backupPath := hookPath + ".backup"
		if err := os.Rename(hookPath, backupPath); err != nil {
			return fmt.Errorf("备份现有 hook 失败: %w", err)
		}
	}

	// 写入新的 hook
	if err := os.WriteFile(hookPath, []byte(content), 0755); err != nil {
		return fmt.Errorf("写入 hook 文件失败: %w", err)
	}

	return nil
}

// InstallPreCommit 安装 pre-commit hook
func (m *HookManager) InstallPreCommit() error {
	content := `#!/bin/sh
# TODO Tracker pre-commit hook
# 检查暂存文件中的 TODO

echo "检查暂存文件中的 TODO..."
todo scan --staged --fail-on="BUG,FIXME!" --ci

if [ $? -ne 0 ]; then
    echo "发现高优先级 TODO，请先处理"
    exit 1
fi

exit 0
`
	return m.Install(HookPreCommit, content)
}

// InstallPrePush 安装 pre-push hook
func (m *HookManager) InstallPrePush() error {
	content := `#!/bin/sh
# TODO Tracker pre-push hook
# 检查过期和孤儿 TODO

echo "检查过期和孤儿 TODO..."
todo stale --ci
todo orphaned --ci

exit 0
`
	return m.Install(HookPrePush, content)
}

// Uninstall 卸载 Git Hook
func (m *HookManager) Uninstall(hookType HookType) error {
	hookPath := filepath.Join(m.hooksDir, string(hookType))

	// 检查是否存在
	if _, err := os.Stat(hookPath); os.IsNotExist(err) {
		return fmt.Errorf("hook 不存在: %s", hookType)
	}

	// 删除 hook
	if err := os.Remove(hookPath); err != nil {
		return fmt.Errorf("删除 hook 失败: %w", err)
	}

	// 恢复备份（如果存在）
	backupPath := hookPath + ".backup"
	if _, err := os.Stat(backupPath); err == nil {
		if err := os.Rename(backupPath, hookPath); err != nil {
			fmt.Printf("警告: 无法恢复备份 hook: %v\n", err)
		}
	}

	return nil
}

// Status 检查 Hook 状态
func (m *HookManager) Status() (map[HookType]HookStatus, error) {
	statuses := make(map[HookType]HookStatus)

	hookTypes := []HookType{HookPreCommit, HookPrePush, HookCommitMsg}

	for _, ht := range hookTypes {
		hookPath := filepath.Join(m.hooksDir, string(ht))

		info, err := os.Stat(hookPath)
		if os.IsNotExist(err) {
			statuses[ht] = HookStatus{
				Installed: false,
			}
			continue
		}

		// 检查是否是 TODO Tracker 的 hook
		content, err := os.ReadFile(hookPath)
		if err != nil {
			statuses[ht] = HookStatus{
				Installed: true,
				IsOurs:    false,
				Error:     err.Error(),
			}
			continue
		}

		isOurs := strings.Contains(string(content), "TODO Tracker")

		statuses[ht] = HookStatus{
			Installed: true,
			IsOurs:    isOurs,
			Path:      hookPath,
			Size:      info.Size(),
			ExecMode:  info.Mode().Perm()&0111 != 0,
		}
	}

	return statuses, nil
}

// HookStatus Hook 状态
type HookStatus struct {
	// 是否已安装
	Installed bool `json:"installed"`
	// 是否是 TODO Tracker 的 hook
	IsOurs bool `json:"is_ours"`
	// Hook 文件路径
	Path string `json:"path,omitempty"`
	// 文件大小
	Size int64 `json:"size,omitempty"`
	// 是否有执行权限
	ExecMode bool `json:"exec_mode"`
	// 错误信息
	Error string `json:"error,omitempty"`
}

// findGitDir 查找 .git 目录
func findGitDir(startPath string) string {
	path := startPath

	for {
		gitPath := filepath.Join(path, ".git")

		// 检查是否存在 .git 目录或 .git 文件（子模块）
		if info, err := os.Stat(gitPath); err == nil {
			if info.IsDir() {
				return gitPath
			}
			// 可能是子模块，读取 .git 文件获取实际路径
			if content, err := os.ReadFile(gitPath); err == nil {
				line := strings.TrimSpace(string(content))
				if strings.HasPrefix(line, "gitdir: ") {
					return strings.TrimPrefix(line, "gitdir: ")
				}
			}
		}

		// 向上一级目录
		parent := filepath.Dir(path)
		if parent == path {
			break
		}
		path = parent
	}

	return ""
}

// GetHookContent 获取 Hook 内容
func (m *HookManager) GetHookContent(hookType HookType) (string, error) {
	hookPath := filepath.Join(m.hooksDir, string(hookType))

	content, err := os.ReadFile(hookPath)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

// ListInstalled 列出已安装的 Hook
func (m *HookManager) ListInstalled() ([]HookType, error) {
	entries, err := os.ReadDir(m.hooksDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var installed []HookType
	for _, entry := range entries {
		name := entry.Name()
		switch name {
		case "pre-commit":
			installed = append(installed, HookPreCommit)
		case "pre-push":
			installed = append(installed, HookPrePush)
		case "commit-msg":
			installed = append(installed, HookCommitMsg)
		}
	}

	return installed, nil
}
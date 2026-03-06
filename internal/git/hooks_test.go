// Package git_test tests Git Hook management functionality
package git

import (
	"os"
	"path/filepath"
	"testing"
)

// TestNewHookManager tests HookManager creation
func TestNewHookManager(t *testing.T) {
	// Test with current directory (should find .git if in a git repo)
	hm, err := NewHookManager(".")
	if err != nil {
		// If not in a git repo, that's expected behavior
		t.Logf("NewHookManager() returned error (expected if not in git repo): %v", err)
	}
	if hm != nil {
		t.Log("NewHookManager() succeeded")
	}
}

// TestNewHookManager_NonGitDir tests HookManager with non-git directory
func TestNewHookManager_NonGitDir(t *testing.T) {
	// Create a temp directory that is not a git repo
	tempDir, err := os.MkdirTemp("", "non-git-dir")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	hm, err := NewHookManager(tempDir)
	if err == nil {
		t.Error("NewHookManager() should return error for non-git directory")
	}
	if hm != nil {
		t.Error("NewHookManager() should return nil for non-git directory")
	}
}

// TestHookManager_Install tests installing a hook
func TestHookManager_Install(t *testing.T) {
	// Create a temp git repo
	tempDir, err := os.MkdirTemp("", "git-hook-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize git repo
	gitDir := filepath.Join(tempDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("Failed to create .git dir: %v", err)
	}

	hm, err := NewHookManager(tempDir)
	if err != nil {
		t.Fatalf("NewHookManager() failed: %v", err)
	}

	// Test installing a hook
	content := "#!/bin/sh\necho 'test hook'"
	err = hm.Install(HookPreCommit, content)
	if err != nil {
		t.Errorf("Install() failed: %v", err)
	}

	// Verify hook was created
	hookPath := filepath.Join(hm.hooksDir, string(HookPreCommit))
	if _, err := os.Stat(hookPath); os.IsNotExist(err) {
		t.Error("Hook file was not created")
	}

	// Verify content
	readContent, err := os.ReadFile(hookPath)
	if err != nil {
		t.Fatalf("Failed to read hook file: %v", err)
	}
	if string(readContent) != content {
		t.Errorf("Hook content = %s, want %s", string(readContent), content)
	}
}

// TestHookManager_Install_ExistingHook tests installing a hook when one already exists
func TestHookManager_Install_ExistingHook(t *testing.T) {
	// Create a temp git repo
	tempDir, err := os.MkdirTemp("", "git-hook-test-existing")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize git repo
	gitDir := filepath.Join(tempDir, ".git")
	hooksDir := filepath.Join(gitDir, "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		t.Fatalf("Failed to create hooks dir: %v", err)
	}

	// Create an existing hook
	existingHook := filepath.Join(hooksDir, "pre-commit")
	if err := os.WriteFile(existingHook, []byte("existing content"), 0755); err != nil {
		t.Fatalf("Failed to create existing hook: %v", err)
	}

	hm, err := NewHookManager(tempDir)
	if err != nil {
		t.Fatalf("NewHookManager() failed: %v", err)
	}

	// Install new hook
	newContent := "#!/bin/sh\necho 'new hook'"
	err = hm.Install(HookPreCommit, newContent)
	if err != nil {
		t.Errorf("Install() failed: %v", err)
	}

	// Verify backup was created
	backupPath := existingHook + ".backup"
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Error("Backup file was not created")
	}

	// Verify new content
	readContent, err := os.ReadFile(existingHook)
	if err != nil {
		t.Fatalf("Failed to read hook file: %v", err)
	}
	if string(readContent) != newContent {
		t.Errorf("Hook content = %s, want %s", string(readContent), newContent)
	}
}

// TestHookManager_InstallPreCommit tests installing pre-commit hook
func TestHookManager_InstallPreCommit(t *testing.T) {
	// Create a temp git repo
	tempDir, err := os.MkdirTemp("", "git-precommit-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize git repo
	gitDir := filepath.Join(tempDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("Failed to create .git dir: %v", err)
	}

	hm, err := NewHookManager(tempDir)
	if err != nil {
		t.Fatalf("NewHookManager() failed: %v", err)
	}

	err = hm.InstallPreCommit()
	if err != nil {
		t.Errorf("InstallPreCommit() failed: %v", err)
	}

	// Verify hook was created
	hookPath := filepath.Join(hm.hooksDir, string(HookPreCommit))
	content, err := os.ReadFile(hookPath)
	if err != nil {
		t.Fatalf("Failed to read hook file: %v", err)
	}

	if len(content) == 0 {
		t.Error("Pre-commit hook content is empty")
	}
}

// TestHookManager_InstallPrePush tests installing pre-push hook
func TestHookManager_InstallPrePush(t *testing.T) {
	// Create a temp git repo
	tempDir, err := os.MkdirTemp("", "git-prepush-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize git repo
	gitDir := filepath.Join(tempDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("Failed to create .git dir: %v", err)
	}

	hm, err := NewHookManager(tempDir)
	if err != nil {
		t.Fatalf("NewHookManager() failed: %v", err)
	}

	err = hm.InstallPrePush()
	if err != nil {
		t.Errorf("InstallPrePush() failed: %v", err)
	}

	// Verify hook was created
	hookPath := filepath.Join(hm.hooksDir, string(HookPrePush))
	content, err := os.ReadFile(hookPath)
	if err != nil {
		t.Fatalf("Failed to read hook file: %v", err)
	}

	if len(content) == 0 {
		t.Error("Pre-push hook content is empty")
	}
}

// TestHookManager_Uninstall tests uninstalling a hook
func TestHookManager_Uninstall(t *testing.T) {
	// Create a temp git repo
	tempDir, err := os.MkdirTemp("", "git-uninstall-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize git repo
	gitDir := filepath.Join(tempDir, ".git")
	hooksDir := filepath.Join(gitDir, "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		t.Fatalf("Failed to create hooks dir: %v", err)
	}

	// Create a hook
	hookPath := filepath.Join(hooksDir, "pre-commit")
	if err := os.WriteFile(hookPath, []byte("#!/bin/sh\necho test"), 0755); err != nil {
		t.Fatalf("Failed to create hook: %v", err)
	}

	hm, err := NewHookManager(tempDir)
	if err != nil {
		t.Fatalf("NewHookManager() failed: %v", err)
	}

	// Uninstall the hook
	err = hm.Uninstall(HookPreCommit)
	if err != nil {
		t.Errorf("Uninstall() failed: %v", err)
	}

	// Verify hook was removed
	if _, err := os.Stat(hookPath); !os.IsNotExist(err) {
		t.Error("Hook file was not removed")
	}
}

// TestHookManager_Uninstall_NonExistent tests uninstalling non-existent hook
func TestHookManager_Uninstall_NonExistent(t *testing.T) {
	// Create a temp git repo
	tempDir, err := os.MkdirTemp("", "git-uninstall-notexist-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize git repo
	gitDir := filepath.Join(tempDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("Failed to create .git dir: %v", err)
	}

	hm, err := NewHookManager(tempDir)
	if err != nil {
		t.Fatalf("NewHookManager() failed: %v", err)
	}

	// Try to uninstall non-existent hook
	err = hm.Uninstall(HookPreCommit)
	if err == nil {
		t.Error("Uninstall() should return error for non-existent hook")
	}
}

// TestHookManager_Uninstall_WithBackup tests uninstalling hook with backup restore
func TestHookManager_Uninstall_WithBackup(t *testing.T) {
	// Create a temp git repo
	tempDir, err := os.MkdirTemp("", "git-uninstall-backup-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize git repo
	gitDir := filepath.Join(tempDir, ".git")
	hooksDir := filepath.Join(gitDir, "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		t.Fatalf("Failed to create hooks dir: %v", err)
	}

	// Create a hook and backup
	hookPath := filepath.Join(hooksDir, "pre-commit")
	backupPath := hookPath + ".backup"
	originalContent := "#!/bin/sh\necho original"

	if err := os.WriteFile(hookPath, []byte("#!/bin/sh\necho new"), 0755); err != nil {
		t.Fatalf("Failed to create hook: %v", err)
	}
	if err := os.WriteFile(backupPath, []byte(originalContent), 0755); err != nil {
		t.Fatalf("Failed to create backup: %v", err)
	}

	hm, err := NewHookManager(tempDir)
	if err != nil {
		t.Fatalf("NewHookManager() failed: %v", err)
	}

	// Uninstall the hook
	err = hm.Uninstall(HookPreCommit)
	if err != nil {
		t.Errorf("Uninstall() failed: %v", err)
	}

	// Verify backup was restored
	content, err := os.ReadFile(hookPath)
	if err != nil {
		t.Fatalf("Failed to read restored hook: %v", err)
	}
	if string(content) != originalContent {
		t.Errorf("Restored content = %s, want %s", string(content), originalContent)
	}

	// Verify backup was removed
	if _, err := os.Stat(backupPath); !os.IsNotExist(err) {
		t.Error("Backup file should be removed after restore")
	}
}

// TestHookManager_Status tests hook status checking
func TestHookManager_Status(t *testing.T) {
	// Create a temp git repo
	tempDir, err := os.MkdirTemp("", "git-status-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize git repo
	gitDir := filepath.Join(tempDir, ".git")
	hooksDir := filepath.Join(gitDir, "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		t.Fatalf("Failed to create hooks dir: %v", err)
	}

	hm, err := NewHookManager(tempDir)
	if err != nil {
		t.Fatalf("NewHookManager() failed: %v", err)
	}

	// Test status with no hooks
	statuses, err := hm.Status()
	if err != nil {
		t.Errorf("Status() failed: %v", err)
	}

	for hookType, status := range statuses {
		if status.Installed {
			t.Errorf("Hook %s should not be installed", hookType)
		}
	}

	// Install a hook
	if err := hm.InstallPreCommit(); err != nil {
		t.Fatalf("InstallPreCommit() failed: %v", err)
	}

	// Test status with hook installed
	statuses, err = hm.Status()
	if err != nil {
		t.Errorf("Status() failed: %v", err)
	}

	preCommitStatus, ok := statuses[HookPreCommit]
	if !ok {
		t.Fatal("Pre-commit hook status not found")
	}
	if !preCommitStatus.Installed {
		t.Error("Pre-commit hook should be installed")
	}
	if !preCommitStatus.IsOurs {
		t.Error("Pre-commit hook should be recognized as ours")
	}
}

// TestHookManager_Status_NonOurHook tests status with non-TODO Tracker hook
func TestHookManager_Status_NonOurHook(t *testing.T) {
	// Create a temp git repo
	tempDir, err := os.MkdirTemp("", "git-status-other-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize git repo
	gitDir := filepath.Join(tempDir, ".git")
	hooksDir := filepath.Join(gitDir, "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		t.Fatalf("Failed to create hooks dir: %v", err)
	}

	// Create a non-TODO Tracker hook
	hookPath := filepath.Join(hooksDir, "pre-commit")
	if err := os.WriteFile(hookPath, []byte("#!/bin/sh\necho 'other hook'"), 0755); err != nil {
		t.Fatalf("Failed to create hook: %v", err)
	}

	hm, err := NewHookManager(tempDir)
	if err != nil {
		t.Fatalf("NewHookManager() failed: %v", err)
	}

	statuses, err := hm.Status()
	if err != nil {
		t.Errorf("Status() failed: %v", err)
	}

	preCommitStatus, ok := statuses[HookPreCommit]
	if !ok {
		t.Fatal("Pre-commit hook status not found")
	}
	if !preCommitStatus.Installed {
		t.Error("Pre-commit hook should be installed")
	}
	if preCommitStatus.IsOurs {
		t.Error("Non-TODO Tracker hook should not be recognized as ours")
	}
}

// TestHookManager_GetHookContent tests getting hook content
func TestHookManager_GetHookContent(t *testing.T) {
	// Create a temp git repo
	tempDir, err := os.MkdirTemp("", "git-getcontent-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize git repo
	gitDir := filepath.Join(tempDir, ".git")
	hooksDir := filepath.Join(gitDir, "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		t.Fatalf("Failed to create hooks dir: %v", err)
	}

	// Create a hook
	hookContent := "#!/bin/sh\necho 'test'"
	hookPath := filepath.Join(hooksDir, "pre-commit")
	if err := os.WriteFile(hookPath, []byte(hookContent), 0755); err != nil {
		t.Fatalf("Failed to create hook: %v", err)
	}

	hm, err := NewHookManager(tempDir)
	if err != nil {
		t.Fatalf("NewHookManager() failed: %v", err)
	}

	// Get hook content
	content, err := hm.GetHookContent(HookPreCommit)
	if err != nil {
		t.Errorf("GetHookContent() failed: %v", err)
	}
	if content != hookContent {
		t.Errorf("Content = %s, want %s", content, hookContent)
	}
}

// TestHookManager_GetHookContent_NonExistent tests getting non-existent hook content
func TestHookManager_GetHookContent_NonExistent(t *testing.T) {
	// Create a temp git repo
	tempDir, err := os.MkdirTemp("", "git-getcontent-notexist-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize git repo
	gitDir := filepath.Join(tempDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("Failed to create .git dir: %v", err)
	}

	hm, err := NewHookManager(tempDir)
	if err != nil {
		t.Fatalf("NewHookManager() failed: %v", err)
	}

	// Get non-existent hook content
	_, err = hm.GetHookContent(HookPreCommit)
	if err == nil {
		t.Error("GetHookContent() should return error for non-existent hook")
	}
}

// TestHookManager_ListInstalled tests listing installed hooks
func TestHookManager_ListInstalled(t *testing.T) {
	// Create a temp git repo
	tempDir, err := os.MkdirTemp("", "git-list-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize git repo
	gitDir := filepath.Join(tempDir, ".git")
	hooksDir := filepath.Join(gitDir, "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		t.Fatalf("Failed to create hooks dir: %v", err)
	}

	hm, err := NewHookManager(tempDir)
	if err != nil {
		t.Fatalf("NewHookManager() failed: %v", err)
	}

	// Test with no hooks
	hooks, err := hm.ListInstalled()
	if err != nil {
		t.Errorf("ListInstalled() failed: %v", err)
	}
	if len(hooks) != 0 {
		t.Errorf("Expected 0 hooks, got %d", len(hooks))
	}

	// Create some hooks
	if err := os.WriteFile(filepath.Join(hooksDir, "pre-commit"), []byte("#!/bin/sh\necho"), 0755); err != nil {
		t.Fatalf("Failed to create pre-commit hook: %v", err)
	}
	if err := os.WriteFile(filepath.Join(hooksDir, "pre-push"), []byte("#!/bin/sh\necho"), 0755); err != nil {
		t.Fatalf("Failed to create pre-push hook: %v", err)
	}

	// Test with hooks
	hooks, err = hm.ListInstalled()
	if err != nil {
		t.Errorf("ListInstalled() failed: %v", err)
	}
	if len(hooks) != 2 {
		t.Errorf("Expected 2 hooks, got %d", len(hooks))
	}
}

// TestHookManager_ListInstalled_NoHooksDir tests listing when hooks dir doesn't exist
func TestHookManager_ListInstalled_NoHooksDir(t *testing.T) {
	// Create a temp git repo without hooks dir
	tempDir, err := os.MkdirTemp("", "git-list-nohooks-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize git repo without hooks dir
	gitDir := filepath.Join(tempDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("Failed to create .git dir: %v", err)
	}

	hm, err := NewHookManager(tempDir)
	if err != nil {
		t.Fatalf("NewHookManager() failed: %v", err)
	}

	// List should return nil/empty when hooks dir doesn't exist
	hooks, err := hm.ListInstalled()
	if err != nil {
		t.Errorf("ListInstalled() should not return error: %v", err)
	}
	if hooks != nil && len(hooks) != 0 {
		t.Errorf("Expected nil or empty list, got %d hooks", len(hooks))
	}
}

// TestHookStatus_Struct tests HookStatus struct
func TestHookStatus_Struct(t *testing.T) {
	status := HookStatus{
		Installed: true,
		IsOurs:    true,
		Path:      "/path/to/hook",
		Size:      100,
		ExecMode:  true,
		Error:     "",
	}

	if !status.Installed {
		t.Error("Installed should be true")
	}
	if !status.IsOurs {
		t.Error("IsOurs should be true")
	}
	if status.Path != "/path/to/hook" {
		t.Errorf("Path = %s, want /path/to/hook", status.Path)
	}
	if status.Size != 100 {
		t.Errorf("Size = %d, want 100", status.Size)
	}
	if !status.ExecMode {
		t.Error("ExecMode should be true")
	}
}

// TestFindGitDir tests the findGitDir function
func TestFindGitDir(t *testing.T) {
	// Test with current directory
	gitDir := findGitDir(".")
	if gitDir == "" {
		t.Log("findGitDir() returned empty - not in git repo or test running in non-git context")
	} else {
		t.Logf("findGitDir() found: %s", gitDir)
	}
}

// TestFindGitDir_NonGitDir tests findGitDir with non-git directory
func TestFindGitDir_NonGitDir(t *testing.T) {
	// Create a temp directory without .git
	tempDir, err := os.MkdirTemp("", "non-git-find-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	gitDir := findGitDir(tempDir)
	if gitDir != "" {
		t.Errorf("findGitDir() = %s, want empty string for non-git directory", gitDir)
	}
}

// TestFindGitDir_Subdirectory tests findGitDir from a subdirectory
func TestFindGitDir_Subdirectory(t *testing.T) {
	// Create a temp git repo with subdirectory
	tempDir, err := os.MkdirTemp("", "git-subdir-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	gitDirPath := filepath.Join(tempDir, ".git")
	if err := os.MkdirAll(gitDirPath, 0755); err != nil {
		t.Fatalf("Failed to create .git dir: %v", err)
	}

	// Create subdirectory
	subDir := filepath.Join(tempDir, "subdir", "nested")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}

	// findGitDir should find the .git from subdirectory
	gitDir := findGitDir(subDir)
	if gitDir == "" {
		t.Error("findGitDir() should find .git from subdirectory")
	}
}

// TestFindGitDir_Submodule tests findGitDir with submodule format
func TestFindGitDir_Submodule(t *testing.T) {
	// Create a temp directory structure simulating a submodule
	tempDir, err := os.MkdirTemp("", "git-submodule-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create parent .git directory
	parentGitDir := filepath.Join(tempDir, ".git", "modules", "mysubmodule")
	if err := os.MkdirAll(parentGitDir, 0755); err != nil {
		t.Fatalf("Failed to create parent .git dir: %v", err)
	}

	// Create submodule directory with .git file pointing to parent
	submoduleDir := filepath.Join(tempDir, "submodule")
	if err := os.MkdirAll(submoduleDir, 0755); err != nil {
		t.Fatalf("Failed to create submodule dir: %v", err)
	}

	// Write gitdir reference
	gitFile := filepath.Join(submoduleDir, ".git")
	gitDirRef := "gitdir: " + parentGitDir
	if err := os.WriteFile(gitFile, []byte(gitDirRef), 0644); err != nil {
		t.Fatalf("Failed to create .git file: %v", err)
	}

	// findGitDir should parse the gitdir reference
	gitDir := findGitDir(submoduleDir)
	if gitDir != parentGitDir {
		t.Errorf("findGitDir() = %s, want %s", gitDir, parentGitDir)
	}
}
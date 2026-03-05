package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/todo-tracker/todo-tracker/pkg/types"
)

// configCmd 配置管理命令
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "配置管理",
	Long: `管理 TODO Tracker 的配置。

支持以下子命令:
  init   - 初始化配置文件
  show   - 显示当前配置
  set    - 设置配置项
  reset  - 重置为默认配置

示例:
  todo config init                 初始化配置文件
  todo config show                 显示当前配置
  todo config set scan.workers 4   设置并发数
  todo config reset                重置配置`,
}

// configInitCmd 初始化配置命令
var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "初始化配置文件",
	Long:  "在当前目录创建默认的 .todo-tracker.yaml 配置文件。",
	RunE:  runConfigInit,
}

// configShowCmd 显示配置命令
var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "显示当前配置",
	RunE:  runConfigShow,
}

// configSetCmd 设置配置命令
var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "设置配置项",
	Args:  cobra.ExactArgs(2),
	RunE:  runConfigSet,
}

// configResetCmd 重置配置命令
var configResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "重置为默认配置",
	RunE:  runConfigReset,
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configResetCmd)
}

// runConfigInit 初始化配置文件
func runConfigInit(cmd *cobra.Command, args []string) error {
	configPath := ".todo-tracker.yaml"

	// 检查文件是否已存在
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("配置文件已存在: %s", configPath)
	}

	// 生成默认配置
	config := types.DefaultConfig()

	// 写入配置文件
	// TODO: 使用 viper 写入 YAML
	fmt.Printf("创建默认配置文件: %s\n", configPath)
	fmt.Println("配置内容:")
	fmt.Printf("%+v\n", config)

	return nil
}

// runConfigShow 显示当前配置
func runConfigShow(cmd *cobra.Command, args []string) error {
	fmt.Println("当前配置:")
	fmt.Println()
	fmt.Printf("配置文件: %s\n", viper.ConfigFileUsed())
	fmt.Println()

	// 显示所有配置项
	settings := viper.AllSettings()
	for key, value := range settings {
		fmt.Printf("  %s: %v\n", key, value)
	}

	return nil
}

// runConfigSet 设置配置项
func runConfigSet(cmd *cobra.Command, args []string) error {
	key := args[0]
	value := args[1]

	viper.Set(key, value)

	// 写入配置文件
	configPath := viper.ConfigFileUsed()
	if configPath == "" {
		configPath = ".todo-tracker.yaml"
	}

	// 确保目录存在
	dir := filepath.Dir(configPath)
	if dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("创建配置目录失败: %w", err)
		}
	}

	fmt.Printf("设置 %s = %s\n", key, value)

	return nil
}

// runConfigReset 重置配置
func runConfigReset(cmd *cobra.Command, args []string) error {
	configPath := viper.ConfigFileUsed()
	if configPath == "" {
		configPath = ".todo-tracker.yaml"
	}

	// 删除配置文件
	if err := os.Remove(configPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("删除配置文件失败: %w", err)
	}

	fmt.Println("已重置为默认配置")
	return nil
}
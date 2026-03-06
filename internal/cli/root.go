// Package cli 提供 TODO Tracker 的命令行接口
package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// 版本信息
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// SetVersion 设置版本信息
func SetVersion(v, c, d string) {
	version = v
	commit = c
	date = d
	rootCmd.Version = v
}

// rootCmd 根命令
var rootCmd = &cobra.Command{
	Use:   "todo",
	Short: "智能 TODO 分诊工具",
	Long: `TODO Tracker 是一个智能的 TODO 分诊工具。

它能帮助你从数百个 TODO 中找出本周需要关注的那几个，
通过 Git 历史分析检测过期 TODO 和孤儿 TODO。

主要命令:
  scan      扫描代码库中的 TODO
  list      列出已扫描的 TODO
  stale     检测过期 TODO（>90天）
  orphaned  检测孤儿 TODO（作者已离开）
  report    生成 TODO 报告
  stats     显示统计信息
  config    配置管理
  hooks     Git Hook 管理

示例:
  todo scan                    扫描当前目录
  todo scan ./src --staged     扫描暂存文件
  todo stale --older-than 90   查找90天前的 TODO
  todo orphaned                查找孤儿 TODO
  todo report -f markdown -o TODO.md  生成报告`,
	Version: version,
}

// cfgFile 配置文件路径
var cfgFile string

func init() {
	cobra.OnInitialize(initConfig)

	// 全局标志
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "配置文件路径（默认为 .todo-tracker.yaml）")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "详细输出")
	rootCmd.PersistentFlags().StringP("output", "o", "", "输出文件路径")

	// 绑定到 viper
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output"))
}

// initConfig 初始化配置
func initConfig() {
	if cfgFile != "" {
		// 使用指定的配置文件
		viper.SetConfigFile(cfgFile)
	} else {
		// 查找配置文件
		viper.AddConfigPath(".")
		viper.AddConfigPath("$HOME/.config/todo-tracker")
		viper.SetConfigName(".todo-tracker")
		viper.SetConfigType("yaml")
	}

	viper.AutomaticEnv() // 读取环境变量

	if err := viper.ReadInConfig(); err == nil {
		if viper.GetBool("verbose") {
			fmt.Fprintf(os.Stderr, "使用配置文件: %s\n", viper.ConfigFileUsed())
		}
	}
}

// Execute 执行根命令
func Execute() error {
	return rootCmd.Execute()
}

// GetRootCmd 获取根命令（用于测试）
func GetRootCmd() *cobra.Command {
	return rootCmd
}

// GetVersion 获取版本信息
func GetVersion() string {
	return fmt.Sprintf("TODO Tracker %s (commit: %s, built: %s)", version, commit, date)
}
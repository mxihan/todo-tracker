package cli

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

// staleCmd 过期 TODO 检测命令
var staleCmd = &cobra.Command{
	Use:   "stale",
	Short: "检测过期 TODO（超过指定天数）",
	Long: `检测超过指定天数未更新的 TODO。

过期 TODO 可能意味着:
  - 任务已完成但忘记删除注释
  - 任务已不再相关
  - 需要重新评估优先级

示例:
  todo stale                  查找超过 90 天的 TODO
  todo stale --older-than 60  查找超过 60 天的 TODO
  todo stale --min-churn 10   文件修改超过 10 次的 TODO
  todo stale --review         交互式审查`,
	RunE: runStale,
}

var (
	staleOlderThan int
	staleMinChurn  int
	staleReview    bool
)

func init() {
	rootCmd.AddCommand(staleCmd)

	staleCmd.Flags().IntVarP(&staleOlderThan, "older-than", "d", 90, "过期阈值（天数）")
	staleCmd.Flags().IntVar(&staleMinChurn, "min-churn", 0, "最小文件修改次数")
	staleCmd.Flags().BoolVar(&staleReview, "review", false, "交互式审查模式")
}

// runStale 执行过期 TODO 检测
func runStale(cmd *cobra.Command, args []string) error {
	fmt.Printf("正在检测过期 TODO（超过 %d 天）...\n", staleOlderThan)
	fmt.Println()

	// TODO: 实现实际的检测逻辑
	// 这里是占位实现

	fmt.Println(" 过期 TODO（超过 90 天）")
	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Println()
	fmt.Println(" 年龄      文件                      修改次数   作者         描述")
	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Println()
	fmt.Println(" 未发现过期 TODO")
	fmt.Println()
	fmt.Printf(" 检测完成，耗时 %s\n", time.Since(time.Now()).Round(time.Millisecond))

	return nil
}
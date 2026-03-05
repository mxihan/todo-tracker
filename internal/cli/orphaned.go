package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// orphanedCmd 孤儿 TODO 检测命令
var orphanedCmd = &cobra.Command{
	Use:   "orphaned",
	Short: "检测孤儿 TODO（作者已离开项目）",
	Long: `检测作者已不活跃的 TODO。

孤儿 TODO 是指由已经离开项目或长期不活跃的作者创建的 TODO。
这些 TODO 通常需要重新分配或评估。

示例:
  todo orphaned                  查找孤儿 TODO
  todo orphaned --inactive 180   作者不活跃阈值为 180 天
  todo orphaned --all            包含所有历史作者`,
	RunE: runOrphaned,
}

var (
	orphanedInactive int
	orphanedAll      bool
)

func init() {
	rootCmd.AddCommand(orphanedCmd)

	orphanedCmd.Flags().IntVar(&orphanedInactive, "inactive", 180, "作者不活跃阈值（天数）")
	orphanedCmd.Flags().BoolVar(&orphanedAll, "all", false, "包含所有历史作者")
}

// runOrphaned 执行孤儿 TODO 检测
func runOrphaned(cmd *cobra.Command, args []string) error {
	fmt.Printf("正在检测孤儿 TODO（作者不活跃超过 %d 天）...\n", orphanedInactive)
	fmt.Println()

	// TODO: 实现实际的检测逻辑
	// 这里是占位实现

	fmt.Println(" 孤儿 TODO（作者已离开）")
	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Println()
	fmt.Println(" 作者          最后提交      TODO数量   优先级分布")
	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Println()
	fmt.Println(" 未发现孤儿 TODO")
	fmt.Println()
	fmt.Println(" 建议：运行 `todo orphaned --reassign` 交互式重新分配")

	return nil
}
// Package main 是 TODO Tracker CLI 的入口点
package main

import (
	"os"

	"github.com/mxihan/todo-tracker/internal/cli"
)

// 构建时注入的版本信息
var (
	// Version 版本号
	Version = "dev"
	// Commit Git 提交哈希
	Commit = "none"
	// Date 构建日期
	Date = "unknown"
)

func main() {
	// 设置版本信息
	cli.SetVersion(Version, Commit, Date)

	// 执行根命令
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
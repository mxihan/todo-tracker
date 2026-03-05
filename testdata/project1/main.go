// Package main 演示Go代码中的TODO格式
package main

import (
	"fmt"
	"os"
)

// TODO: 这是基本的TODO格式
func basicFunction() {
	fmt.Println("Hello, World!")
}

// TODO!: 这是一个高优先级的TODO
func urgentFunction() {
	// TODO(@alice): 分配给alice的TODO
	fmt.Println("Urgent task")
}

// FIXME: 这里需要修复一个问题
func brokenFunction() int {
	// TODO(#123): 关联到Issue #123
	return 0
}

// TODO(JIRA-456): 关联Jira工单
func jiraLinkedFunction() {
	// TODO [high]: 显式高优先级
	// TODO:2024-12-31: 带截止日期
}

// HACK: 这是一个临时解决方案，需要改进
func temporarySolution() {
	/*
	   FIXME: 多行注释中的TODO
	   这里详细描述了问题
	*/
}

// BUG: 已知缺陷，需要修复
func buggyFunction() {
	// XXX: 警告性标记
}

// TODO(@bob) #789!: 组合格式 - 分配给bob，关联issue，高优先级
func combinedFormatFunction() {
}

func main() {
	basicFunction()
	urgentFunction()
}
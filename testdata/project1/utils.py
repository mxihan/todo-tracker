#!/usr/bin/env python3
"""
演示Python代码中的TODO格式
"""

# TODO: 基本的Python TODO
def basic_function():
    print("Hello, World!")


# TODO!: 高优先级TODO
def urgent_function():
    # TODO(@alice): 分配给alice
    pass


# FIXME: 需要修复的问题
def broken_function():
    # TODO(#456): 关联Issue
    return 0


# TODO(JIRA-789): 关联Jira工单
def jira_linked_function():
    """
    TODO: 这是文档字符串中的TODO
    """
    pass


# HACK: 临时解决方案
def temporary_solution():
    """
    多行文档字符串
    FIXME: 文档字符串中需要修复的问题
    """
    pass


# BUG: 已知缺陷
def buggy_function():
    # XXX: 警告标记
    pass


# TODO(@bob) #999!: 组合格式
def combined_format_function():
    pass
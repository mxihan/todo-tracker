# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
-

### Changed
-

### Fixed
-

## [0.1.0] - 2026-03-05

### Added
- 初始版本发布
- `todo scan` 命令：扫描代码库中的TODO
- `todo list` 命令：列出已扫描的TODO
- `todo stale` 命令：检测过期TODO
- `todo orphaned` 命令：检测孤儿TODO
- `todo report` 命令：生成报告（支持table/json/markdown/html格式）
- `todo stats` 命令：显示统计信息
- `todo config` 命令：配置管理
- `todo hooks` 命令：Git Hook管理
- Git blame集成：自动获取TODO作者信息
- SQLite缓存：支持增量扫描
- 多语言支持：Go/JavaScript/TypeScript/Python/Java/Rust等
- CI模式：适合CI/CD集成的输出格式
- Homebrew公式：支持brew安装
- Docker支持：提供官方Docker镜像

### Features
- 智能TODO分诊：从数百个TODO中找出需要关注的
- 过期检测：识别超过90天未处理的僵尸TODO
- 孤儿检测：识别作者已离开的TODO
- Git Churn评分：基于文件修改频率评估TODO相关性
- 丰富的注释语法：支持优先级、指派、工单关联等元数据

### Technical
- 使用Go 1.21+开发
- Cobra CLI框架
- SQLite存储
- 支持Linux/macOS/Windows
- 支持amd64/arm64架构

---

## 版本说明

### 版本命名规则

- **主版本号 (Major)**: 不兼容的API变更
- **次版本号 (Minor)**: 向后兼容的功能新增
- **修订号 (Patch)**: 向后兼容的问题修复

### 发布周期

- **Patch版本**: 根据需要随时发布
- **Minor版本**: 每2-4周发布
- **Major版本**: 根据需要发布，提前公告

### 支持策略

- 当前主版本：完全支持
- 前一主版本：仅安全更新
- 更早版本：不再支持
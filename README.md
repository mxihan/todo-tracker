# TODO Tracker CLI

[English](#english) | [中文](#中文)

---

<a name="中文"></a>

## 中文

### 简介

**TODO Tracker CLI** 是一个智能的TODO分诊工具，帮助开发团队从数百个TODO中找出真正需要关注的那几个。

与简单的TODO扫描器不同，TODO Tracker关注：

- **过期检测**: 找出超过90天未处理的僵尸TODO
- **孤儿检测**: 识别作者已离开的TODO，避免代码交接盲点
- **Git Churn评分**: 基于文件修改频率识别可能过时的TODO

### 安装

#### Homebrew (推荐)

```bash
brew tap mxihan/tap
brew install todo-tracker
```

#### 二进制下载

从 [GitHub Releases](https://github.com/mxihan/todo-tracker/releases) 下载对应平台的二进制文件：

```bash
# Linux/macOS
curl -sSL https://github.com/mxihan/todo-tracker/releases/latest/download/todo-tracker-$(uname -s)-$(uname -m) -o todo
chmod +x todo
sudo mv todo /usr/local/bin/

# Windows (PowerShell)
Invoke-WebRequest -Uri "https://github.com/mxihan/todo-tracker/releases/latest/download/todo-tracker-windows-amd64.exe" -OutFile "todo.exe"
```

#### 从源码构建

```bash
git clone https://github.com/mxihan/todo-tracker.git
cd todo-tracker
go build -o todo ./cmd/todo
```

### 快速开始

#### 1. 初始化配置

```bash
# 在项目根目录创建配置文件
todo config init

# 查看当前配置
todo config show
```

#### 2. 扫描代码库

```bash
# 扫描当前目录
todo scan

# 扫描指定目录
todo scan ./src

# 仅扫描暂存文件（pre-commit场景）
todo scan --staged

# 扫描特定commit后的变更
todo scan --since HEAD~10
```

#### 3. 检查过期TODO

```bash
# 找出超过90天的TODO
todo stale

# 自定义过期阈值
todo stale --older-than 180d

# 交互式审查
todo stale --review
```

#### 4. 检查孤儿TODO

```bash
# 找出作者已离开的TODO
todo orphaned

# 自定义不活跃阈值
todo orphaned --inactive 365d
```

#### 5. 生成报告

```bash
# Markdown格式
todo report --format markdown --output TODO.md

# JSON格式（用于CI集成）
todo report --format json --output report.json

# HTML报告
todo report --format html --output report.html
```

### 命令文档

#### `todo scan`

扫描代码库中的TODO。

```bash
todo scan [path] [flags]

Flags:
  --staged          仅扫描Git暂存文件
  --since <ref>     扫描指定commit后的变更
  --watch           监视模式，文件变更时自动扫描
  --ci              CI模式，输出适合CI的格式
  --fail-on <types> 遇到指定类型时返回非零退出码
```

#### `todo list`

列出已扫描的TODO。

```bash
todo list [flags]

Flags:
  --priority <level>  按优先级过滤 (high/medium/low)
  --author <name>     按作者过滤
  --type <type>       按类型过滤 (TODO/FIXME/HACK/BUG)
  --status <status>   按状态过滤 (open/resolved/wontfix)
```

#### `todo stale`

检测过期TODO。

```bash
todo stale [flags]

Flags:
  --older-than <days>  过期阈值天数（默认90）
  --min-churn <count>  最小文件修改次数
  --review             交互式审查模式
```

#### `todo orphaned`

检测孤儿TODO（作者已离开）。

```bash
todo orphaned [flags]

Flags:
  --inactive <days>  作者不活跃阈值（默认180天）
  --all              包含所有历史作者
```

#### `todo report`

生成报告。

```bash
todo report [flags]

Flags:
  --format <fmt>    输出格式 (table/json/markdown/html)
  --output <file>   输出文件路径
  --stale-only      仅包含过期TODO
  --orphan-only     仅包含孤儿TODO
```

#### `todo stats`

显示统计信息。

```bash
todo stats [flags]

Flags:
  --by-author      按作者统计
  --by-component   按组件统计
  --trend          显示趋势
```

#### `todo config`

配置管理。

```bash
todo config init              # 初始化配置文件
todo config show              # 显示当前配置
todo config set <key> <value> # 设置配置项
todo config reset              # 重置为默认配置
```

#### `todo hooks`

Git Hook管理。

```bash
todo hooks install    # 安装Git hooks
todo hooks uninstall  # 卸载Git hooks
todo hooks check      # 检查hook状态
```

### 配置说明

配置文件 `.todo-tracker.yaml` 示例：

```yaml
# 版本
version: 1

# 扫描配置
scan:
  paths:
    - src/
    - lib/
  exclude:
    - "**/node_modules/**"
    - "**/vendor/**"
    - "**/*.min.js"
    - "**/dist/**"
  workers: 0  # 0 = 自动检测CPU核心数

# TODO模式识别
patterns:
  types:
    - TODO
    - FIXME
    - HACK
    - XXX
    - BUG

  # 优先级标记
  priority_markers:
    high: ["!", "URGENT", "CRITICAL"]
    medium: [">", "MEDIUM"]
    low: []

  # 元数据提取
  metadata:
    assignee_pattern: '\(([^)]+)\)|@(\w+)'
    ticket_pattern: '#(\d+)|([A-Z]+-\d+)'
    date_pattern: '\[(\d{4}-\d{2}-\d{2})\]'

# Git集成
git:
  enabled: true
  blame: true
  default_base: main

# 过期检测
stale:
  threshold_days: 90
  churn_threshold: 10

# 孤儿检测
orphan:
  inactive_days: 180

# CI配置
ci:
  fail_on_stale: false
  fail_on_orphan: false
  max_stale_days: 180
```

### TODO注释语法

TODO Tracker支持丰富的注释语法：

```go
// TODO: 基本格式
// TODO!: 高优先级（感叹号）
// TODO(@alice): 分配给alice
// TODO(#123): 关联Issue #123
// TODO(JIRA-456): 关联Jira工单
// TODO [high]: 显式优先级
// TODO:2024-12-31: 带截止日期
// TODO(@alice) #123!: 组合格式

/* FIXME: 多行注释
   详细描述问题 */

// BUG: 已知问题
// HACK: 临时方案
// XXX: 危险代码
```

### CI/CD集成

#### GitHub Actions

```yaml
# .github/workflows/todo-check.yml
name: TODO检查

on:
  push:
    branches: [main]
  pull_request:

jobs:
  check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: 安装todo-tracker
        run: |
          curl -sSL https://github.com/mxihan/todo-tracker/releases/latest/download/todo-tracker-linux-amd64 -o todo
          chmod +x todo

      - name: 检查过期TODO
        run: ./todo stale --ci --fail-on=180d

      - name: 检查孤儿TODO
        run: ./todo orphaned --ci --fail-count=5
```

#### Git Hooks

```bash
# 安装pre-commit和pre-push hooks
todo hooks install

# 或手动复制
cp scripts/pre-commit .git/hooks/
cp scripts/pre-push .git/hooks/
chmod +x .git/hooks/*
```

---

<a name="english"></a>

## English

### Introduction

**TODO Tracker CLI** is an intelligent TODO triage tool that helps development teams find the TODOs that actually matter from hundreds of items.

Unlike simple TODO scanners, TODO Tracker focuses on:

- **Stale Detection**: Find zombie TODOs that haven't been addressed for over 90 days
- **Orphan Detection**: Identify TODOs whose authors have left, avoiding blind spots during code handoffs
- **Git Churn Score**: Identify potentially outdated TODOs based on file modification frequency

### Installation

#### Homebrew (Recommended)

```bash
brew tap mxihan/tap
brew install todo-tracker
```

#### Binary Download

Download the binary for your platform from [GitHub Releases](https://github.com/mxihan/todo-tracker/releases):

```bash
# Linux/macOS
curl -sSL https://github.com/mxihan/todo-tracker/releases/latest/download/todo-tracker-$(uname -s)-$(uname -m) -o todo
chmod +x todo
sudo mv todo /usr/local/bin/

# Windows (PowerShell)
Invoke-WebRequest -Uri "https://github.com/mxihan/todo-tracker/releases/latest/download/todo-tracker-windows-amd64.exe" -OutFile "todo.exe"
```

#### Build from Source

```bash
git clone https://github.com/mxihan/todo-tracker.git
cd todo-tracker
go build -o todo ./cmd/todo
```

### Quick Start

#### 1. Initialize Configuration

```bash
# Create configuration file in project root
todo config init

# View current configuration
todo config show
```

#### 2. Scan Codebase

```bash
# Scan current directory
todo scan

# Scan specific directory
todo scan ./src

# Scan staged files only (pre-commit scenario)
todo scan --staged

# Scan changes after a specific commit
todo scan --since HEAD~10
```

#### 3. Check Stale TODOs

```bash
# Find TODOs older than 90 days
todo stale

# Custom stale threshold
todo stale --older-than 180d

# Interactive review
todo stale --review
```

#### 4. Check Orphaned TODOs

```bash
# Find TODOs whose authors have left
todo orphaned

# Custom inactive threshold
todo orphaned --inactive 365d
```

#### 5. Generate Reports

```bash
# Markdown format
todo report --format markdown --output TODO.md

# JSON format (for CI integration)
todo report --format json --output report.json

# HTML report
todo report --format html --output report.html
```

### Configuration

See the Chinese section above for a complete configuration example.

### License

MIT License - see [LICENSE](LICENSE) for details.

### Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.
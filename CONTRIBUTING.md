# 贡献指南

感谢您有兴趣为TODO Tracker做出贡献！本文档将帮助您了解如何参与项目开发。

## 目录

- [行为准则](#行为准则)
- [如何贡献](#如何贡献)
- [开发环境设置](#开发环境设置)
- [代码规范](#代码规范)
- [提交规范](#提交规范)
- [Pull Request流程](#pull-request流程)
- [问题报告](#问题报告)
- [功能请求](#功能请求)

## 行为准则

本项目采用贡献者公约作为行为准则。参与本项目即表示您同意遵守其条款。请阅读 [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) 了解详情。

## 如何贡献

### 报告Bug

如果您发现了bug，请通过[GitHub Issues](https://github.com/your-org/todo-tracker/issues)提交报告。提交前请：

1. 搜索现有issues，确认没有重复报告
2. 使用Bug报告模板
3. 提供尽可能详细的信息：
   - 操作系统和版本
   - TODO Tracker版本
   - 复现步骤
   - 预期行为和实际行为
   - 相关日志或截图

### 提交功能请求

欢迎提出新功能建议！请：

1. 先搜索现有issues，确认没有类似请求
2. 使用功能请求模板
3. 清楚描述功能需求和使用场景

### 提交代码

1. Fork本仓库
2. 创建功能分支 (`git checkout -b feature/amazing-feature`)
3. 进行修改
4. 提交更改 (`git commit -m 'feat: add amazing feature'`)
5. 推送到分支 (`git push origin feature/amazing-feature`)
6. 创建Pull Request

## 开发环境设置

### 前置要求

- Go 1.21 或更高版本
- Git
- Make (可选)

### 克隆仓库

```bash
git clone https://github.com/your-org/todo-tracker.git
cd todo-tracker
```

### 安装依赖

```bash
go mod download
```

### 构建项目

```bash
# 使用Go直接构建
go build -o bin/todo ./cmd/todo

# 或使用Makefile
make build
```

### 运行测试

```bash
# 运行所有测试
go test ./...

# 运行测试并显示覆盖率
go test -cover ./...

# 运行特定测试
go test -run TestScanCommand ./internal/cli/...
```

### 代码检查

```bash
# 安装golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# 运行lint
golangci-lint run
```

## 代码规范

### Go代码规范

- 遵循 [Effective Go](https://golang.org/doc/effective_go) 指南
- 使用 `gofmt` 格式化代码
- 使用 `goimports` 管理导入
- 添加必要的注释，特别是导出的函数和类型

### 项目结构

```
todo-tracker/
├── cmd/                    # 命令入口
│   └── todo/
│       └── main.go
├── internal/               # 内部实现（不对外暴露）
│   ├── cli/               # CLI命令
│   ├── scanner/           # 文件扫描
│   ├── parser/            # TODO解析
│   ├── git/               # Git集成
│   ├── cache/             # 缓存实现
│   └── reporter/          # 报告生成
├── pkg/                    # 公共库（可对外暴露）
│   └── types/
├── configs/               # 配置文件
├── scripts/               # 脚本文件
├── .github/               # GitHub配置
└── docs/                  # 文档
```

### 命名约定

- 包名：小写单词，不使用下划线
- 文件名：小写，使用下划线分隔
- 函数/变量：驼峰命名
- 常量：驼峰命名，导出常量可使用全大写
- 接口：以 `-er` 结尾（如 `Scanner`, `Parser`）

### 错误处理

```go
// 好的做法：包装错误，提供上下文
if err := scanner.Scan(); err != nil {
    return fmt.Errorf("failed to scan files: %w", err)
}

// 避免：忽略错误
scanner.Scan() // 错误！
```

### 测试规范

- 测试文件命名为 `*_test.go`
- 测试函数命名为 `Test<FunctionName>`
- 使用表驱动测试
- 测试覆盖率目标：80%+

```go
func TestScanCommand(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    int
        wantErr bool
    }{
        {
            name:  "basic scan",
            input: "testdata/basic",
            want:  3,
            wantErr: false,
        },
        // ... 更多测试用例
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // 测试逻辑
        })
    }
}
```

## 提交规范

本项目使用 [Conventional Commits](https://www.conventionalcommits.org/) 规范。

### 提交消息格式

```
<type>(<scope>): <description>

[optional body]

[optional footer(s)]
```

### 类型 (type)

- `feat`: 新功能
- `fix`: Bug修复
- `docs`: 文档更新
- `style`: 代码格式（不影响功能）
- `refactor`: 重构
- `perf`: 性能优化
- `test`: 测试相关
- `chore`: 构建/工具相关
- `ci`: CI配置相关

### 示例

```
feat(scan): add support for staged files scanning

Add --staged flag to scan command to only scan files that are
staged in git. This is useful for pre-commit hooks.

Closes #123
```

```
fix(parser): handle multi-line TODOs correctly

Multi-line TODOs were being truncated. Now they are parsed
correctly with the line_end field properly set.
```

## Pull Request流程

1. **创建分支**
   ```bash
   git checkout -b feature/your-feature
   ```

2. **进行修改**
   - 遵循代码规范
   - 添加必要的测试
   - 更新相关文档

3. **本地测试**
   ```bash
   make test
   make lint
   make build
   ```

4. **提交PR**
   - 填写PR模板
   - 关联相关Issue
   - 等待CI通过

5. **代码审查**
   - 响应审查意见
   - 进行必要修改
   - 保持提交历史整洁

6. **合并**
   - 至少需要1个批准
   - CI全部通过
   - 由维护者合并

## 问题报告

如果您在使用过程中遇到问题：

1. 检查 [FAQ](docs/FAQ.md)
2. 搜索 [Issues](https://github.com/your-org/todo-tracker/issues)
3. 提交新Issue，包含：
   - 问题描述
   - 复现步骤
   - 环境信息
   - 日志/截图

## 功能请求

我们欢迎新功能建议！请：

1. 描述功能需求和使用场景
2. 说明为什么这个功能对项目有价值
3. 如果可能，提供实现思路

---

再次感谢您的贡献！如有任何问题，欢迎在[Issues](https://github.com/your-org/todo-tracker/issues)中提问。
# Homebrew Formula for TODO Tracker
#
# 使用方法:
# 1. 将此文件放入 your-org/homebrew-tap/Formula/todo-tracker.rb
# 2. 更新 url 和 checksum 为实际值
#
# 安装命令:
#   brew tap your-org/tap
#   brew install todo-tracker

class TodoTracker < Formula
  desc "Intelligent TODO triage tool for codebases"
  homepage "https://github.com/your-org/todo-tracker"
  version "0.1.0"
  license "MIT"

  # 在发布新版本时更新这些URL
  if OS.mac? && Hardware::CPU.arm?
    url "https://github.com/your-org/todo-tracker/releases/download/v#{version}/todo-tracker-darwin-arm64.tar.gz"
    sha256 "ARM64_MACOS_CHECKSUM_PLACEHOLDER"
  elsif OS.mac? && Hardware::CPU.intel?
    url "https://github.com/your-org/todo-tracker/releases/download/v#{version}/todo-tracker-darwin-amd64.tar.gz"
    sha256 "AMD64_MACOS_CHECKSUM_PLACEHOLDER"
  elsif OS.linux? && Hardware::CPU.arm?
    url "https://github.com/your-org/todo-tracker/releases/download/v#{version}/todo-tracker-linux-arm64.tar.gz"
    sha256 "ARM64_LINUX_CHECKSUM_PLACEHOLDER"
  elsif OS.linux? && Hardware::CPU.intel?
    url "https://github.com/your-org/todo-tracker/releases/download/v#{version}/todo-tracker-linux-amd64.tar.gz"
    sha256 "AMD64_LINUX_CHECKSUM_PLACEHOLDER"
  end

  head do
    url "https://github.com/your-org/todo-tracker.git", branch: "main"
    depends_on "go" => :build
  end

  # 可选依赖
  depends_on "git" => :recommended

  def install
    if build.head?
      # 从源码构建
      system "go", "build", *std_go_args(ldflags: "-s -w -X main.version=#{version}"), "./cmd/todo"
    else
      # 使用预编译二进制
      bin.install "todo-tracker-#{os_name}-#{arch_name}" => "todo"
    end

    # 安装shell补全
    generate_completions_from_executable(bin/"todo", "completion")

    # 安装手册页（如果有）
    # man1.install "docs/man/todo.1"
  end

  test do
    # 测试版本命令
    assert_match version.to_s, shell_output("#{bin}/todo --version")

    # 创建测试文件
    (testpath/"test.go").write <<~EOS
      package main

      // TODO: this is a test todo
      func main() {
          // FIXME: this needs fixing
          println("hello")
      }
    EOS

    # 测试扫描命令
    output = shell_output("#{bin}/todo scan #{testpath}")
    assert_match "TODO", output
    assert_match "FIXME", output
  end

  private

  def os_name
    return "darwin" if OS.mac?
    return "linux" if OS.linux?
    "unknown"
  end

  def arch_name
    return "arm64" if Hardware::CPU.arm?
    return "amd64" if Hardware::CPU.intel?
    "unknown"
  end
end

# 发布新版本时的更新步骤:
#
# 1. 构建并上传发布资产到GitHub Releases
# 2. 计算每个资产的SHA256校验和:
#    shasum -a 256 todo-tracker-darwin-arm64.tar.gz
# 3. 更新上述URL和checksum
# 4. 更新version
# 5. 提交PR到homebrew-tap仓库
#
# 自动化更新（使用brew bump）:
#   brew bump-formula-pr todo-tracker --url https://github.com/your-org/todo-tracker/releases/download/v0.2.0/todo-tracker-darwin-arm64.tar.gz
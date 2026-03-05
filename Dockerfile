# Build stage
FROM golang:1.26-alpine AS builder

# 安装构建依赖
RUN apk add --no-cache git ca-certificates tzdata

# 设置工作目录
WORKDIR /build

# 复制go.mod和go.sum先下载依赖（利用缓存）
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建二进制
ARG VERSION=dev
ARG COMMIT=unknown
ARG DATE=unknown

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}" \
    -o todo \
    ./cmd/todo

# Final stage
FROM alpine:3.19

# 安装运行时依赖
RUN apk add --no-cache ca-certificates tzdata git

# 创建非root用户
RUN addgroup -S todogroup && adduser -S todouser -G todogroup

# 设置工作目录
WORKDIR /app

# 从builder复制二进制
COPY --from=builder /build/todo /usr/local/bin/todo

# 复制默认配置
COPY configs/.todo-tracker.yaml /etc/todo-tracker/config.yaml

# 设置权限
RUN chmod +x /usr/local/bin/todo && \
    chown -R todouser:todogroup /app

# 切换到非root用户
USER todouser

# 设置环境变量
ENV TODO_CONFIG=/etc/todo-tracker/config.yaml

# 入口点
ENTRYPOINT ["todo"]
CMD ["--help"]

# Labels
LABEL org.opencontainers.image.title="TODO Tracker"
LABEL org.opencontainers.image.description="Intelligent TODO triage tool for codebases"
LABEL org.opencontainers.image.source="https://github.com/your-org/todo-tracker"
LABEL org.opencontainers.image.licenses="MIT"
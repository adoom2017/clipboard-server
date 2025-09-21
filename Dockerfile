# Go Server Build Stage - 使用 Ubuntu 基础镜像以避免 Alpine 的 SQLite3 编译问题
FROM golang:1.21-bullseye AS builder

# 设置工作目录
WORKDIR /app

# 更新包列表并安装必要的包
RUN sed -i 's/deb.debian.org/mirrors.aliyun.com/g' /etc/apt/sources.list \
    && apt-get update && apt-get install -y \
    gcc \
    libc6-dev \
    sqlite3 \
    libsqlite3-dev \
    && rm -rf /var/lib/apt/lists/*

# 复制 go mod 文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 设置环境变量
ENV CGO_ENABLED=1
ENV GOOS=linux

# 构建应用
RUN go build -a -ldflags="-s -w" -o main .

# Production Stage
FROM debian:bullseye-slim

# 安装必要的运行时包
RUN sed -i 's/deb.debian.org/mirrors.aliyun.com/g' /etc/apt/sources.list \
    && apt-get update && apt-get install -y \
    ca-certificates \
    sqlite3 \
    wget \
    && rm -rf /var/lib/apt/lists/*

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/main .

# 复制启动脚本
COPY docker-entrypoint.sh .
RUN chmod +x docker-entrypoint.sh

# 创建必要的目录
RUN mkdir -p data logs uploads

# 暴露端口
EXPOSE 8080

# 设置环境变量
ENV GIN_MODE=release
ENV GO_ENV=production

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# 使用启动脚本
ENTRYPOINT ["./docker-entrypoint.sh"]
CMD ["./main"]
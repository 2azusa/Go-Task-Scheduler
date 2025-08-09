# 阶段 1: 构建环境
# 使用基于 Alpine Linux 的 Go 镜像作为构建器
FROM golang:1.19-alpine AS builder

# 在容器内设置工作目录
WORKDIR /app

# 复制 go.mod 和 go.sum 文件，并下载依赖项
# 这一步可以利用 Docker 的层缓存机制
COPY go.mod go.sum ./
RUN go mod download

# 复制所有源代码
COPY . .

# 编译应用
# CGO_ENABLED=0 禁用 Cgo，以构建静态链接的二进制文件
# GOOS=linux 指定目标操作系统为 Linux
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# 阶段 2: 生产环境
# 使用一个非常轻量级的 Alpine 镜像作为基础
FROM alpine:latest

# 设置工作目录
WORKDIR /root/

# 从构建环境中复制编译好的二进制文件
COPY --from=builder /app/main .

# 暴露应用程序正在监听的端口 (根据您的配置是 8089)
EXPOSE 8089

# 启动容器时运行的命令
CMD ["./main"]
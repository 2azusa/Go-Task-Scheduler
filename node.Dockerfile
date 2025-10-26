# --- 第 1 阶段: 构建阶段 ---
# 使用官方的 Golang 镜像作为构建环境
FROM golang:1.21-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制 Go 模块文件并下载依赖
# 这一步可以利用 Docker 的层缓存，只有在 go.mod/go.sum 变化时才重新下载依赖
COPY go.mod go.sum ./
RUN go mod download

# 复制整个项目的源代码
COPY . .

# 编译 node 应用
# -o 指定输出文件路径
# CGO_ENABLED=0 和 -ldflags '-s -w' 是为了生成静态链接的、体积更小的可执行文件
RUN CGO_ENABLED=0 go build -ldflags '-s -w' -o /app/bin/pulsenode ./node/cmd/main.go


# --- 第 2 阶段: 运行阶段 ---
# 使用一个非常小的基础镜像，比如 alpine
FROM alpine:latest

# 设置工作目录
WORKDIR /app

# 从构建阶段复制编译好的可执行文件
COPY --from=builder /app/bin/pulsenode .
# 复制配置文件模板目录。我们将在 docker-compose 中覆盖它，但把它包含进来是个好习惯。
COPY node/conf ./node/conf

# 暴露节点可能需要监听的端口（如果节点本身是HTTP服务器）
# 这里假设节点不直接对外提供服务，所以是可选的
# EXPOSE 3001 

# 容器启动时运行的命令
# 我们将使用 docker-compose 来覆盖这个命令，以传递不同的配置文件名
ENTRYPOINT ["/app/pulsenode"]
# syntax=docker/dockerfile:1
# 本 Dockerfile 采用多阶段构建：
# 1) 在 golang 官方镜像中编译 Go 可执行文件（带缓存，体积大但速度快）。
# 2) 将编译产物复制到精简的运行镜像中，得到体积极小的最终镜像。

# --- Builder stage ---
# 构建阶段：使用 Alpine 版 Go 1.21
FROM golang:1.21-alpine AS builder
WORKDIR /src

# 仅拷贝依赖声明文件，优先下载依赖以最大化利用构建缓存
COPY go.mod go.sum ./
# 预拉取依赖，后续源代码变更不会导致重复下载
RUN go mod download

# 再拷贝整个项目源码
COPY . .

# 关闭 CGO，编译为静态 Linux 可执行文件；默认架构跟随构建镜像（常见为 amd64/arm64）
ENV CGO_ENABLED=0 GOOS=linux
# 编译根模块（main.go）输出到 /out
RUN go build -o /out/helloworld ./

# --- 运行阶段 ---
# 体积约 7MB 的精简运行镜像
FROM alpine:3.19
WORKDIR /app
# 仅复制编译好的二进制
COPY --from=builder /out/helloworld /app/helloworld
# 声明容器对外提供的端口（文档用途）
EXPOSE 8080
# 容器启动后执行的命令
ENTRYPOINT ["/app/helloworld"]

# 本地构建：
#   docker build -t helloworld:local .
# 本地运行（验证）：
#   docker run --rm -p 8080:8080 helloworld:local
#   curl http://localhost:8080/

package main // 声明可独立运行的可执行程序；包名必须为 main 才能编译为二进制

import (
	"net/http" // 标准库：HTTP 常量与工具（这里用于 http.StatusOK）

	"github.com/gin-gonic/gin" // 第三方 Web 框架 Gin，用于快速构建 HTTP 服务
)

func main() {
	// 创建 Gin 默认引擎：包含 Logger（请求日志）与 Recovery（崩溃恢复）中间件
	// 等价于手动添加 gin.New() + r.Use(gin.Logger(), gin.Recovery())
	r := gin.Default()

	// 注册一个 GET 路由：当访问根路径 "/" 时返回纯文本 "hello world"
	// c.String 会设置 Content-Type 为 text/plain，并写入状态码 200
	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "hello world")
	})

	// 启动 HTTP 服务并监听 0.0.0.0:8080（容器/本机默认端口）
	// 可改为 r.Run(":80") 或从环境变量读取端口以适配不同平台
	if err := r.Run(":8080"); err != nil {
		// 若端口被占用或监听失败，将在此处报错退出
		panic(err)
	}
}

# Asynq 最小 MVP 示例

本示例演示如何将 [Asynq](https://github.com/hibiken/asynq) 融合到 Gin 服务中：服务启动时内置 worker，生产者通过 HTTP 接口注册 Scheduler 定时任务。

## 目录
- `main.go`：主服务入口（Gin + Asynq 集成）。
- `asynq/taskqueue.go`：封装 Asynq 服务器、调度器和任务注册逻辑。

## 前置条件
- 可用的 Redis（默认 `127.0.0.1:6379`）。
- Go 环境可拉取依赖（需网络访问）。

## 安装依赖
如需运行示例，请先拉取 Asynq 依赖（示例使用 v0.24.1，可替换为更新版本）：
```bash
go get github.com/hibiken/asynq@v0.24.1
```
（无外网时可使用内网代理或预下载依赖，然后 `go mod tidy`。）

## 运行示例（Gin + HTTP 生产者）
1) 启动服务（内置 worker 与 HTTP 接口）：
```bash
REDIS_ADDR=127.0.0.1:6379 go run .
```

2) 通过 HTTP 注册定时任务（生产者）：
```bash
curl -X POST http://localhost:8080/tasks/welcome \
  -H "Content-Type: application/json" \
  -d '{"user_id":123,"email":"user@example.com","cron":"*/1 * * * *"}'
```
成功会返回 scheduler entry_id、cron 表达式和任务类型；若未提供 `cron` 字段，默认每分钟（`*/1 * * * *`）。

## 工作原理（示例）
- 任务类型：`email:welcome`，载荷包含 `user_id`、`email`。
- 生产者：HTTP 路由 `/tasks/welcome` 将请求体转为任务，并通过 Asynq Scheduler 注册定时任务（默认每分钟执行一次，支持自定义 5 字段 cron）。
- 消费者：服务启动时注册 handler，并以单并发在后台消费，日志模拟“发送欢迎邮件”。

## 常见问题
- **依赖下载失败**：需联网或使用内网代理，手动将 Asynq 依赖加入 `go.mod` 后再 `go mod tidy`。
- **Redis 无法连接**：确认地址/端口或鉴权参数，必要时调整环境变量 `REDIS_ADDR`。
- **未看到日志**：确保服务已启动且 Redis 可连接；日志输出在服务标准输出中。

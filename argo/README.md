# Argo 使用示例

本文档提供一个最小的 Argo Workflows 案例，帮助在本项目中快速体验 Argo 的工作流能力。

## 前置条件
- 已有可访问的 Kubernetes 集群，`kubectl` 已配置好上下文。
- 推荐在独立命名空间（如 `argo`）中部署 Argo 组件和运行工作流。

## 安装 Argo Workflows（示例）
1. 创建命名空间：
   ```bash
   kubectl create namespace argo
   ```
2. 使用官方发布清单安装（需可访问 GitHub Releases）：
   ```bash
   kubectl apply -n argo -f \
     https://github.com/argoproj/argo-workflows/releases/latest/download/install.yaml
   ```
   - 如需指定版本，可将 `latest` 替换为具体版本号（例如 `v3.5.3`）。
   - 无外网环境可预先下载离线 YAML 再应用，或改用企业镜像仓库。
3. 可选：安装 CLI（视环境自行下载二进制并加入 `PATH`）。

安装完成后，可通过 `kubectl -n argo get pods` 确认控制器、服务器等组件已就绪。

## 目录说明
- `workflow-hello.yaml`：一个两步式的示例工作流，展示串行步骤与简单容器任务。

## 示例工作流说明
工作流包含两个步骤：
1. `greet`（模板 `say-hello`）：打印欢迎语和当前时间。
2. `confirm`（模板 `second-step`）：打印完成信息并输出运行所在节点主机名。
3. 已设置 `ttlSecondsAfterFinished: 300`，工作流完成后约 5 分钟自动清理相关 Pod/Workflow 资源。

## 使用步骤
1. 确认 Argo Workflows 控制器已在集群中运行：
   - 控制器通常位于命名空间 `argo`，Pod 名形如 `workflow-controller-...`。
   - Web UI 可通过 `kubectl port-forward -n argo svc/argo-server 2746:2746` 访问。
2. 部署示例工作流：
   ```bash
   kubectl apply -n argo -f argo/workflow-hello.yaml
   ```
   （如果 Argo 使用的命名空间不是 `default`，在文件中添加 `namespace: <your-ns>` 或使用 `-n <your-ns>` 参数。）
3. 提交并观察运行：
   ```bash
   # 方式 A：使用 Argo CLI
   argo submit argo/workflow-hello.yaml --watch

   # 方式 B：使用 kubectl
   kubectl create -n argo -f argo/workflow-hello.yaml
   kubectl get wf
   kubectl get pods -n argo -l example=hello-argo
   # 选取实际 Pod 名查看日志
   kubectl logs -n argo -f <pod-name> -c main
   ```
4. 结果查看：
   - 当两个步骤均成功后，工作流状态应为 `Succeeded`。
   - 日志中可看到欢迎语、时间和节点主机名输出。

## 清理
```bash
kubectl delete -f argo/workflow-hello.yaml
```
（已配置 TTL，工作流完成后约 5 分钟会自动清理，无需手动删除；若需立即清理可执行上方命令。）

## 常见问题
- 安装未完成或 CRD 缺失：`kubectl apply` 时报 `no matches for kind Workflow`，请先安装 Argo Workflows。
- 权限问题：若使用自定义命名空间，确保对应 ServiceAccount 及 RBAC 已配置允许创建 Pod。
- 镜像拉取失败：可将 `alpine:3.19` 替换为已在集群可用的镜像仓库地址。

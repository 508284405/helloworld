# Argo 使用示例

本文档提供一个最小的 Argo Workflows 案例，帮助在本项目中快速体验 Argo 的工作流能力。

## 前置条件
- 已有可访问的 Kubernetes 集群，`kubectl` 已配置好上下文。
- 推荐在独立命名空间（如 `argo`）中部署 Argo 组件和运行工作流。

## 安装与访问 Argo CD（图形化管理）
1. 创建命名空间并安装（官方快速安装清单，需外网）：
   ```bash
   kubectl create namespace argocd
   kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml
   ```
2. 等待组件就绪：
   ```bash
   kubectl -n argocd get pods
   ```
   所有 Pod 处于 `Running`/`Ready` 后继续。
3. 本地访问 UI（临时端口转发）：
   ```bash
   kubectl -n argocd port-forward svc/argocd-server 8080:80
   ```
   浏览器打开 `http://localhost:8080`。
4. 初始账号与密码：
   - 账号：`admin`
   - 密码：通过 secret 获取（每次安装可能不同）：
     ```bash
     kubectl -n argocd get secret argocd-initial-admin-secret \
       -o jsonpath="{.data.password}" | base64 -d && echo
     ```
5. 可选对外暴露：
   - 若集群支持 LoadBalancer/Ingress，可直接暴露 `argocd-server`，使用外网地址或域名访问。
   - 生产环境请替换默认密码并开启 HTTPS/TLS。
6. 长期访问（推荐使用 Ingress/LB，避免依赖本地 port-forward）：
   - **LoadBalancer（简便）**：如果集群支持，给 `argocd-server` 补充类型或创建单独 Service。
     ```bash
     kubectl -n argocd patch svc argocd-server -p '{"spec":{"type":"LoadBalancer"}}'
     kubectl -n argocd get svc argocd-server   # 查看 EXTERNAL-IP
     ```
   - **Ingress（可复用域名/证书）**：示例清单，按需修改域名/证书/注解（以下示例启用 TLS，使用已有的 `argocd-tls` secret）。
     ```yaml
     # argo/argocd-ingress.yaml
     apiVersion: networking.k8s.io/v1
     kind: Ingress
     metadata:
       name: argocd-server
       namespace: argocd
       annotations:
         nginx.ingress.kubernetes.io/backend-protocol: "HTTPS"
         nginx.ingress.kubernetes.io/ssl-redirect: "true"
     spec:
       tls:
       - hosts:
         - argocd.example.com
         secretName: argocd-tls
       rules:
       - host: argocd.example.com
         http:
           paths:
           - path: /
             pathType: Prefix
             backend:
               service:
                 name: argocd-server
                 port:
                   number: 443
     ```
     应用并验证：
     ```bash
     kubectl apply -f argo/argocd-ingress.yaml
     kubectl -n argocd get ingress argocd-server
     # 将域名解析到 Ingress Controller 的对外地址，然后用浏览器访问 https://argocd.example.com
     ```
   - 若需自签或使用 ACME，可用 cert-manager 管理证书；或在 Ingress Controller 层统一终止 TLS。

## 安装与访问 Argo Workflows
本项目的示例工作流依赖 Argo Workflows（不是 Argo CD）。Argo Workflows 主要包含：
- `workflow-controller`：控制器
- `argo-server`：Web UI + API Server

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

### （可选）免登录访问 UI（仅建议本地/开发环境）
Argo Workflows v3 默认使用 `--auth-mode=client`，UI 会要求提供 Kubernetes token。

如果你希望“打开 UI 即可用”（免登录），可以让 `argo-server` 以自身 ServiceAccount 权限工作：
```bash
kubectl -n argo patch deployment argo-server --type='json' -p='[
  {"op":"replace","path":"/spec/template/spec/containers/0/args","value":["server","--auth-mode=server"]}
]'
kubectl -n argo rollout status deploy/argo-server --timeout=120s
```

注意：这种方式意味着“所有访问者都以 `argo-server` 的权限操作”，安全风险很高，请勿用于生产环境。

### 通过域名访问 Argo Workflows UI（Ingress 转发）
仓库内提供了 Ingress 示例清单：`argo/argo-workflows-ingress.yaml`。

1. 生成（或准备）TLS 证书并创建 secret（本地开发可用自签）：
   ```bash
   HOST=argo-workflows.example.com
   TMP_DIR=$(mktemp -d)
   openssl req -x509 -nodes -newkey rsa:2048 -days 365 \
     -keyout "$TMP_DIR/tls.key" -out "$TMP_DIR/tls.crt" \
     -subj "/CN=$HOST" -addext "subjectAltName=DNS:$HOST"

   kubectl -n argo create secret tls argo-server-tls \
     --cert="$TMP_DIR/tls.crt" --key="$TMP_DIR/tls.key"
   rm -rf "$TMP_DIR"
   ```
2. 应用 Ingress（默认使用 `ingressClassName: nginx`，后端走 HTTPS）：
   ```bash
   kubectl apply -f argo/argo-workflows-ingress.yaml
   kubectl -n argo get ingress argo-server
   ```
3. 让域名解析到 Ingress Controller 地址：
   - Docker Desktop 场景下 Ingress LB 常见是 `localhost`，可以在本机 `/etc/hosts` 加一条：
     ```bash
     # 需要 sudo
     echo "127.0.0.1 argo-workflows.example.com" | sudo tee -a /etc/hosts
     ```
4. 浏览器访问：
   - `https://argo-workflows.example.com`
   - 若使用自签证书，浏览器会提示不受信任，确认继续即可。

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
   - Web UI 可通过 Ingress 域名访问（推荐），或 `kubectl port-forward -n argo svc/argo-server 2746:2746` 临时访问。
2. 部署示例工作流：
   ```bash
   # 给 argo 命名空间的 default ServiceAccount 补齐 RBAC（避免 wait 容器因权限不足失败）
   kubectl apply -f argo/workflow-sa-rbac.yaml

   # 注意：该示例使用 generateName，建议用 create（而不是 apply）
   kubectl create -n argo -f argo/workflow-hello.yaml
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
# generateName 会生成实际名称，按 label 清理更方便
kubectl delete wf -n argo -l example=hello-argo
```
（已配置 TTL，工作流完成后约 5 分钟会自动清理，无需手动删除；若需立即清理可执行上方命令。）

## 常见问题
- 安装未完成或 CRD 缺失：`kubectl apply` 时报 `no matches for kind Workflow`，请先安装 Argo Workflows。
- 权限问题：若使用自定义命名空间，确保对应 ServiceAccount 及 RBAC 已配置允许创建 Pod。
- 节点显示失败但 main 容器日志正常：通常是 `wait` 容器写入 `workflowtaskresults.argoproj.io` 被 RBAC 拒绝；为工作流指定 `spec.serviceAccountName` 并授予 `workflowtaskresults` 权限即可（见 `argo/workflow-sa-rbac.yaml`）。
- 镜像拉取失败：可将 `alpine:3.19` 替换为已在集群可用的镜像仓库地址。

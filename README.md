# Helloworld (Gin) — 本地与 Kubernetes 部署指南

本项目是使用 Gin 的最简 Web 服务，`GET /` 返回 `hello world`。支持本地运行、容器化以及在本地 Kubernetes（Docker Desktop / kind / minikube）部署与访问。

## 本地运行

- 安装依赖并启动：
  - `go mod tidy`
  - `go run .`
- 验证：`curl http://localhost:8080/`

## 构建与运行容器

- 构建镜像：`docker build -t helloworld:local .`
- 运行验证：`docker run --rm -p 8080:8080 helloworld:local`

## 在本地 Kubernetes 部署

1) 将镜像放入集群
- Docker Desktop Kubernetes：无需额外操作，直接使用本地镜像。

2) 应用清单并等待就绪
- `kubectl apply -f k8s/deployment.yaml -f k8s/service.yaml`
- `kubectl rollout status deployment/helloworld`

## 访问方式与原理（为何可免端口转发）

默认的 `Service` 类型是 `ClusterIP`，仅在集群内部可达。如果使用 `kubectl port-forward`，其实是在本机与 Pod 之间建立一个临时隧道；想要“直接访问”则需要让服务对主机网络可达，常见做法是 NodePort 或 Ingress：

- NodePort（L4 暴露）：在每个节点开放一个固定端口（默认范围 30000–32767），将流量转发到后端 `Service`。
  - 应用：`kubectl apply -f k8s/service.nodeport.yaml`
  - 访问：
    - Docker Desktop：`http://localhost:30080`
  - 原理：Node 的主机网络开放了一个端口，直接把外部请求转到集群内服务，无需 `port-forward`。

- Ingress（L7 路由）：需要安装 Ingress Controller（如 nginx-ingress）。Ingress 在 80/443 端口接受 HTTP(S) 请求，并按域名/路径转发到后端 `Service`。
  - 清单：`k8s/ingress.yaml`（默认 host 为 `helloworld.localdev.me`）
  - 安装 Controller：
    - Docker Desktop：使用 Helm 安装 `ingress-nginx`，其 Service 通常为 LoadBalancer，可直接通过 `localhost` 访问。
      # 安装 ingress-nginx（参考官方快速开始），再应用 k8s/ingress.yaml(kubectl apply -f k8s/ingress.yaml)
  - 访问：`curl -H "Host: helloworld.localdev.me" http://localhost/`
  - 原理：Ingress Controller 作为反向代理(独立服务)在节点对外监听 80/443，按规则转发 HTTP 流量，无需 `port-forward`。

提示：两者的选择
- 仅需暴露单个端口时，NodePort 更简单。
- 需要基于域名/路径的 HTTP 路由、TLS 证书等能力时，使用 Ingress。


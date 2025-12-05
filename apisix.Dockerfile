# 基于官方 APISIX 镜像，采用 Standalone（YAML）模式，将路由/配置直接烘焙进镜像。
# 适用场景：本地演示、PoC、或简化部署；生产建议 etcd/控制面进行动态下发。

FROM apache/apisix:3.8.0-alpine

# 将本仓库的 APISIX 配置复制到镜像中的默认配置路径。
# 注意：不要在指令行尾追加注释，否则可能触发 dockerfile 解析错误。
COPY conf/config.yaml /usr/local/apisix/conf/config.yaml
COPY conf/apisix.yaml /usr/local/apisix/conf/apisix.yaml

# 暴露默认端口：
# - 9080: 网关 HTTP 入口
# - 9443: 网关 HTTPS 入口
# - 9180: Admin API（仅开发环境暴露，生产需加密、鉴权或内网隔离）
EXPOSE 9080 9443 9180

# 启动命令由基础镜像提供（docker-entrypoint.sh），此处无需再次指定。


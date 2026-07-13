# Kubernetes 部署示例

本目录提供 **最小可读的 K8s manifest 示例**，用于展示服务容器化上云思路。  
当前项目日常联调以 **docker-compose** 为主；生产环境可按需扩展为完整 Helm / Kustomize。

## 前置条件

- 集群内已有：MySQL、Redis、RabbitMQ、Etcd、Nacos（或使用云服务）
- 已构建并推送镜像到仓库（参考 `docker-compose.prod.yml` 中的 `lqdockerhub3214/*`）
- Nacos 中已配置 `amap-lq-docker` 或等价 DataId

## 文件说明

| 文件 | 说明 |
|------|------|
| `namespace.yaml` | 命名空间 `amap` |
| `configmap-nacos.yaml` | Nacos 连接信息（非业务密钥） |
| `deployment-rpcorder.yaml` | 订单服务 Deployment 示例 |
| `deployment-apigateway.yaml` | 网关 Deployment + Service（对外暴露） |

## 部署顺序

```bash
kubectl apply -f namespace.yaml
kubectl apply -f configmap-nacos.yaml
kubectl apply -f deployment-rpcorder.yaml
kubectl apply -f deployment-apigateway.yaml
```

## 验证

```bash
kubectl -n amap get pods
kubectl -n amap get svc
```

## 扩展建议

| 组件 | 建议 |
|------|------|
| rpcUser / rpcDriver / rpcMap | 复制 rpcOrder Deployment，改镜像名与端口 |
| Ingress | 将 `apigateway` Service 接入 Ingress，配置 `/api`、`/ws` |
| HPA | 对 `rpcOrder`、`apigateway` 按 CPU / QPS 水平扩缩 |
| Config | 敏感配置放 Secret，通过 Nacos 或 External Secrets 注入 |

## 与 docker-compose 的关系

| 方式 | 适用场景 |
|------|----------|
| `docker-compose.yml` | 本地开发、单机演示 |
| `docker-compose.prod.yml` | 单机 / 小团队生产，拉取预构建镜像 |
| `deploy/k8s/` | 多副本、自动扩缩、集群运维 |

# ========== 阶段1：编译 ==========
FROM golang:1.26-alpine AS builder

WORKDIR /build
  
# go.work 要求工作区内所有模块都存在，学习阶段整包复制最省事
COPY amap/go.work amap/go.work.sum ./
COPY amap/common ./common
COPY amap/apiGateway ./apiGateway
COPY amap/rpcDriver ./rpcDriver
COPY amap/rpcMap ./rpcMap
COPY amap/rpcOrder ./rpcOrder
COPY amap/rpcUser ./rpcUser
# 编译 Linux 静态二进制
RUN cd rpcUser && CGO_ENABLED=0 GOOS=linux go build -o /out/rpcuser .
  
  # 编译 Linux 静态二进制
RUN cd rpcUser && CGO_ENABLED=0 GOOS=linux go build -o /out/rpcuser .
  
  # ========== 阶段2：运行 ==========
FROM alpine:3.20
  
  # 时区 + HTTPS 证书（访问 Nacos 等外网需要）
RUN apk add --no-cache ca-certificates tzdata \
&& cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
  
  # 保持和本机一致的目录结构（nacos 相对路径才能找到）
WORKDIR /app/rpcUser

COPY --from=builder /out/rpcuser .
COPY amap/rpcUser/etc ./etc
COPY amap/common/yaml ../common/yaml

EXPOSE 8081

CMD ["./rpcuser", "-f", "etc/rpcuser.docker.yaml"]
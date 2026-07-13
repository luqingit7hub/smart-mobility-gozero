# ========== 阶段1：编译 ==========
FROM golang:1.26-alpine AS builder

WORKDIR /build

COPY amap/go.work amap/go.work.sum ./
COPY amap/common ./common
COPY amap/apiGateway ./apiGateway
COPY amap/rpcDriver ./rpcDriver
COPY amap/rpcMap ./rpcMap
COPY amap/rpcOrder ./rpcOrder
COPY amap/rpcUser ./rpcUser

RUN cd rpcMap && CGO_ENABLED=0 GOOS=linux go build -o /out/rpcmap .

# ========== 阶段2：运行 ==========
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

WORKDIR /app/rpcMap

COPY --from=builder /out/rpcmap .
COPY amap/rpcMap/etc ./etc
COPY amap/common/yaml ../common/yaml

EXPOSE 8084

CMD ["./rpcmap", "-f", "etc/rpcmap.docker.yaml"]
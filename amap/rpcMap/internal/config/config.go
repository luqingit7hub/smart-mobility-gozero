package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf
	RpcUser   zrpc.RpcClientConf
	RpcDriver zrpc.RpcClientConf
	RpcOrder  zrpc.RpcClientConf
}

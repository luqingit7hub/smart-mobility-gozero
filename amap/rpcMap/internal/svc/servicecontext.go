package svc

import (
	"rpcDriver/rpcdriverclient"
	"rpcMap/internal/config"
	"rpcOrder/rpcorderclient"
	"rpcUser/rpcuserclient"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config    config.Config
	RpcUser   rpcuserclient.RpcUser
	RpcDriver rpcdriverclient.RpcDriver
	RpcOrder  rpcorderclient.RpcOrder
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:    c,
		RpcUser:   rpcuserclient.NewRpcUser(zrpc.MustNewClient(c.RpcUser)),
		RpcDriver: rpcdriverclient.NewRpcDriver(zrpc.MustNewClient(c.RpcDriver)),
		RpcOrder:  rpcorderclient.NewRpcOrder(zrpc.MustNewClient(c.RpcOrder)),
	}
}

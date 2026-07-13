// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"apiGateway/internal/config"
	"apiGateway/internal/middleware"
	"rpcDriver/rpcdriverclient"
	"rpcMap/rpcmapclient"
	rpcOrderClient "rpcOrder/rpcorderclient"
	"rpcUser/rpcuserclient"

	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config    config.Config
	UserAuth  rest.Middleware
	RpcUser   rpcuserclient.RpcUser
	RpcDriver rpcdriverclient.RpcDriver
	RpcOrder  rpcOrderClient.RpcOrder
	RpcMap    rpcmapclient.RpcMap
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:    c,
		UserAuth:  middleware.NewUserAuthMiddleware().Handle,
		RpcUser:   rpcuserclient.NewRpcUser(zrpc.MustNewClient(c.RpcUser)),
		RpcDriver: rpcdriverclient.NewRpcDriver(zrpc.MustNewClient(c.RpcDriver)),
		RpcOrder:  rpcOrderClient.NewRpcOrder(zrpc.MustNewClient(c.RpcOrder)),
		RpcMap:    rpcmapclient.NewRpcMap(zrpc.MustNewClient(c.RpcMap)),
	}
}

package main

// rpcOrder 进程入口：对外提供 gRPC，并在后台启动两个消费者（第6步 Stream、第7步 延迟 MQ）。
import (
	commonconfig "common/config"
	_ "common/init"
	"common/pool"
	"common/rmq"
	"context"
	"flag"
	"fmt"
	"rpcOrder/internal/config"
	"rpcOrder/internal/server"
	"rpcOrder/internal/svc"
	"rpcOrder/rpcOrder"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/rpcorder.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		rpcOrder.RegisterRpcOrderServer(grpcServer, server.NewRpcOrderServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	// 抢单 Stream 消费者：异步将 Redis 抢单事件落库 MySQL
	consumerCtx, cancelConsumer := context.WithCancel(commonconfig.Ctx)
	defer cancelConsumer()
	go func() {
		if err := pool.StartOrderGrabbedConsumer(consumerCtx); err != nil {
			fmt.Printf("[rpcOrder] stream consumer 退出: %v\n", err)
		}
	}()
	// 延迟队列消费者：6 分钟推司机、10 分钟无人接单取消
	rmq.StartConsumerWithRetry(consumerCtx, "delay", func(ctx context.Context) error {
		return rmq.StartDelayConsumer(ctx, pool.HandleOrderDelay)
	})

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}

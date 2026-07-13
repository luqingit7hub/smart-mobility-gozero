// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package main

import (
	"apiGateway/internal/config"
	"apiGateway/internal/handler"
	"apiGateway/internal/middleware"
	"apiGateway/internal/svc"
	commonconfig "common/config"
	_ "common/init"
	"common/rmq"
	"common/ws"
	"context"
	"flag"
	"fmt"
	"net/http"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/apigateway-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	// 【第9步】WebSocket 长连接 + RabbitMQ 实时通知消费（与上方 REST 路由并列，手动注册）
	hub := ws.NewHub()
	notifyCtx, cancelNotify := context.WithCancel(commonconfig.Ctx)
	defer cancelNotify()
	rmq.StartConsumerWithRetry(notifyCtx, "notify", func(ctx context.Context) error {
		return ws.StartOrderNotifyConsumer(ctx, hub)
	})
	parseToken := func(token string) (int64, error) {
		claims, err := middleware.TokenGet(token)
		if err != nil {
			return 0, err
		}
		return ws.ParseUserIDFromClaims(claims["userId"])
	}
	server.AddRoutes([]rest.Route{
		{
			Method:  http.MethodGet,
			Path:    "/ws/user",
			Handler: ws.ServeUserWS(hub, parseToken), // 乘客 WS：收接单/取消/完单
		},
		{
			Method:  http.MethodGet,
			Path:    "/ws/driver",
			Handler: ws.ServeDriverWS(hub, parseToken), // 司机 WS：收附近新单
		},
	})

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}

package logic

import (
	"common/config"
	"common/model"
	"common/pkg"
	"context"
	"errors"
	"fmt"

	"rpcOrder/internal/svc"
	"rpcOrder/rpcOrder"

	"github.com/zeromicro/go-zero/core/logx"
)

type JourneyLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewJourneyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *JourneyLogic {
	return &JourneyLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *JourneyLogic) Journey(in *rpcOrder.JourneyReq) (*rpcOrder.JourneyResp, error) {
	//用户id输入两个地址
	//两个地址解析两个经纬度
	startingPoint, err := pkg.GetCoordinates(in.StartingPoint)
	if err != nil {
		return nil, errors.New("获取起点经纬度失败")
	}
	beginLng := startingPoint.Lng
	beginLat := startingPoint.Lat
	destination, err := pkg.GetCoordinates(in.Destination)
	if err != nil {
		return nil, errors.New("获取终点经纬度失败")
	}
	destLng := destination.Lng
	destLat := destination.Lat
	//两个经纬度获取:价格,时间,时长
	pathPlan := pkg.PathPlanReq{
		OriginLng:      beginLng,
		OriginLat:      beginLat,
		DestinationLng: destLng,
		DestinationLat: destLat,
	}
	simpleRoute, err := pkg.GetPathPlan(pathPlan)
	if err != nil {
		return nil, errors.New("经纬度转换路线失败")
	}
	if len(simpleRoute) == 0 {
		return nil, errors.New("未获取到可用路线")
	}
	simpleRouteData := simpleRoute[0]
	times := simpleRouteData.Duration
	mi := simpleRouteData.Distance
	price := simpleRouteData.Toll
	status := simpleRouteData.TrafficCondition

	//判断用户余额是否可以下订单
	var userModel model.User
	if err := userModel.UserModelFindId(config.DB, in.Uid); err != nil {
		return nil, errors.New("用户异常,查找失败")
	}
	//if userModel.Balance < price {
	//	return nil, errors.New("用户余额不足,请先充值余额")
	//}

	startAdcode, startCityName, err := pkg.GetCityByLocation(beginLng, beginLat)
	if err != nil {
		return nil, errors.New("获取起点城市编码失败")
	}

	fmt.Println("用户查看路线成功, uid:", userModel.ID)

	routePoints := make([]*rpcOrder.RoutePoint, 0, len(simpleRouteData.RoutePoints))
	for _, p := range simpleRouteData.RoutePoints {
		routePoints = append(routePoints, &rpcOrder.RoutePoint{
			Lng: p.Lng,
			Lat: p.Lat,
		})
	}

	return &rpcOrder.JourneyResp{
		Price:         float32(price),
		Distance:      int64(mi),
		Duration:      int64(times),
		Status:        int64(status),
		Waittine:      5,
		StartAdcode:   startAdcode,
		StartCityName: startCityName,
		StartLng:      beginLng,
		StartLat:      beginLat,
		EndLng:        destLng,
		EndLat:        destLat,
		RoutePoints:   routePoints,
	}, nil
}

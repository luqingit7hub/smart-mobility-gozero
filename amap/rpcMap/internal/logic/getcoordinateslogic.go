package logic

import (
	"common/config"
	"common/model"
	"common/pkg"
	"context"
	"errors"
	"fmt"

	"rpcMap/internal/svc"
	"rpcMap/rpcMap"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type GetCoordinatesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCoordinatesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCoordinatesLogic {
	return &GetCoordinatesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetCoordinatesLogic) GetCoordinates(in *rpcMap.GetCoordinatesReq) (*rpcMap.GetCoordinatesResp, error) {
	if in.Address == "" {
		return nil, errors.New("地址不能为空")
	}
	if in.Uid <= 0 {
		return nil, errors.New("用户id不能为空")
	}
	if in.Type != 1 && in.Type != 2 {
		return nil, errors.New("type=1(用户),type=2(司机),不可为其他")
	}
	location, err := pkg.GetCoordinates(in.Address)
	if err != nil {
		return nil, errors.New("获取经纬度失败")
	}
	if in.Type == 1 {
		var userModel model.User
		if err := userModel.UserModelFindId(config.DB, in.Uid); err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, errors.New("用户不存在")
			}
			return nil, errors.New("用户查询失败")
		}
		userModel.CurrentLng = location.Lng
		userModel.CurrentLat = location.Lat
		if err := userModel.UserModelUpd(config.DB); err != nil {
			return nil, errors.New("用户经纬度修改失败")
		}
		fmt.Println("用户地址转经纬度成功, uid:", userModel.ID)
	} else {
		var driverModel model.Driver
		if err := driverModel.DriverModelFindId(config.DB, in.Uid); err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, errors.New("司机不存在")
			}
			return nil, errors.New("司机查询失败")
		}
		driverModel.CurrentLng = location.Lng
		driverModel.CurrentLat = location.Lat
		driverModel.OnlineStatus = 1
		if err := driverModel.DriverModelUpd(config.DB); err != nil {
			return nil, errors.New("司机经纬度修改失败")
		}
		fmt.Println("司机地址转经纬度成功, driver_id:", driverModel.ID)
	}
	return &rpcMap.GetCoordinatesResp{
		Address: in.Address,
		Lng:     location.Lng,
		Lat:     location.Lat,
	}, nil
}

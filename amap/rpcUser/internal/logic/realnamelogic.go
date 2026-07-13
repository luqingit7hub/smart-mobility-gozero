package logic

import (
	"common/config"
	"common/model"
	"common/pkg"
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"rpcUser/internal/svc"
	"rpcUser/rpcUser"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type RealNameLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRealNameLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RealNameLogic {
	return &RealNameLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RealNameLogic) RealName(in *rpcUser.RealNameReq) (*rpcUser.RealNameResp, error) {
	if in.Uid <= 0 {
		return nil, errors.New("用户id不能为空")
	}
	if in.Nickname == "" || in.Email == "" || in.Avatar == "" {
		return nil, errors.New("请上传基本信息")
	}
	if in.Gender != 1 && in.Gender != 2 {
		return nil, errors.New("请上传基本信息")
	}
	cardNo := strings.TrimSpace(in.CardNo)
	realName := strings.TrimSpace(in.RealName)
	if cardNo == "" || realName == "" {
		return nil, errors.New("身份证号和真实姓名不能为空")
	}
	if regexp.MustCompile(`^\d{17}[\dXx]$`).MatchString(cardNo) == false {
		return nil, errors.New("身份证号格式不正确")
	}
	var userModel model.User
	if err := userModel.UserModelFindId(config.DB, in.Uid); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("用户不存在")
		}
		return nil, errors.New("用户查询失败")
	}
	if userModel.Status == 1 {
		return nil, errors.New("已完成实名认证")
	}
	if userModel.Status == 3 {
		return nil, errors.New("账号已禁用")
	}
	if userModel.Status != 2 {
		return nil, errors.New("账号状态异常，无法实名认证")
	}
	if in.Gender != 0 && in.Gender != 1 && in.Gender != 2 {
		return nil, errors.New("gender=0(未知),gender=1(男),gender=2(女),不可为其他")
	}

	result, err := pkg.RealName(cardNo, realName)
	if err != nil {
		return nil, errors.New("实名认证接口调用失败")
	}
	if result.ErrorCode != 0 || result.Result.Isok == false {
		fmt.Println("实名认证失败:", result.Reason)
		return nil, errors.New("实名认证失败，姓名与身份证号不匹配")
	}
	userModel.Name = realName
	userModel.IdCard = cardNo
	userModel.Nickname = in.Nickname
	userModel.Avatar = in.Avatar
	userModel.Gender = int(in.Gender)
	userModel.Email = in.Email
	userModel.Status = 1
	if err := userModel.UserModelUpd(config.DB); err != nil {
		return nil, errors.New("实名认证信息保存失败")
	}
	fmt.Println("用户实名认证成功, uid:", userModel.ID)
	return &rpcUser.RealNameResp{Msg: "实名认证成功"}, nil
}

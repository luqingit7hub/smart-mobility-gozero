package middleware

import "apiGateway/internal/types"

func FailResponse(msg string) (*types.CommonResp, error) {
	return &types.CommonResp{
		Code: 400,
		Msg:  msg,
		Data: nil,
	}, nil
}
func SuccessResponse(data interface{}) (*types.CommonResp, error) {
	return &types.CommonResp{
		Code: 200,
		Msg:  "success",
		Data: data,
	}, nil
}

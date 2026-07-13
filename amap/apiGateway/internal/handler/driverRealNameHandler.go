// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package handler

import (
	"net/http"

	"apiGateway/internal/logic"
	"apiGateway/internal/svc"
	"apiGateway/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func DriverRealNameHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.DriverRealNameReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		l := logic.NewDriverRealNameLogic(r.Context(), svcCtx)
		resp, err := l.DriverRealName(&req, r)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

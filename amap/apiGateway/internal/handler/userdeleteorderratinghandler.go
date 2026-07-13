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

func UserDeleteOrderRatingHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UserOrderRatingReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewUserDeleteOrderRatingLogic(r.Context(), svcCtx)
		resp, err := l.UserDeleteOrderRating(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

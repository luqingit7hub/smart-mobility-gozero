// Code scaffolded by goctl. Safe to edit.

package handler

import (
	"net/http"

	"apiGateway/internal/logic"
	"apiGateway/internal/svc"
	"apiGateway/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func OrderDriverOngoingHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.OrderDriverOngoingReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewOrderDriverOngoingLogic(r.Context(), svcCtx)
		resp, err := l.OrderDriverOngoing(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package handler

import (
	"fmt"
	"net/http"

	"apiGateway/internal/logic"
	"apiGateway/internal/svc"
)

func UserBalanceNotifyHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			fmt.Println("解析失败")
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("fail"))
			return
		}
		form := make(map[string]string)
		for k, v := range r.PostForm {
			form[k] = v[0]
		}
		l := logic.NewUserBalanceNotifyLogic(r.Context(), svcCtx)
		if result := l.UserBalanceNotify(form); result != "success" {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("fail"))
			return
		}
		_, _ = w.Write([]byte("success"))
	}
}

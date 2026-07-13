// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package middleware

import (
	"apiGateway/internal/types"
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/zeromicro/go-zero/rest/httpx"
)

type UserAuthMiddleware struct {
}

func NewUserAuthMiddleware() *UserAuthMiddleware {
	return &UserAuthMiddleware{}
}

func (m *UserAuthMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO generate middleware implement function, delete after code implementation
		// Passthrough to next handler if need
		token := r.Header.Get("token")
		if token == "" {
			rest := &types.CommonResp{
				Code: 1,
				Msg:  "头部请求的token不能为空",
				Data: nil,
			}
			httpx.OkJsonCtx(r.Context(), w, rest)
			return
		}
		if getToken, err := TokenGet(token); err != nil {
			rest := &types.CommonResp{
				Code: 1,
				Msg:  "token异常" + err.Error(),
				Data: nil,
			}
			httpx.OkJsonCtx(r.Context(), w, rest)
			return
		} else {
			userId, _ := strconv.Atoi(getToken["userId"].(string))
			fmt.Println("token获取的uid", token, userId)
			ctx := context.WithValue(r.Context(), "userId", userId)
			next(w, r.WithContext(ctx))
		}
	}
}

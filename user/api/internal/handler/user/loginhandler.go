package user

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/hd2yao/ecshop/common/app"
	"github.com/hd2yao/ecshop/common/errcode"
	"github.com/hd2yao/ecshop/user/api/internal/logic/user"
	"github.com/hd2yao/ecshop/user/api/internal/svc"
	"github.com/hd2yao/ecshop/user/api/internal/types"
)

// LoginHandler 用户登录
func LoginHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.LoginRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := user.NewLoginLogic(r.Context(), svcCtx)
		resp, err := l.Login(&req)

		response := app.NewResponse(r.Context(), w)
		if err != nil {
			response.Error(errcode.CommonServerError.WithCause(err))
		} else {
			response.Success(resp)
		}
	}
}

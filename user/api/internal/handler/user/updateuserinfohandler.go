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

// UpdateUserInfoHandler 修改用户个人信息
func UpdateUserInfoHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UpdateUserInfoRequest
		if err := httpx.Parse(r, &req); err != nil {
			app.NewResponse(r.Context(), w).Error(errcode.CommonParamError.WithCause(err))
			return
		}

		l := user.NewUpdateUserInfoLogic(r.Context(), svcCtx)
		resp, err := l.UpdateUserInfo(&req)

		response := app.NewResponse(r.Context(), w)
		if err != nil {
			response.Error(errcode.CommonServerError.WithCause(err))
		} else {
			response.Success(resp)
		}
	}
}

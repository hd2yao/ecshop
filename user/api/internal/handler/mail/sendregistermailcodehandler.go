package mail

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/hd2yao/ecshop/user/api/internal/logic/mail"
	"github.com/hd2yao/ecshop/user/api/internal/svc"
	"github.com/hd2yao/ecshop/user/api/internal/types"
)

// SendRegisterMailCodeHandler 注册时发送邮件验证码（需要先验证图形验证码）
func SendRegisterMailCodeHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SendRegisterMailCodeRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := mail.NewSendRegisterMailCodeLogic(r.Context(), svcCtx)
		resp, err := l.SendRegisterMailCode(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

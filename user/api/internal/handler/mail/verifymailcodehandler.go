package mail

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"github.com/hd2yao/ecshop/user/api/internal/logic/mail"
	"github.com/hd2yao/ecshop/user/api/internal/svc"
	"github.com/hd2yao/ecshop/user/api/internal/types"
)

func VerifyMailCodeHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.VerifyMailCodeRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := mail.NewVerifyMailCodeLogic(r.Context(), svcCtx)
		resp, err := l.VerifyMailCode(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
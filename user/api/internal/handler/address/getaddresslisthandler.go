package address

import (
	"net/http"

	"github.com/hd2yao/ecshop/user/api/internal/logic/address"
	"github.com/hd2yao/ecshop/user/api/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 获取用户全部地址列表
func GetAddressListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := address.NewGetAddressListLogic(r.Context(), svcCtx)
		resp, err := l.GetAddressList()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

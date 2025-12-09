package food

import (
	"net/http"

	"github.com/hd2yao/ecshop/food/api/internal/logic/food"
	"github.com/hd2yao/ecshop/food/api/internal/svc"
	"github.com/hd2yao/ecshop/food/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 查询我的美食列表
func GetMyFoodListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetMyFoodListRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := food.NewGetMyFoodListLogic(r.Context(), svcCtx)
		resp, err := l.GetMyFoodList(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

package food

import (
	"net/http"

	"github.com/hd2yao/ecshop/food/api/internal/logic/food"
	"github.com/hd2yao/ecshop/food/api/internal/svc"
	"github.com/hd2yao/ecshop/food/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 新增/修改美食信息（food_id为0或未提供时为新增，否则为修改）
func CreateOrUpdateFoodHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CreateOrUpdateFoodRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := food.NewCreateOrUpdateFoodLogic(r.Context(), svcCtx)
		resp, err := l.CreateOrUpdateFood(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

package feed

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/hd2yao/ecshop/common/app"
	"github.com/hd2yao/ecshop/common/errcode"
	"github.com/hd2yao/ecshop/home/api/internal/logic/feed"
	"github.com/hd2yao/ecshop/home/api/internal/svc"
	"github.com/hd2yao/ecshop/home/api/internal/types"
)

// GetHomeFeedHandler 获取首页 feed 食谱列表
func GetHomeFeedHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.HomeFeedRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := feed.NewGetHomeFeedLogic(r.Context(), svcCtx)
		resp, err := l.GetHomeFeed(&req)

		response := app.NewResponse(r.Context(), w)
		if err != nil {
			response.Error(errcode.CommonServerError.WithCause(err))
		} else {
			response.Success(resp)
		}
	}
}

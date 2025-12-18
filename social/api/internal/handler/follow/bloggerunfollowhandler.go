package follow

import (
	"net/http"

	"github.com/hd2yao/ecshop/social/api/internal/logic/follow"
	"github.com/hd2yao/ecshop/social/api/internal/svc"
	"github.com/hd2yao/ecshop/social/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 博主取消关注
func BloggerUnfollowHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.FollowRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := follow.NewBloggerUnfollowLogic(r.Context(), svcCtx)
		resp, err := l.BloggerUnfollow(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

package feed

import (
	"context"

	"github.com/hd2yao/ecshop/home/api/internal/svc"
	"github.com/hd2yao/ecshop/home/api/internal/types"
	"github.com/hd2yao/ecshop/home/rpc/types/home"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetHomeFeedLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取首页 feed 食谱列表
func NewGetHomeFeedLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetHomeFeedLogic {
	return &GetHomeFeedLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetHomeFeedLogic) GetHomeFeed(req *types.HomeFeedRequest) (resp *types.HomeFeedResponse, err error) {
	// 1. 组装 RPC 请求
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}

	rpcReq := &home.HomeFeedReq{
		UserId:      req.UserId,
		FeedVersion: req.FeedVersion,
		PullRefresh: req.PullRefresh,
		Page:        int32(req.Page),
		PageSize:    int32(pageSize),
	}

	// 2. 调用 RPC
	rpcResp, err := l.svcCtx.HomeRpc.GetHomeFeed(l.ctx, rpcReq)
	if err != nil {
		return nil, err
	}

	// 3. 转成 HTTP 层响应
	resp = &types.HomeFeedResponse{
		Code:        int(rpcResp.Code),
		Message:     rpcResp.Message,
		FeedVersion: rpcResp.FeedVersion,
		CurrPage:    int64(rpcResp.CurrPage),
		HasNextPage: rpcResp.HasNextPage,
	}

	if len(rpcResp.RecipesList) > 0 {
		resp.RecipesList = make([]types.RecipesInfo, 0, len(rpcResp.RecipesList))
		for _, r := range rpcResp.RecipesList {
			resp.RecipesList = append(resp.RecipesList, types.RecipesInfo{
				Id:                 r.Id,
				RecipesName:        r.RecipesName,
				RecipesUrl:         r.RecipesUrl,
				RecipesUserName:    r.RecipesUserName,
				RecipesUserAvatar:  r.RecipesUserAvatar,
				RecipesDescription: r.RecipesDescription,
			})
		}
	}

	return resp, nil
}

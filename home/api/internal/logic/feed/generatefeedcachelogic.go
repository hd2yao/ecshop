package feed

import (
	"context"

	"github.com/hd2yao/ecshop/home/api/internal/svc"
	"github.com/hd2yao/ecshop/home/api/internal/types"
	"github.com/hd2yao/ecshop/home/rpc/types/home"

	"github.com/zeromicro/go-zero/core/logx"
)

type GenerateFeedCacheLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGenerateFeedCacheLogic 生成首页 feed 食谱缓存（写入 Redis）
func NewGenerateFeedCacheLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GenerateFeedCacheLogic {
	return &GenerateFeedCacheLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GenerateFeedCacheLogic) GenerateFeedCache(req *types.GenerateFeedCacheRequest) (resp *types.GenerateFeedCacheResponse, err error) {
	// 调用 home.rpc 的 GenerateFeedCache
	rpcResp, err := l.svcCtx.HomeRpc.GenerateFeedCache(l.ctx, &home.GenerateFeedCacheReq{
		Operator: req.Operator,
		Force:    req.Force,
	})
	if err != nil {
		return nil, err
	}

	// 转成 HTTP 层的响应结构
	resp = &types.GenerateFeedCacheResponse{
		Code:        int(rpcResp.Code),
		Message:     rpcResp.Message,
		FeedVersion: rpcResp.FeedVersion,
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

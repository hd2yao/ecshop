package feed

import (
	"context"

	"github.com/hd2yao/ecshop/home/api/internal/svc"
	"github.com/hd2yao/ecshop/home/api/internal/types"

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
	// todo: add your logic here and delete this line

	return
}

package logic

import (
	"context"

	"github.com/hd2yao/ecshop/home/rpc/internal/svc"
	"github.com/hd2yao/ecshop/home/rpc/types/home"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetHomeFeedLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetHomeFeedLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetHomeFeedLogic {
	return &GetHomeFeedLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取首页 feed 食谱列表
func (l *GetHomeFeedLogic) GetHomeFeed(in *home.HomeFeedReq) (*home.HomeFeedResp, error) {
	// todo: add your logic here and delete this line

	return &home.HomeFeedResp{}, nil
}

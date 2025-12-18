package follow

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/hd2yao/ecshop/common/errcode"
	commonMiddleware "github.com/hd2yao/ecshop/common/middleware"
	"github.com/hd2yao/ecshop/social/api/internal/svc"
	"github.com/hd2yao/ecshop/social/api/internal/types"
	"github.com/hd2yao/ecshop/social/rpc/types/social"
)

type FollowerFollowLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewFollowerFollowLogic 粉丝关注（回关粉丝）
func NewFollowerFollowLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FollowerFollowLogic {
	return &FollowerFollowLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *FollowerFollowLogic) FollowerFollow(req *types.FollowRequest) (resp *types.Response, err error) {
	// 1. 从 context 中获取当前登录用户 ID（操作者）
	operatorId, ok := commonMiddleware.GetUserIDFromContext(l.ctx)
	if !ok || operatorId <= 0 {
		return &types.Response{
			Code:    errcode.UserTokenInvalid.Code(),
			Message: errcode.UserTokenInvalid.Msg(),
		}, nil
	}

	// 2. 组装 RPC 请求
	rpcReq := &social.FollowReq{
		OperatorId: operatorId,
		UserId:     req.UserId,
	}

	// 3. 调用 RPC
	rpcResp, err := l.svcCtx.SocialRpc.FollowerFollow(l.ctx, rpcReq)
	if err != nil {
		l.Errorf("调用 FollowerFollow RPC 失败: %v", err)
		return &types.Response{
			Code:    errcode.CommonServerError.Code(),
			Message: errcode.CommonServerError.Msg(),
		}, nil
	}

	// 4. 返回结果
	return &types.Response{
		Code:    int(rpcResp.Code),
		Message: rpcResp.Message,
	}, nil
}

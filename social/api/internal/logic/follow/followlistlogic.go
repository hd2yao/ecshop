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

type FollowListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewFollowListLogic 关注列表
func NewFollowListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FollowListLogic {
	return &FollowListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *FollowListLogic) FollowList(req *types.FollowListRequest) (resp *types.FollowListResponse, err error) {
	// 1. 从 context 中获取当前登录用户 ID（操作者）
	operatorId, ok := commonMiddleware.GetUserIDFromContext(l.ctx)
	if !ok || operatorId <= 0 {
		return &types.FollowListResponse{
			Code:    errcode.UserTokenInvalid.Code(),
			Message: errcode.UserTokenInvalid.Msg(),
		}, nil
	}

	// 2. 组装 RPC 请求
	rpcReq := &social.FollowListReq{
		OperatorId: operatorId,
		Page:       req.Page,
		Size:       req.Size,
	}

	// 3. 调用 RPC
	rpcResp, err := l.svcCtx.SocialRpc.FollowList(l.ctx, rpcReq)
	if err != nil {
		l.Errorf("调用 FollowList RPC 失败: %v", err)
		return &types.FollowListResponse{
			Code:    errcode.CommonServerError.Code(),
			Message: errcode.CommonServerError.Msg(),
		}, nil
	}

	// 4. 转换为 HTTP 响应结构
	resp = &types.FollowListResponse{
		Code:    int(rpcResp.Code),
		Message: rpcResp.Message,
		Total:   rpcResp.Total,
		Page:    rpcResp.Page,
		Size:    rpcResp.Size,
	}

	if len(rpcResp.List) > 0 {
		resp.List = make([]types.UserBrief, 0, len(rpcResp.List))
		for _, u := range rpcResp.List {
			resp.List = append(resp.List, types.UserBrief{
				UserId:   u.UserId,
				Name:     u.Name,
				Avatar:   u.Avatar,
				IsFollow: u.IsFollow,
				IsMutual: u.IsMutual,
			})
		}
	}

	return resp, nil

	return
}

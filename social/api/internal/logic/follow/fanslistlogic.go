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

type FansListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewFansListLogic 粉丝列表
func NewFansListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FansListLogic {
	return &FansListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *FansListLogic) FansList(req *types.FansListRequest) (resp *types.FansListResponse, err error) {
	// 1. 从 context 中获取当前登录用户 ID（操作者）
	operatorId, ok := commonMiddleware.GetUserIDFromContext(l.ctx)
	if !ok || operatorId <= 0 {
		return &types.FansListResponse{
			Code:    errcode.UserTokenInvalid.Code(),
			Message: errcode.UserTokenInvalid.Msg(),
		}, nil
	}

	// 2. 组装 RPC 请求
	userId := req.UserId
	rpcReq := &social.FansListReq{
		OperatorId: operatorId,
		UserId:     userId,
		Page:       req.Page,
		Size:       req.Size,
	}

	// 3. 调用 RPC
	rpcResp, err := l.svcCtx.SocialRpc.FansList(l.ctx, rpcReq)
	if err != nil {
		l.Errorf("调用 FansList RPC 失败: %v", err)
		return &types.FansListResponse{
			Code:    errcode.CommonServerError.Code(),
			Message: errcode.CommonServerError.Msg(),
		}, nil
	}

	// 4. 转换为 HTTP 响应结构
	resp = &types.FansListResponse{
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
}

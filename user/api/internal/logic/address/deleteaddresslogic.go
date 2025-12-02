package address

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/hd2yao/ecshop/common/errcode"
	"github.com/hd2yao/ecshop/common/middleware"
	"github.com/hd2yao/ecshop/user/api/internal/svc"
	"github.com/hd2yao/ecshop/user/api/internal/types"
	"github.com/hd2yao/ecshop/user/rpc/types/user"
)

type DeleteAddressLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewDeleteAddressLogic 删除用户地址
func NewDeleteAddressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteAddressLogic {
	return &DeleteAddressLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteAddressLogic) DeleteAddress(req *types.DeleteAddressRequest) (resp *types.DeleteAddressResponse, err error) {
	// 1. 从 context 中获取用户 ID
	userID := middleware.MustGetUserIDFromContext(l.ctx)

	l.Infof("删除地址，用户ID: %d, 地址ID: %d", userID, req.AddressId)

	// 2. 调用 RPC 服务删除地址
	rpcResp, err := l.svcCtx.UserRpc.DeleteAddress(l.ctx, &user.DeleteAddressReq{
		UserId:    userID,
		AddressId: req.AddressId,
	})

	if err != nil {
		l.Errorf("调用 RPC 删除地址失败: %v", err)
		return &types.DeleteAddressResponse{
			Code:    errcode.CommonServerError.Code(),
			Message: errcode.CommonServerError.Msg(),
		}, nil
	}

	// 3. 返回响应
	return &types.DeleteAddressResponse{
		Code:    int(rpcResp.Code),
		Message: rpcResp.Message,
	}, nil
}

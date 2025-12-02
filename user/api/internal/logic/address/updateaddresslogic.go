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

type UpdateAddressLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewUpdateAddressLogic 修改用户地址
func NewUpdateAddressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateAddressLogic {
	return &UpdateAddressLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateAddressLogic) UpdateAddress(req *types.UpdateAddressRequest) (resp *types.UpdateAddressResponse, err error) {
	// 1. 从 context 中获取用户 ID
	userID := middleware.MustGetUserIDFromContext(l.ctx)

	l.Infof("修改地址，用户ID: %d, 地址ID: %d, 是否默认: %d", userID, req.AddressId, req.DefaultStatus)

	// 2. 调用 RPC 服务修改地址
	rpcResp, err := l.svcCtx.UserRpc.UpdateAddress(l.ctx, &user.UpdateAddressReq{
		UserId:        userID,
		AddressId:     req.AddressId,
		DefaultStatus: int32(req.DefaultStatus),
		ReceiveName:   req.ReceiveName,
		Phone:         req.Phone,
		Province:      req.Province,
		City:          req.City,
		Region:        req.Region,
		DetailAddress: req.DetailAddress,
	})

	if err != nil {
		l.Errorf("调用 RPC 修改地址失败: %v", err)
		return &types.UpdateAddressResponse{
			Code:    errcode.CommonServerError.Code(),
			Message: errcode.CommonServerError.Msg(),
		}, nil
	}

	// 3. 返回响应
	return &types.UpdateAddressResponse{
		Code:    int(rpcResp.Code),
		Message: rpcResp.Message,
	}, nil
}

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

type AddAddressLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewAddAddressLogic 新增用户地址
func NewAddAddressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddAddressLogic {
	return &AddAddressLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AddAddressLogic) AddAddress(req *types.AddAddressRequest) (resp *types.AddAddressResponse, err error) {
	// 1. 从 context 中获取用户 ID
	userID := middleware.MustGetUserIDFromContext(l.ctx)

	l.Infof("新增地址，用户ID: %d, 是否默认: %d", userID, req.DefaultStatus)

	// 2. 调用 RPC 服务新增地址
	rpcResp, err := l.svcCtx.UserRpc.AddAddress(l.ctx, &user.AddAddressReq{
		UserId:        userID,
		DefaultStatus: int32(req.DefaultStatus),
		ReceiveName:   req.ReceiveName,
		Phone:         req.Phone,
		Province:      req.Province,
		City:          req.City,
		Region:        req.Region,
		DetailAddress: req.DetailAddress,
	})

	if err != nil {
		l.Errorf("调用 RPC 新增地址失败: %v", err)
		return &types.AddAddressResponse{
			Code:    errcode.CommonServerError.Code(),
			Message: errcode.CommonServerError.Msg(),
		}, nil
	}

	// 3. 返回响应
	return &types.AddAddressResponse{
		Code:      int(rpcResp.Code),
		Message:   rpcResp.Message,
		AddressId: rpcResp.AddressId,
	}, nil
}

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

type GetAddressDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGetAddressDetailLogic 获取地址详情
func NewGetAddressDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAddressDetailLogic {
	return &GetAddressDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAddressDetailLogic) GetAddressDetail(req *types.GetAddressDetailRequest) (resp *types.GetAddressDetailResponse, err error) {
	// 1. 从 context 中获取用户 ID
	userID := middleware.MustGetUserIDFromContext(l.ctx)

	l.Infof("获取地址详情，用户ID: %d, 地址ID: %d", userID, req.AddressId)

	// 2. 调用 RPC 服务获取地址详情
	rpcResp, err := l.svcCtx.UserRpc.GetAddressDetail(l.ctx, &user.GetAddressDetailReq{
		UserId:    userID,
		AddressId: req.AddressId,
	})

	if err != nil {
		l.Errorf("调用 RPC 获取地址详情失败: %v", err)
		return &types.GetAddressDetailResponse{
			Code:    errcode.CommonServerError.Code(),
			Message: errcode.CommonServerError.Msg(),
		}, nil
	}

	// 3. 转换响应
	resp = &types.GetAddressDetailResponse{
		Code:    int(rpcResp.Code),
		Message: rpcResp.Message,
	}

	// 4. 如果成功，填充地址信息
	if rpcResp.Code == int32(errcode.Success.Code()) && rpcResp.Address != nil {
		resp.Address = types.UserAddress{
			Id:            rpcResp.Address.Id,
			UserId:        rpcResp.Address.UserId,
			DefaultStatus: int(rpcResp.Address.DefaultStatus),
			ReceiveName:   rpcResp.Address.ReceiveName,
			Phone:         rpcResp.Address.Phone,
			Province:      rpcResp.Address.Province,
			City:          rpcResp.Address.City,
			Region:        rpcResp.Address.Region,
			DetailAddress: rpcResp.Address.DetailAddress,
			CreateTime:    rpcResp.Address.CreateTime,
			UpdateTime:    rpcResp.Address.UpdateTime,
		}
	}

	return resp, nil
}

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

type GetAddressListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGetAddressListLogic 获取用户全部地址列表
func NewGetAddressListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAddressListLogic {
	return &GetAddressListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAddressListLogic) GetAddressList() (resp *types.GetAddressListResponse, err error) {
	// 1. 从 context 中获取用户 ID（JWT 中间件已验证并存入）
	userID := middleware.MustGetUserIDFromContext(l.ctx)

	l.Infof("获取地址列表，用户ID: %d", userID)

	// 2. 调用 RPC 服务获取地址列表
	rpcResp, err := l.svcCtx.UserRpc.GetAddressList(l.ctx, &user.GetAddressListReq{
		UserId: userID,
	})

	if err != nil {
		l.Errorf("调用 RPC 获取地址列表失败: %v", err)
		return &types.GetAddressListResponse{
			Code:    errcode.CommonServerError.Code(),
			Message: errcode.CommonServerError.Msg(),
		}, nil
	}

	// 3. 转换响应
	resp = &types.GetAddressListResponse{
		Code:    int(rpcResp.Code),
		Message: rpcResp.Message,
	}

	// 4. 如果成功，填充地址列表
	if rpcResp.Code == int32(errcode.Success.Code()) && len(rpcResp.AddressList) > 0 {
		for _, addr := range rpcResp.AddressList {
			resp.AddressList = append(resp.AddressList, types.UserAddress{
				Id:            addr.Id,
				UserId:        addr.UserId,
				DefaultStatus: int(addr.DefaultStatus),
				ReceiveName:   addr.ReceiveName,
				Phone:         addr.Phone,
				Province:      addr.Province,
				City:          addr.City,
				Region:        addr.Region,
				DetailAddress: addr.DetailAddress,
				CreateTime:    addr.CreateTime,
				UpdateTime:    addr.UpdateTime,
			})
		}
	}

	return resp, nil
}

package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/hd2yao/ecshop/common/errcode"
	"github.com/hd2yao/ecshop/user/rpc/internal/svc"
	"github.com/hd2yao/ecshop/user/rpc/types/user"
)

type GetAddressListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetAddressListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAddressListLogic {
	return &GetAddressListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetAddressList 获取用户地址列表
func (l *GetAddressListLogic) GetAddressList(in *user.GetAddressListReq) (*user.GetAddressListResp, error) {
	// 1. 参数验证
	if in.UserId <= 0 {
		return &user.GetAddressListResp{
			Code:    int32(errcode.CommonParamError.Code()),
			Message: "用户 ID 无效",
		}, nil
	}

	l.Infof("查询用户地址列表，用户 ID: %d", in.UserId)

	// 2. 查询用户所有地址（已按默认地址和创建时间排序）
	addresses, err := l.svcCtx.UserAddressModel.FindByUserId(l.ctx, in.UserId)
	if err != nil {
		l.Errorf("查询用户地址列表失败: %v", err)
		return &user.GetAddressListResp{
			Code:    int32(errcode.CommonServerError.Code()),
			Message: errcode.CommonServerError.Msg(),
		}, nil
	}

	// 3. 转换为响应格式
	var addressList []*user.AddressInfo
	for _, addr := range addresses {
		addressList = append(addressList, &user.AddressInfo{
			Id:            int64(addr.Id),
			UserId:        addr.UserId.Int64,
			DefaultStatus: int32(addr.DefaultStatus.Int64),
			ReceiveName:   addr.ReceiveName.String,
			Phone:         addr.Phone.String,
			Province:      addr.Province.String,
			City:          addr.City.String,
			Region:        addr.Region.String,
			DetailAddress: addr.DetailAddress.String,
			CreateTime:    addr.CreateTime.Time.Format("2006-01-02 15:04:05"),
			UpdateTime:    addr.UpdateTime.Time.Format("2006-01-02 15:04:05"),
		})
	}

	return &user.GetAddressListResp{
		Code:        int32(errcode.Success.Code()),
		Message:     errcode.Success.Msg(),
		AddressList: addressList,
	}, nil
}

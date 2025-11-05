package logic

import (
	"context"
	"errors"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/hd2yao/ecshop/common/errcode"
	"github.com/hd2yao/ecshop/user/model"
	"github.com/hd2yao/ecshop/user/rpc/internal/svc"
	"github.com/hd2yao/ecshop/user/rpc/types/user"
)

type GetAddressDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetAddressDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAddressDetailLogic {
	return &GetAddressDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetAddressDetail 获取地址详情
func (l *GetAddressDetailLogic) GetAddressDetail(in *user.GetAddressDetailReq) (*user.GetAddressDetailResp, error) {
	// 1. 参数验证
	if in.UserId <= 0 || in.AddressId <= 0 {
		return &user.GetAddressDetailResp{
			Code:    int32(errcode.CommonParamError.Code()),
			Message: "参数错误",
		}, nil
	}

	l.Infof("查询地址详情，用户ID: %d, 地址ID: %d", in.UserId, in.AddressId)

	// 2. 查询地址（同时验证权限）
	address, err := l.svcCtx.UserAddressModel.FindOneByUserIdAndAddressId(l.ctx, in.UserId, uint64(in.AddressId))
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			l.Infof("地址不存在或无权限访问，用户ID: %d, 地址ID: %d", in.UserId, in.AddressId)
			return &user.GetAddressDetailResp{
				Code:    int32(errcode.CommonParamError.Code()),
				Message: "地址不存在或无权限访问",
			}, nil
		}
		l.Errorf("查询地址详情失败: %v", err)
		return &user.GetAddressDetailResp{
			Code:    int32(errcode.CommonServerError.Code()),
			Message: errcode.CommonServerError.Msg(),
		}, nil
	}

	// 3. 返回地址信息
	return &user.GetAddressDetailResp{
		Code:    int32(errcode.Success.Code()),
		Message: errcode.Success.Msg(),
		Address: &user.AddressInfo{
			Id:            int64(address.Id),
			UserId:        address.UserId.Int64,
			DefaultStatus: int32(address.DefaultStatus.Int64),
			ReceiveName:   address.ReceiveName.String,
			Phone:         address.Phone.String,
			Province:      address.Province.String,
			City:          address.City.String,
			Region:        address.Region.String,
			DetailAddress: address.DetailAddress.String,
			CreateTime:    address.CreateTime.Time.Format("2006-01-02 15:04:05"),
			UpdateTime:    address.UpdateTime.Time.Format("2006-01-02 15:04:05"),
		},
	}, nil
}

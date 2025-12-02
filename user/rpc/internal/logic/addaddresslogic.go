package logic

import (
	"context"
	"database/sql"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/hd2yao/ecshop/common/errcode"
	"github.com/hd2yao/ecshop/user/model"
	"github.com/hd2yao/ecshop/user/rpc/internal/svc"
	"github.com/hd2yao/ecshop/user/rpc/types/user"
)

type AddAddressLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAddAddressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddAddressLogic {
	return &AddAddressLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// AddAddress 新增地址
func (l *AddAddressLogic) AddAddress(in *user.AddAddressReq) (*user.AddAddressResp, error) {
	// 1. 参数验证
	if in.UserId <= 0 {
		return &user.AddAddressResp{
			Code:    int32(errcode.CommonParamError.Code()),
			Message: "用户ID无效",
		}, nil
	}

	if in.ReceiveName == "" || in.Phone == "" || in.Province == "" || in.City == "" || in.Region == "" || in.DetailAddress == "" {
		return &user.AddAddressResp{
			Code:    int32(errcode.CommonParamError.Code()),
			Message: "地址信息不完整",
		}, nil
	}

	l.Infof("新增地址，用户ID: %d, 是否默认: %d", in.UserId, in.DefaultStatus)

	// 2. 构建地址对象
	address := &model.UserAddress{
		UserId:        sql.NullInt64{Int64: in.UserId, Valid: true},
		DefaultStatus: sql.NullInt64{Int64: int64(in.DefaultStatus), Valid: true},
		ReceiveName:   sql.NullString{String: in.ReceiveName, Valid: true},
		Phone:         sql.NullString{String: in.Phone, Valid: true},
		Province:      sql.NullString{String: in.Province, Valid: true},
		City:          sql.NullString{String: in.City, Valid: true},
		Region:        sql.NullString{String: in.Region, Valid: true},
		DetailAddress: sql.NullString{String: in.DetailAddress, Valid: true},
	}

	var addressId int64

	// 3. 如果设置为默认地址，需要先清除其他默认地址（使用事务）
	if in.DefaultStatus == 1 {
		err := l.svcCtx.UserAddressModel.Trans(l.ctx, func(ctx context.Context, session sqlx.Session) error {
			// 3.1 清除该用户的所有默认地址
			if err := l.svcCtx.UserAddressModel.ClearDefaultStatus(ctx, in.UserId); err != nil {
				l.Errorf("清除默认地址状态失败: %v", err)
				return err
			}

			// 3.2 插入新地址
			result, err := l.svcCtx.UserAddressModel.Insert(ctx, address)
			if err != nil {
				l.Errorf("插入地址失败: %v", err)
				return err
			}

			// 3.3 获取插入的地址ID
			id, err := result.LastInsertId()
			if err != nil {
				l.Errorf("获取插入ID失败: %v", err)
				return err
			}
			addressId = id

			return nil
		})

		if err != nil {
			return &user.AddAddressResp{
				Code:    int32(errcode.CommonServerError.Code()),
				Message: errcode.CommonServerError.Msg(),
			}, nil
		}
	} else {
		// 4. 非默认地址，直接插入
		result, err := l.svcCtx.UserAddressModel.Insert(l.ctx, address)
		if err != nil {
			l.Errorf("插入地址失败: %v", err)
			return &user.AddAddressResp{
				Code:    int32(errcode.CommonServerError.Code()),
				Message: errcode.CommonServerError.Msg(),
			}, nil
		}

		addressId, _ = result.LastInsertId()
	}

	l.Infof("地址新增成功，地址ID: %d", addressId)

	return &user.AddAddressResp{
		Code:      int32(errcode.Success.Code()),
		Message:   errcode.Success.Msg(),
		AddressId: addressId,
	}, nil
}

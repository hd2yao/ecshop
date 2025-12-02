package logic

import (
	"context"
	"database/sql"
	"errors"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/hd2yao/ecshop/common/errcode"
	"github.com/hd2yao/ecshop/user/model"
	"github.com/hd2yao/ecshop/user/rpc/internal/svc"
	"github.com/hd2yao/ecshop/user/rpc/types/user"
)

type UpdateAddressLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateAddressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateAddressLogic {
	return &UpdateAddressLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UpdateAddress 修改地址
func (l *UpdateAddressLogic) UpdateAddress(in *user.UpdateAddressReq) (*user.UpdateAddressResp, error) {
	// 1. 参数验证
	if in.UserId <= 0 || in.AddressId <= 0 {
		return &user.UpdateAddressResp{
			Code:    int32(errcode.CommonParamError.Code()),
			Message: "参数错误",
		}, nil
	}

	if in.ReceiveName == "" || in.Phone == "" || in.Province == "" || in.City == "" || in.Region == "" || in.DetailAddress == "" {
		return &user.UpdateAddressResp{
			Code:    int32(errcode.CommonParamError.Code()),
			Message: "地址信息不完整",
		}, nil
	}

	l.Infof("修改地址，用户ID: %d, 地址ID: %d, 是否默认: %d", in.UserId, in.AddressId, in.DefaultStatus)

	// 2. 验证地址是否存在且属于当前用户
	existAddr, err := l.svcCtx.UserAddressModel.FindOneByUserIdAndAddressId(l.ctx, in.UserId, uint64(in.AddressId))
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			l.Infof("地址不存在或无权限访问，用户ID: %d, 地址ID: %d", in.UserId, in.AddressId)
			return &user.UpdateAddressResp{
				Code:    int32(errcode.CommonParamError.Code()),
				Message: "地址不存在或无权限访问",
			}, nil
		}
		l.Errorf("查询地址失败: %v", err)
		return &user.UpdateAddressResp{
			Code:    int32(errcode.CommonServerError.Code()),
			Message: errcode.CommonServerError.Msg(),
		}, nil
	}

	// 3. 更新地址对象
	existAddr.DefaultStatus = sql.NullInt64{Int64: int64(in.DefaultStatus), Valid: true}
	existAddr.ReceiveName = sql.NullString{String: in.ReceiveName, Valid: true}
	existAddr.Phone = sql.NullString{String: in.Phone, Valid: true}
	existAddr.Province = sql.NullString{String: in.Province, Valid: true}
	existAddr.City = sql.NullString{String: in.City, Valid: true}
	existAddr.Region = sql.NullString{String: in.Region, Valid: true}
	existAddr.DetailAddress = sql.NullString{String: in.DetailAddress, Valid: true}

	// 4. 如果设置为默认地址，需要先清除其他默认地址（使用事务）
	if in.DefaultStatus == 1 {
		err := l.svcCtx.UserAddressModel.Trans(l.ctx, func(ctx context.Context, session sqlx.Session) error {
			// 4.1 清除该用户的所有默认地址
			if err := l.svcCtx.UserAddressModel.ClearDefaultStatus(ctx, in.UserId); err != nil {
				l.Errorf("清除默认地址状态失败: %v", err)
				return err
			}

			// 4.2 更新地址
			if err := l.svcCtx.UserAddressModel.Update(ctx, existAddr); err != nil {
				l.Errorf("更新地址失败: %v", err)
				return err
			}

			return nil
		})

		if err != nil {
			return &user.UpdateAddressResp{
				Code:    int32(errcode.CommonServerError.Code()),
				Message: errcode.CommonServerError.Msg(),
			}, nil
		}
	} else {
		// 5. 非默认地址，直接更新
		if err := l.svcCtx.UserAddressModel.Update(l.ctx, existAddr); err != nil {
			l.Errorf("更新地址失败: %v", err)
			return &user.UpdateAddressResp{
				Code:    int32(errcode.CommonServerError.Code()),
				Message: errcode.CommonServerError.Msg(),
			}, nil
		}
	}

	l.Infof("地址修改成功，地址ID: %d", in.AddressId)

	return &user.UpdateAddressResp{
		Code:    int32(errcode.Success.Code()),
		Message: errcode.Success.Msg(),
	}, nil
}

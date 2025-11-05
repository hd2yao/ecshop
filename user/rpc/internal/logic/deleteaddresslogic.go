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

type DeleteAddressLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteAddressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteAddressLogic {
	return &DeleteAddressLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// DeleteAddress 删除地址
func (l *DeleteAddressLogic) DeleteAddress(in *user.DeleteAddressReq) (*user.DeleteAddressResp, error) {
	// 1. 参数验证
	if in.UserId <= 0 || in.AddressId <= 0 {
		return &user.DeleteAddressResp{
			Code:    int32(errcode.CommonParamError.Code()),
			Message: "参数错误",
		}, nil
	}

	l.Infof("删除地址，用户ID: %d, 地址ID: %d", in.UserId, in.AddressId)

	// 2. 验证地址是否存在且属于当前用户
	_, err := l.svcCtx.UserAddressModel.FindOneByUserIdAndAddressId(l.ctx, in.UserId, uint64(in.AddressId))
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			l.Infof("地址不存在或无权限访问，用户ID: %d, 地址ID: %d", in.UserId, in.AddressId)
			return &user.DeleteAddressResp{
				Code:    int32(errcode.CommonParamError.Code()),
				Message: "地址不存在或无权限访问",
			}, nil
		}
		l.Errorf("查询地址失败: %v", err)
		return &user.DeleteAddressResp{
			Code:    int32(errcode.CommonServerError.Code()),
			Message: errcode.CommonServerError.Msg(),
		}, nil
	}

	// 3. 删除地址
	if err := l.svcCtx.UserAddressModel.Delete(l.ctx, uint64(in.AddressId)); err != nil {
		l.Errorf("删除地址失败: %v", err)
		return &user.DeleteAddressResp{
			Code:    int32(errcode.CommonServerError.Code()),
			Message: errcode.CommonServerError.Msg(),
		}, nil
	}

	l.Infof("地址删除成功，地址ID: %d", in.AddressId)

	return &user.DeleteAddressResp{
		Code:    int32(errcode.Success.Code()),
		Message: errcode.Success.Msg(),
	}, nil
}

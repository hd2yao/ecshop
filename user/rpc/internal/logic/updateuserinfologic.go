package logic

import (
	"context"
	"database/sql"
	"errors"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/hd2yao/ecshop/common/errcode"
	"github.com/hd2yao/ecshop/user/model"
	"github.com/hd2yao/ecshop/user/rpc/internal/svc"
	"github.com/hd2yao/ecshop/user/rpc/types/user"
)

type UpdateUserInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserInfoLogic {
	return &UpdateUserInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UpdateUserInfo 修改用户信息
// 使用缓存服务保证强一致性
func (l *UpdateUserInfoLogic) UpdateUserInfo(in *user.UpdateUserInfoReq) (*user.UpdateUserInfoResp, error) {
	// 1. 参数验证
	if in.UserId <= 0 {
		return &user.UpdateUserInfoResp{
			Code:    int32(errcode.CommonParamError.Code()),
			Message: "用户ID无效",
		}, nil
	}

	// 检查是否有需要更新的字段
	if in.Name == "" && in.Avatar == "" && in.Phone == "" && in.Sex < 0 {
		return &user.UpdateUserInfoResp{
			Code:    int32(errcode.CommonParamError.Code()),
			Message: "至少需要提供一个要更新的字段",
		}, nil
	}

	l.Infof("修改用户信息，用户ID: %d, 更新字段: name=%s, avatar=%s, phone=%s, sex=%d",
		in.UserId, in.Name, in.Avatar, in.Phone, in.Sex)

	// 2. 查询用户信息（从数据库查询，不使用缓存，因为我们要更新它）
	userInfo, err := l.svcCtx.UserModel.FindOne(l.ctx, uint64(in.UserId))
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			l.Infof("用户不存在，用户ID: %d", in.UserId)
			return &user.UpdateUserInfoResp{
				Code:    int32(errcode.UserAccountUnregister.Code()),
				Message: errcode.UserAccountUnregister.Msg(),
			}, nil
		}
		l.Errorf("查询用户信息失败: %v", err)
		return &user.UpdateUserInfoResp{
			Code:    int32(errcode.CommonServerError.Code()),
			Message: errcode.CommonServerError.Msg(),
		}, nil
	}

	// 3. 更新字段（只更新提供的字段）
	if in.Name != "" {
		userInfo.Name = sql.NullString{String: in.Name, Valid: true}
	}
	if in.Avatar != "" {
		userInfo.Avatar = sql.NullString{String: in.Avatar, Valid: true}
	}
	if in.Phone != "" {
		userInfo.Phone = sql.NullString{String: in.Phone, Valid: true}
	}
	if in.Sex >= 0 {
		userInfo.Sex = int64(in.Sex)
	}

	// 4. 使用缓存服务更新用户信息（保证强一致性）
	err = l.svcCtx.UserModel.CacheService().UpdateUserWithCache(l.ctx, userInfo)
	if err != nil {
		l.Errorf("更新用户信息失败: %v", err)
		return &user.UpdateUserInfoResp{
			Code:    int32(errcode.CommonServerError.Code()),
			Message: errcode.CommonServerError.Msg(),
		}, nil
	}

	// 5. 重新查询更新后的用户信息（使用缓存服务）
	updatedUserInfo, err := l.svcCtx.UserModel.CacheService().GetUserById(l.ctx, uint64(in.UserId))
	if err != nil {
		l.Errorf("查询更新后的用户信息失败: %v", err)
		return &user.UpdateUserInfoResp{
			Code:    int32(errcode.CommonServerError.Code()),
			Message: errcode.CommonServerError.Msg(),
		}, nil
	}

	// 6. 返回更新后的用户信息
	return &user.UpdateUserInfoResp{
		Code:    int32(errcode.Success.Code()),
		Message: errcode.Success.Msg(),
		UserInfo: &user.UserInfo{
			Id:         int64(updatedUserInfo.Id),
			Name:       updatedUserInfo.Name,
			Avatar:     updatedUserInfo.Avatar,
			Email:      updatedUserInfo.Mail,
			Phone:      updatedUserInfo.Phone,
			Sex:        int32(updatedUserInfo.Sex),
			Points:     int32(updatedUserInfo.Points),
			CreateTime: updatedUserInfo.CreateTime.Format("2006-01-02 15:04:05"),
		},
	}, nil
}


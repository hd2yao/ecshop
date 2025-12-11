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

type GetUserInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserInfoLogic {
	return &GetUserInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetUserInfo 获取用户信息
func (l *GetUserInfoLogic) GetUserInfo(in *user.GetUserInfoReq) (*user.GetUserInfoResp, error) {
	// 1. 参数验证
	if in.UserId <= 0 {
		return &user.GetUserInfoResp{
			Code:    int32(errcode.CommonParamError.Code()),
			Message: "用户ID无效",
		}, nil
	}

	l.Infof("查询用户信息，用户ID: %d", in.UserId)

	// 2. 优先从缓存查询用户信息
	userDTO, source, err := l.svcCtx.UserModel.CacheService().GetUserByIdWithSource(l.ctx, uint64(in.UserId))
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			// 用户不存在
			l.Infof("用户不存在，用户ID: %d", in.UserId)
			return &user.GetUserInfoResp{
				Code:    int32(errcode.UserAccountUnregister.Code()),
				Message: errcode.UserAccountUnregister.Msg(),
			}, nil
		}
		// 数据库错误
		l.Errorf("查询用户信息失败: %v", err)
		return &user.GetUserInfoResp{
			Code:    int32(errcode.CommonServerError.Code()),
			Message: errcode.CommonServerError.Msg(),
		}, nil
	}

	if userDTO == nil {
		l.Infof("缓存未命中且数据库未查询到用户，用户ID: %d", in.UserId)
		return &user.GetUserInfoResp{
			Code:    int32(errcode.UserAccountUnregister.Code()),
			Message: errcode.UserAccountUnregister.Msg(),
		}, nil
	}

	switch source {
	case model.DataSourceCache:
		l.Infof("用户信息来自缓存，用户ID: %d", in.UserId)
	case model.DataSourceDatabase:
		l.Infof("用户信息来自数据库，用户ID: %d", in.UserId)
	default:
		l.Infof("用户信息来源未知(%s)，用户ID: %d", source, in.UserId)
	}

	createTime := ""
	if !userDTO.CreateTime.IsZero() {
		createTime = userDTO.CreateTime.Format("2006-01-02 15:04:05")
	}

	// 3. 获取关注数和粉丝数（带缓存）
	followCount, followerCount, err := l.svcCtx.GetFollowStat(l.ctx, in.UserId)
	if err != nil {
		l.Errorf("获取关注统计失败: %v", err)
		// 关注统计获取失败不影响用户信息返回，使用默认值 0
		followCount = 0
		followerCount = 0
	}

	// 4. 返回用户信息
	return &user.GetUserInfoResp{
		Code:    int32(errcode.Success.Code()),
		Message: errcode.Success.Msg(),
		UserInfo: &user.UserInfo{
			Id:            int64(userDTO.Id),
			Name:          userDTO.Name,
			Avatar:        userDTO.Avatar,
			Email:         userDTO.Mail,
			Phone:         userDTO.Phone,
			Sex:           int32(userDTO.Sex),
			Points:        int32(userDTO.Points),
			CreateTime:    createTime,
			FollowCount:   followCount,
			FollowerCount: followerCount,
		},
	}, nil
}

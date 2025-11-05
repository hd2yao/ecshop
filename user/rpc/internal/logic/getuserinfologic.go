package logic

import (
	"context"
	"database/sql"
	"errors"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/hd2yao/ecshop/common/errcode"
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

	// 2. 从数据库查询用户信息
	userInfo, err := l.svcCtx.UserModel.FindOne(l.ctx, uint64(in.UserId))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
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

	// 3. 返回用户信息
	return &user.GetUserInfoResp{
		Code:    int32(errcode.Success.Code()),
		Message: errcode.Success.Msg(),
		UserInfo: &user.UserInfo{
			Id:         int64(userInfo.Id),
			Name:       userInfo.Name.String,
			Avatar:     userInfo.Avatar.String,
			Email:      userInfo.Mail.String,
			Phone:      userInfo.Phone.String,
			Sex:        int32(userInfo.Sex),
			Points:     int32(userInfo.Points),
			CreateTime: userInfo.CreateTime.Time.Format("2006-01-02 15:04:05"),
		},
	}, nil
}

package user

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/hd2yao/ecshop/common/errcode"
	"github.com/hd2yao/ecshop/common/middleware"
	"github.com/hd2yao/ecshop/user/api/internal/svc"
	"github.com/hd2yao/ecshop/user/api/internal/types"
	"github.com/hd2yao/ecshop/user/rpc/types/user"
)

type GetUserInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGetUserInfoLogic 获取当前登录用户信息
func NewGetUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserInfoLogic {
	return &GetUserInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserInfoLogic) GetUserInfo() (resp *types.GetUserInfoResponse, err error) {
	// 1. 从 context 中获取用户 ID（JWT 中间件已验证并存入）
	userID, ok := middleware.GetUserIDFromContext(l.ctx)
	if !ok {
		// 正常情况下不会到这里，因为中间件已经验证过 token
		return &types.GetUserInfoResponse{
			Code:    errcode.UserTokenInvalid.Code(),
			Message: errcode.UserTokenInvalid.Msg(),
		}, nil
	}

	l.Infof("获取用户信息，用户ID: %d", userID)

	// 2. 调用 RPC 服务获取用户信息
	rpcResp, err := l.svcCtx.UserRpc.GetUserInfo(l.ctx, &user.GetUserInfoReq{
		UserId: userID,
	})

	if err != nil {
		l.Errorf("调用 RPC 获取用户信息失败: %v", err)
		return &types.GetUserInfoResponse{
			Code:    errcode.CommonServerError.Code(),
			Message: errcode.CommonServerError.Msg(),
		}, nil
	}

	// 3. 转换响应
	resp = &types.GetUserInfoResponse{
		Code:    int(rpcResp.Code),
		Message: rpcResp.Message,
	}

	// 4. 如果成功，填充用户信息
	if rpcResp.Code == int32(errcode.Success.Code()) && rpcResp.UserInfo != nil {
		resp.UserInfo = types.UserInfo{
			Id:         rpcResp.UserInfo.Id,
			Name:       rpcResp.UserInfo.Name,
			Avatar:     rpcResp.UserInfo.Avatar,
			Email:      rpcResp.UserInfo.Email,
			Phone:      rpcResp.UserInfo.Phone,
			Sex:        int(rpcResp.UserInfo.Sex),
			Points:     int(rpcResp.UserInfo.Points),
			CreateTime: rpcResp.UserInfo.CreateTime,
		}
	}

	return resp, nil
}

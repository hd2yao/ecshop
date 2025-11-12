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

type UpdateUserInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserInfoLogic {
	return &UpdateUserInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateUserInfoLogic) UpdateUserInfo(req *types.UpdateUserInfoRequest) (resp *types.UpdateUserInfoResponse, err error) {
	// 1. 从 context 中获取用户 ID（JWT 中间件已验证并存入）
	userID, ok := middleware.GetUserIDFromContext(l.ctx)
	if !ok {
		// 正常情况下不会到这里，因为中间件已经验证过 token
		return &types.UpdateUserInfoResponse{
			Code:    errcode.UserTokenInvalid.Code(),
			Message: errcode.UserTokenInvalid.Msg(),
		}, nil
	}

	l.Infof("修改用户信息，用户ID: %d, 更新字段: name=%s, avatar=%s, phone=%s, sex=%d",
		userID, req.Name, req.Avatar, req.Phone, req.Sex)

	// 2. 调用 RPC 服务修改用户信息
	rpcResp, err := l.svcCtx.UserRpc.UpdateUserInfo(l.ctx, &user.UpdateUserInfoReq{
		UserId: userID,
		Name:   req.Name,
		Avatar: req.Avatar,
		Phone:  req.Phone,
		Sex:    int32(req.Sex),
	})

	if err != nil {
		l.Errorf("调用 RPC 修改用户信息失败: %v", err)
		return &types.UpdateUserInfoResponse{
			Code:    errcode.CommonServerError.Code(),
			Message: errcode.CommonServerError.Msg(),
		}, nil
	}

	// 3. 转换响应
	resp = &types.UpdateUserInfoResponse{
		Code:    int(rpcResp.Code),
		Message: rpcResp.Message,
	}

	// 4. 如果成功，填充更新后的用户信息
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


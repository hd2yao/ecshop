package user

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/hd2yao/ecshop/common/errcode"
	"github.com/hd2yao/ecshop/user/api/internal/svc"
	"github.com/hd2yao/ecshop/user/api/internal/types"
	"github.com/hd2yao/ecshop/user/rpc/types/user"
)

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewLoginLogic 用户登录
func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginLogic) Login(req *types.LoginRequest) (resp *types.LoginResponse, err error) {
	// 1. 参数验证
	if req.Email == "" && req.Phone == "" {
		return &types.LoginResponse{
			Code:    errcode.CommonParamError.Code(),
			Message: "邮箱和手机号至少提供一个",
		}, nil
	}

	if req.Password == "" {
		return &types.LoginResponse{
			Code:    errcode.CommonParamError.Code(),
			Message: "密码不能为空",
		}, nil
	}

	// 2. 调用 RPC 服务进行登录
	loginResp, err := l.svcCtx.UserRpc.Login(l.ctx, &user.LoginReq{
		Email:    req.Email,
		Phone:    req.Phone,
		Password: req.Password,
	})

	if err != nil {
		l.Errorf("调用 RPC 登录服务失败: %v", err)
		return &types.LoginResponse{
			Code:    errcode.CommonServerError.Code(),
			Message: errcode.CommonServerError.Msg(),
		}, nil
	}

	// 3. 转换响应
	resp = &types.LoginResponse{
		Code:    int(loginResp.Code),
		Message: loginResp.Message,
	}

	// 4. 如果登录成功，填充用户信息和Token
	if loginResp.Code == int32(errcode.Success.Code()) {
		resp.UserId = loginResp.UserId
		resp.AccessToken = loginResp.AccessToken
		resp.RefreshToken = loginResp.RefreshToken
		resp.AccessTokenExpireTime = loginResp.AccessTokenExpireTime
		if loginResp.UserInfo != nil {
			resp.UserInfo = types.UserInfo{
				Id:         loginResp.UserInfo.Id,
				Name:       loginResp.UserInfo.Name,
				Avatar:     loginResp.UserInfo.Avatar,
				Email:      loginResp.UserInfo.Email,
				Phone:      loginResp.UserInfo.Phone,
				Sex:        int(loginResp.UserInfo.Sex),
				Points:     int(loginResp.UserInfo.Points),
				CreateTime: loginResp.UserInfo.CreateTime,
			}
		}
	}

	return resp, nil
}

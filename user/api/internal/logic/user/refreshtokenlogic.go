package user

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/hd2yao/ecshop/common/errcode"
	"github.com/hd2yao/ecshop/user/api/internal/svc"
	"github.com/hd2yao/ecshop/user/api/internal/types"
	"github.com/hd2yao/ecshop/user/rpc/types/user"
)

type RefreshTokenLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewRefreshTokenLogic 刷新Token
func NewRefreshTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefreshTokenLogic {
	return &RefreshTokenLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RefreshTokenLogic) RefreshToken(req *types.RefreshTokenRequest) (resp *types.RefreshTokenResponse, err error) {
	// 1. 参数验证
	if req.RefreshToken == "" {
		return &types.RefreshTokenResponse{
			Code:    errcode.CommonParamError.Code(),
			Message: "Refresh Token 不能为空",
		}, nil
	}

	// 2. 调用 RPC 服务刷新Token
	refreshResp, err := l.svcCtx.UserRpc.RefreshToken(l.ctx, &user.RefreshTokenReq{
		RefreshToken: req.RefreshToken,
	})

	if err != nil {
		l.Errorf("调用 RPC 刷新Token服务失败: %v", err)
		return &types.RefreshTokenResponse{
			Code:    errcode.CommonServerError.Code(),
			Message: errcode.CommonServerError.Msg(),
		}, nil
	}

	// 3. 转换响应
	resp = &types.RefreshTokenResponse{
		Code:    int(refreshResp.Code),
		Message: refreshResp.Message,
	}

	// 4. 如果刷新成功，填充Token
	if refreshResp.Code == int32(errcode.Success.Code()) {
		resp.AccessToken = refreshResp.AccessToken
		resp.RefreshToken = refreshResp.RefreshToken
	}

	return resp, nil
}


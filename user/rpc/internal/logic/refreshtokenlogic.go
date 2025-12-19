package logic

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/hd2yao/ecshop/common/errcode"
	"github.com/hd2yao/ecshop/common/utils/jwt"
	"github.com/hd2yao/ecshop/user/model"
	"github.com/hd2yao/ecshop/user/rpc/internal/svc"
	"github.com/hd2yao/ecshop/user/rpc/types/user"
)

type RefreshTokenLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRefreshTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefreshTokenLogic {
	return &RefreshTokenLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// RefreshToken 刷新Token
func (l *RefreshTokenLogic) RefreshToken(in *user.RefreshTokenReq) (*user.RefreshTokenResp, error) {
	// 1. 验证 Refresh Token 不能为空
	if in.RefreshToken == "" {
		return &user.RefreshTokenResp{
			Code:    int32(errcode.CommonParamError.Code()),
			Message: "Refresh Token 不能为空",
		}, nil
	}

	// 2. 解析 Refresh Token
	claims, err := jwt.ParseRefreshToken(in.RefreshToken)
	if err != nil {
		l.Errorf("解析 Refresh Token 失败: %v", err)
		// 根据错误类型返回不同的错误码
		errMsg := err.Error()
		if errMsg == "token 已过期" {
			return &user.RefreshTokenResp{
				Code:    int32(errcode.UserTokenExpired.Code()),
				Message: errcode.UserTokenExpired.Msg(),
			}, nil
		} else if errMsg == "token 格式错误" {
			return &user.RefreshTokenResp{
				Code:    int32(errcode.UserTokenMalformed.Code()),
				Message: errcode.UserTokenMalformed.Msg(),
			}, nil
		}
		return &user.RefreshTokenResp{
			Code:    int32(errcode.UserTokenInvalid.Code()),
			Message: errcode.UserTokenInvalid.Msg(),
		}, nil
	}

	// 3. 从 Redis 验证 Refresh Token 是否有效
	var storedToken string
	if err := l.svcCtx.RefreshTokenCache.Get(l.ctx, claims.Email, &storedToken); err != nil || storedToken != in.RefreshToken {
		l.Errorf("Refresh Token 在 Redis 中不存在或不匹配: userId=%d, email=%s", claims.UserID, claims.Email)
		return &user.RefreshTokenResp{
			Code:    int32(errcode.UserRefreshTokenInvalid.Code()),
			Message: errcode.UserRefreshTokenInvalid.Msg(),
		}, nil
	}

	// 4. 验证用户是否存在
	foundUser, err := l.svcCtx.UserModel.FindOne(l.ctx, uint64(claims.UserID))
	if err != nil {
		if errors.Is(err, model.ErrNotFound) || errors.Is(err, sql.ErrNoRows) {
			l.Errorf("用户不存在: userId=%d", claims.UserID)
			return &user.RefreshTokenResp{
				Code:    int32(errcode.UserAccountUnregister.Code()),
				Message: errcode.UserAccountUnregister.Msg(),
			}, nil
		}
		l.Errorf("查询用户失败: %v", err)
		return &user.RefreshTokenResp{
			Code:    int32(errcode.CommonServerError.Code()),
			Message: errcode.CommonServerError.Msg(),
		}, nil
	}

	// 5. 生成新的 Access Token
	newAccessToken, err := jwt.GenerateToken(claims.UserID, foundUser.Name.String, foundUser.Mail.String)
	if err != nil {
		l.Errorf("生成新的 Access Token 失败: %v", err)
		return &user.RefreshTokenResp{
			Code:    int32(errcode.CommonServerError.Code()),
			Message: "生成令牌失败",
		}, nil
	}

	// 6. 生成新的 Refresh Token（滚动刷新策略）
	newRefreshToken, err := jwt.GenerateRefreshToken(claims.UserID, foundUser.Name.String, foundUser.Mail.String)
	if err != nil {
		l.Errorf("生成新的 Refresh Token 失败: %v", err)
		return &user.RefreshTokenResp{
			Code:    int32(errcode.CommonServerError.Code()),
			Message: "生成刷新令牌失败",
		}, nil
	}

	// 7. 更新 Redis 中的 Refresh Token
	if err := l.svcCtx.RefreshTokenCache.Set(l.ctx, claims.Email, newRefreshToken, jwt.RefreshTokenExpiration); err != nil {
		l.Errorf("更新 Redis 中的 Refresh Token 失败: %v", err)
		// 不影响刷新流程，只记录日志
	}

	l.Infof("Token刷新成功: userId=%d", claims.UserID)

	return &user.RefreshTokenResp{
		Code:                  int32(errcode.Success.Code()),
		Message:               errcode.Success.Msg(),
		AccessToken:           newAccessToken,
		RefreshToken:          newRefreshToken,
		AccessTokenExpireTime: time.Now().Add(jwt.AccessTokenExpiration).Format(time.RFC3339),
	}, nil
}

package logic

import (
	"context"
	"database/sql"
	"errors"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"

	"github.com/hd2yao/ecshop/common/errcode"
	"github.com/hd2yao/ecshop/common/utils"
	"github.com/hd2yao/ecshop/user/model"
	"github.com/hd2yao/ecshop/user/rpc/internal/svc"
	"github.com/hd2yao/ecshop/user/rpc/types/user"
)

type LoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Login 用户登录
func (l *LoginLogic) Login(in *user.LoginReq) (*user.LoginResp, error) {
	// 1. 参数验证：邮箱和手机号至少提供一个
	if in.Email == "" && in.Phone == "" {
		return &user.LoginResp{
			Code:    int32(errcode.CommonParamError.Code()),
			Message: "邮箱和手机号至少提供一个",
		}, nil
	}

	// 2. 验证密码不能为空
	if in.Password == "" {
		return &user.LoginResp{
			Code:    int32(errcode.CommonParamError.Code()),
			Message: "密码不能为空",
		}, nil
	}

	// 3. 根据邮箱或手机号查询用户
	var foundUser *model.User
	var err error

	if in.Email != "" {
		// 通过邮箱查询
		foundUser, err = l.svcCtx.UserModel.FindOneByMail(l.ctx, sql.NullString{String: in.Email, Valid: true})
		if err != nil {
			if errors.Is(err, model.ErrNotFound) || errors.Is(err, sql.ErrNoRows) {
				l.Errorf("用户不存在: email=%s", in.Email)
				return &user.LoginResp{
					Code:    int32(errcode.UserAccountUnregister.Code()),
					Message: errcode.UserAccountUnregister.Msg(),
				}, nil
			}
			l.Errorf("查询用户失败: %v", err)
			return &user.LoginResp{
				Code:    int32(errcode.CommonServerError.Code()),
				Message: errcode.CommonServerError.Msg(),
			}, nil
		}
	} else {
		// 通过手机号查询
		foundUser, err = l.svcCtx.UserModel.FindOneByPhone(l.ctx, sql.NullString{String: in.Phone, Valid: true})
		if err != nil {
			if errors.Is(err, model.ErrNotFound) || errors.Is(err, sql.ErrNoRows) {
				l.Errorf("用户不存在: phone=%s", in.Phone)
				return &user.LoginResp{
					Code:    int32(errcode.UserAccountUnregister.Code()),
					Message: errcode.UserAccountUnregister.Msg(),
				}, nil
			}
			l.Errorf("查询用户失败: %v", err)
			return &user.LoginResp{
				Code:    int32(errcode.CommonServerError.Code()),
				Message: errcode.CommonServerError.Msg(),
			}, nil
		}
	}

	// 4. 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(foundUser.Pwd.String), []byte(in.Password)); err != nil {
		l.Errorf("密码验证失败: userId=%d", foundUser.Id)
		return &user.LoginResp{
			Code:    int32(errcode.UserAccountPwdError.Code()),
			Message: errcode.UserAccountPwdError.Msg(),
		}, nil
	}

	// 5. 生成 JWT Token (默认7天过期)
	token, err := jwt.GenerateToken(int64(foundUser.Id), foundUser.Name.String, foundUser.Mail.String)
	if err != nil {
		l.Errorf("生成 token 失败: %v", err)
		return &user.LoginResp{
			Code:    int32(errcode.CommonServerError.Code()),
			Message: "生成令牌失败",
		}, nil
	}

	// 6. 构建用户信息响应
	userInfo := &user.UserInfo{
		Id:         int64(foundUser.Id),
		Name:       foundUser.Name.String,
		Avatar:     foundUser.Avatar.String,
		Email:      foundUser.Mail.String,
		Phone:      foundUser.Phone.String,
		Sex:        int32(foundUser.Sex),
		Points:     int32(foundUser.Points),
		CreateTime: foundUser.CreateTime.Time.Format("2006-01-02 15:04:05"),
	}

	l.Infof("用户登录成功: userId=%d, email=%s, phone=%s", foundUser.Id, in.Email, in.Phone)

	return &user.LoginResp{
		Code:     int32(errcode.Success.Code()),
		Message:  errcode.Success.Msg(),
		UserId:   int64(foundUser.Id),
		Token:    token,
		UserInfo: userInfo,
	}, nil
}

package logic

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"regexp"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"

	"github.com/hd2yao/ecshop/user/model"
	"github.com/hd2yao/ecshop/user/rpc/internal/svc"
	"github.com/hd2yao/ecshop/user/rpc/types/user"
)

type RegisterLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Register 用户注册
func (l *RegisterLogic) Register(in *user.RegisterReq) (*user.RegisterResp, error) {
	// 1. 参数验证
	if in.Email == "" || in.EmailCode == "" || in.Password == "" || in.Name == "" {
		return &user.RegisterResp{
			Code:    400,
			Message: "邮箱、验证码、密码和昵称不能为空",
		}, nil
	}

	// 2. 验证邮箱格式
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(in.Email) {
		return &user.RegisterResp{
			Code:    400,
			Message: "邮箱格式不正确",
		}, nil
	}

	// 3. 验证密码长度
	if len(in.Password) < 6 {
		return &user.RegisterResp{
			Code:    400,
			Message: "密码长度至少6位",
		}, nil
	}

	// 4. 验证昵称长度
	if len(in.Name) < 2 || len(in.Name) > 32 {
		return &user.RegisterResp{
			Code:    400,
			Message: "昵称长度应在2-32位之间",
		}, nil
	}

	// 5. 验证邮箱验证码
	if !l.svcCtx.MailService.VerifyCode(in.Email, in.EmailCode) {
		l.Errorf("邮箱验证码验证失败: email=%s, code=%s", in.Email, in.EmailCode)
		return &user.RegisterResp{
			Code:    400,
			Message: "邮箱验证码错误或已过期",
		}, nil
	}

	// 6. 检查邮箱是否已注册 (通过数据库唯一索引也会报错，但这里提前检查给出友好提示)
	existingUser, err := l.svcCtx.UserModel.FindOneByMail(l.ctx, sql.NullString{String: in.Email, Valid: true})
	if err != nil && err != sql.ErrNoRows {
		l.Errorf("查询用户失败: %v", err)
		return &user.RegisterResp{
			Code:    500,
			Message: "系统错误",
		}, nil
	}
	if existingUser != nil {
		return &user.RegisterResp{
			Code:    400,
			Message: "该邮箱已被注册",
		}, nil
	}

	// 7. 检查手机号是否已注册（如果提供了手机号）
	if in.Phone != "" {
		existingUserByPhone, err := l.svcCtx.UserModel.FindOneByPhone(l.ctx, sql.NullString{String: in.Phone, Valid: true})
		if err != nil && err != sql.ErrNoRows {
			l.Errorf("查询用户失败: %v", err)
			return &user.RegisterResp{
				Code:    500,
				Message: "系统错误",
			}, nil
		}
		if existingUserByPhone != nil {
			return &user.RegisterResp{
				Code:    400,
				Message: "该手机号已被注册",
			}, nil
		}
	}

	// 8. 密码加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		l.Errorf("密码加密失败: %v", err)
		return &user.RegisterResp{
			Code:    500,
			Message: "系统错误",
		}, nil
	}

	// 9. 生成用户密钥
	secret, err := generateSecret()
	if err != nil {
		l.Errorf("生成用户密钥失败: %v", err)
		return &user.RegisterResp{
			Code:    500,
			Message: "系统错误",
		}, nil
	}

	// 10. 设置默认性别
	sex := in.Sex
	if sex != 0 && sex != 1 {
		sex = 1 // 默认男性
	}

	// 11. 插入用户信息
	now := time.Now()
	newUser := &model.User{
		Name:       sql.NullString{String: in.Name, Valid: true},
		Pwd:        sql.NullString{String: string(hashedPassword), Valid: true},
		Sex:        int64(sex),
		Points:     0,
		Mail:       sql.NullString{String: in.Email, Valid: true},
		Secret:     sql.NullString{String: secret, Valid: true},
		CreateTime: sql.NullTime{Time: now, Valid: true},
		UpdateTime: sql.NullTime{Time: now, Valid: true},
	}

	// 设置手机号（如果提供）
	if in.Phone != "" {
		newUser.Phone = sql.NullString{String: in.Phone, Valid: true}
	}

	result, err := l.svcCtx.UserModel.Insert(l.ctx, newUser)
	if err != nil {
		l.Errorf("插入用户失败: %v", err)
		return &user.RegisterResp{
			Code:    500,
			Message: "注册失败",
		}, nil
	}

	userId, err := result.LastInsertId()
	if err != nil {
		l.Errorf("获取用户ID失败: %v", err)
		return &user.RegisterResp{
			Code:    500,
			Message: "注册失败",
		}, nil
	}

	// 12. 生成访问令牌 (这里简单使用用户ID，实际项目中应该使用JWT)
	token := generateToken(userId)

	// 13. 返回用户信息
	userInfo := &user.UserInfo{
		Id:         userId,
		Name:       in.Name,
		Avatar:     "", // 默认头像为空
		Email:      in.Email,
		Phone:      in.Phone,
		Sex:        sex,
		Points:     0,
		CreateTime: now.Format("2006-01-02 15:04:05"),
	}

	l.Infof("用户注册成功: userId=%d, email=%s", userId, in.Email)

	return &user.RegisterResp{
		Code:     200,
		Message:  "注册成功",
		UserId:   userId,
		Token:    token,
		UserInfo: userInfo,
	}, nil
}

// generateSecret 生成12位随机密钥
func generateSecret() (string, error) {
	bytes := make([]byte, 6)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// generateToken 生成访问令牌 (简单实现，实际项目应使用JWT)
func generateToken(userId int64) string {
	return fmt.Sprintf("token_%d_%d", userId, time.Now().Unix())
}

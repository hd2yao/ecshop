package logic

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"

	"github.com/hd2yao/ecshop/common/errcode"
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
			Code:    int32(errcode.CommonParamError.Code()),
			Message: errcode.CommonParamError.Msg(),
		}, nil
	}

	// 2. 验证邮箱格式
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(in.Email) {
		return &user.RegisterResp{
			Code:    int32(errcode.CommonParamError.Code()),
			Message: "邮箱格式不正确",
		}, nil
	}

	// 3. 验证密码长度
	if len(in.Password) < 6 {
		return &user.RegisterResp{
			Code:    int32(errcode.CommonParamError.Code()),
			Message: "密码长度至少6位",
		}, nil
	}

	// 4. 验证昵称长度
	if len(in.Name) < 2 || len(in.Name) > 32 {
		return &user.RegisterResp{
			Code:    int32(errcode.CommonParamError.Code()),
			Message: "昵称长度应在2-32位之间",
		}, nil
	}

	// 5. 验证邮箱验证码
	if !l.svcCtx.MailService.VerifyCode(in.Email, in.EmailCode) {
		l.Errorf("邮箱验证码验证失败: email=%s, code=%s", in.Email, in.EmailCode)
		return &user.RegisterResp{
			Code:    int32(errcode.UserCodeEmailError.Code()),
			Message: errcode.UserCodeEmailError.Msg(),
		}, nil
	}

	// 6. 检查邮箱是否已注册 (通过数据库唯一索引也会报错，但这里提前检查给出友好提示)
	existingUser, err := l.svcCtx.UserModel.FindOneByMail(l.ctx, sql.NullString{String: in.Email, Valid: true})
	if err != nil && err != sql.ErrNoRows {
		l.Errorf("查询用户失败: %v", err)
		return &user.RegisterResp{
			Code:    int32(errcode.CommonServerError.Code()),
			Message: errcode.CommonServerError.Msg(),
		}, nil
	}
	if existingUser != nil {
		return &user.RegisterResp{
			Code:    int32(errcode.UserAccountExist.Code()),
			Message: errcode.UserAccountExist.Msg(),
		}, nil
	}

	// 7. 检查手机号是否已注册（如果提供了手机号）
	if in.Phone != "" {
		existingUserByPhone, err := l.svcCtx.UserModel.FindOneByPhone(l.ctx, sql.NullString{String: in.Phone, Valid: true})
		if err != nil && err != sql.ErrNoRows {
			l.Errorf("查询用户失败: %v", err)
			return &user.RegisterResp{
				Code:    int32(errcode.CommonServerError.Code()),
				Message: errcode.CommonServerError.Msg(),
			}, nil
		}
		if existingUserByPhone != nil {
			return &user.RegisterResp{
				Code:    int32(errcode.UserAccountExist.Code()),
				Message: "该手机号已被注册",
			}, nil
		}
	}

	// 8. 密码加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		l.Errorf("密码加密失败: %v", err)
		return &user.RegisterResp{
			Code:    int32(errcode.CommonServerError.Code()),
			Message: errcode.CommonServerError.Msg(),
		}, nil
	}

	// 9. 生成用户密钥
	secret, err := generateSecret()
	if err != nil {
		l.Errorf("生成用户密钥失败: %v", err)
		return &user.RegisterResp{
			Code:    int32(errcode.CommonServerError.Code()),
			Message: errcode.CommonServerError.Msg(),
		}, nil
	}

	// 10. 处理头像（临时文件移动到正式目录）
	var avatarUrl string
	if in.Avatar != "" && isValidURL(in.Avatar) {
		// 检查是否为临时文件URL
		if strings.Contains(in.Avatar, "/temp/") {
			l.Infof("检测到临时头像URL，准备移动到正式目录: %s", in.Avatar)
			// 先插入用户获取userId，然后移动文件
			avatarUrl = in.Avatar // 暂时使用原URL，后面会更新
		} else {
			// 直接使用提供的URL
			avatarUrl = in.Avatar
		}
	}

	// 11. 插入用户信息
	now := time.Now()
	newUser := &model.User{
		Name:       sql.NullString{String: in.Name, Valid: true},
		Pwd:        sql.NullString{String: string(hashedPassword), Valid: true},
		Phone:      sql.NullString{String: in.Phone, Valid: true},
		Avatar:     sql.NullString{String: avatarUrl, Valid: true},
		Sex:        int64(in.Sex),
		Points:     0,
		Mail:       sql.NullString{String: in.Email, Valid: true},
		Secret:     sql.NullString{String: secret, Valid: true},
		CreateTime: sql.NullTime{Time: now, Valid: true},
		UpdateTime: sql.NullTime{Time: now, Valid: true},
	}

	result, err := l.svcCtx.UserModel.Insert(l.ctx, newUser)
	if err != nil {
		l.Errorf("插入用户失败: %v", err)
		return &user.RegisterResp{
			Code:    int32(errcode.CommonServerError.Code()),
			Message: "注册失败",
		}, nil
	}

	userId, err := result.LastInsertId()
	if err != nil {
		l.Errorf("获取用户ID失败: %v", err)
		return &user.RegisterResp{
			Code:    int32(errcode.CommonServerError.Code()),
			Message: "注册失败",
		}, nil
	}

	// 12. 处理临时头像文件移动
	if in.Avatar != "" && strings.Contains(in.Avatar, "/temp/") {
		finalUrl, err := l.moveTempAvatarToFinal(in.Avatar, userId)
		if err != nil {
			l.Errorf("移动临时头像失败: %v", err)
			// 移动失败不影响注册，继续使用临时URL
		} else {
			// 更新头像URL为正式URL
			avatarUrl = finalUrl
			// 更新数据库中的头像字段
			updatedUser := &model.User{
				Id:     uint64(userId),
				Avatar: sql.NullString{String: finalUrl, Valid: true},
			}
			if updateErr := l.svcCtx.UserModel.Update(l.ctx, updatedUser); updateErr != nil {
				l.Errorf("更新用户头像失败: %v", updateErr)
			} else {
				l.Infof("临时头像移动成功: %s -> %s", in.Avatar, finalUrl)
			}
		}
	}

	// 13. 返回用户信息
	userInfo := &user.UserInfo{
		Id:         userId,
		Name:       in.Name,
		Avatar:     avatarUrl, // 头像URL
		Email:      in.Email,
		Phone:      in.Phone,
		Sex:        in.Sex,
		Points:     0,
		CreateTime: now.Format("2006-01-02 15:04:05"),
	}

	l.Infof("用户注册成功: userId=%d, email=%s", userId, in.Email)

	return &user.RegisterResp{
		Code:     int32(errcode.Success.Code()),
		Message:  errcode.Success.Msg(),
		UserId:   userId,
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

// isValidURL 简单的URL验证
func isValidURL(str string) bool {
	return strings.HasPrefix(str, "http://") || strings.HasPrefix(str, "https://")
}

// moveTempAvatarToFinal 将临时头像文件移动到正式目录
func (l *RegisterLogic) moveTempAvatarToFinal(tempUrl string, userId int64) (string, error) {
	// 1. 从URL中提取相对路径
	// 例如: https://hd2yao.oss-cn-shanghai.aliyuncs.com/avatars/temp/2025/09/18/ffec0039-c05e-4fe6-a721-30a6af7b5707.png
	// 需要提取: avatars/temp/2025/09/18/ffec0039-c05e-4fe6-a721-30a6af7b5707.png

	// 找到域名后的路径部分
	parts := strings.SplitN(tempUrl, ".com/", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("无效的临时URL格式: %s", tempUrl)
	}

	relativePath := parts[1] // avatars/temp/2025/09/18/ffec0039-c05e-4fe6-a721-30a6af7b5707.png

	// 处理URL中可能包含的查询参数
	if strings.Contains(relativePath, "?") {
		relativePath = strings.Split(relativePath, "?")[0]
	}

	// 2. 从相对路径中提取文件名和扩展名
	pathParts := strings.Split(relativePath, "/")
	if len(pathParts) < 2 {
		return "", fmt.Errorf("无效的文件路径格式: %s", relativePath)
	}

	tempFileName := pathParts[len(pathParts)-1] // ffec0039-c05e-4fe6-a721-30a6af7b5707.png
	fileExt := filepath.Ext(tempFileName)       // .png

	// 3. 生成新的文件名和路径
	newFileName := fmt.Sprintf("avatar_%d_%d%s", userId, time.Now().UnixNano(), fileExt)

	// 4. 构建OSS中的源路径和目标路径
	sourcePath := relativePath // 直接使用完整的相对路径
	destPath := fmt.Sprintf("avatars/users/user_%d/%s", userId, newFileName)

	l.Infof("准备复制头像文件: 源路径=%s, 目标路径=%s", sourcePath, destPath)

	finalUrl, err := l.svcCtx.OssClient.Copy(sourcePath, destPath)
	if err != nil {
		return "", fmt.Errorf("复制文件失败: %v", err)
	}

	// 4. 删除临时文件
	if err := l.svcCtx.OssClient.Delete(sourcePath); err != nil {
		l.Errorf("删除临时文件失败: %v", err)
		// 不影响主流程，继续返回新URL
	}

	l.Infof("头像文件移动成功: %s -> %s", sourcePath, destPath)
	return finalUrl, nil
}

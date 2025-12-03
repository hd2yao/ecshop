package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/hd2yao/ecshop/common/errcode"
	"github.com/hd2yao/ecshop/common/utils/jwt"
)

// 定义 Context Key 类型，避免冲突
type contextKey string

const (
	// UserIDKey 用户 ID 在 context 中的 key
	UserIDKey contextKey = "user_id"
	// UsernameKey 用户名在 context 中的 key
	UsernameKey contextKey = "username"
	// EmailKey 邮箱在 context 中的 key
	EmailKey contextKey = "email"
)

// JWTAuthMiddleware JWT 认证中间件
// 通过 context.Context 在请求链路中传递用户信息
func JWTAuthMiddleware() rest.Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// 1. 从请求头中获取 Authorization token
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				err := errcode.UserTokenMalformed
				httpx.WriteJson(w, err.HttpStatusCode(), map[string]interface{}{
					"code":    err.Code(),
					"message": "请求头中缺少 Authorization",
				})
				return
			}

			// 2. 验证 token 格式：Bearer <token>
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				err := errcode.UserTokenMalformed
				httpx.WriteJson(w, err.HttpStatusCode(), map[string]interface{}{
					"code":    err.Code(),
					"message": err.Msg(),
				})
				return
			}

			tokenString := parts[1]

			// 3. 解析并验证 token
			claims, err := jwt.ParseToken(tokenString)
			if err != nil {
				// 根据错误类型返回不同的错误码
				var errCode *errcode.AppError
				if err.Error() == "token 已过期" {
					errCode = errcode.UserTokenExpired
				} else {
					errCode = errcode.UserTokenInvalid
				}
				httpx.WriteJson(w, errCode.HttpStatusCode(), map[string]interface{}{
					"code":    errCode.Code(),
					"message": errCode.Msg(),
				})
				return
			}

			// 4. 将用户信息存入 context, 会在整个请求处理链路中传递
			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, UsernameKey, claims.Username)
			ctx = context.WithValue(ctx, EmailKey, claims.Email)

			// 5. 将包含用户信息的 context 传递给下一个处理器
			next(w, r.WithContext(ctx))
		}
	}
}

// GetUserIDFromContext 从 context 中获取用户ID
// 这个函数用于在 handler 或 logic 中获取当前登录用户的ID
func GetUserIDFromContext(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(UserIDKey).(int64)
	return userID, ok
}

// GetUsernameFromContext 从 context 中获取用户名
func GetUsernameFromContext(ctx context.Context) (string, bool) {
	username, ok := ctx.Value(UsernameKey).(string)
	return username, ok
}

// GetEmailFromContext 从 context 中获取邮箱
func GetEmailFromContext(ctx context.Context) (string, bool) {
	email, ok := ctx.Value(EmailKey).(string)
	return email, ok
}

// MustGetUserIDFromContext 从 context 中获取用户ID，如果不存在则返回 0
func MustGetUserIDFromContext(ctx context.Context) int64 {
	userID, _ := GetUserIDFromContext(ctx)
	return userID
}

// MustGetUsernameFromContext 从 context 中获取用户名，如果不存在则返回空字符串
func MustGetUsernameFromContext(ctx context.Context) string {
	username, _ := GetUsernameFromContext(ctx)
	return username
}

// MustGetEmailFromContext 从 context 中获取邮箱，如果不存在则返回空字符串
func MustGetEmailFromContext(ctx context.Context) string {
	email, _ := GetEmailFromContext(ctx)
	return email
}

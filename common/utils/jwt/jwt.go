package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

const (
	secret                 = "ecshop-jwt-secret"
	refreshSecret          = "ecshop-refresh-secret"
	AccessTokenExpiration  = 2 * time.Hour      // Access Token 2小时
	RefreshTokenExpiration = 7 * 24 * time.Hour // Refresh Token 7天
)

type Claims struct {
	UserID    int64  `json:"user_id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	TokenType string `json:"token_type"` // "access" 或 "refresh"
	jwt.RegisteredClaims
}

// GenerateToken 生成 Access Token（2小时有效期）
func GenerateToken(userID int64, username, email string) (string, error) {
	return GenerateTokenWithExpire(userID, username, email, "access", AccessTokenExpiration)
}

// GenerateRefreshToken 生成 Refresh Token（7天有效期）
func GenerateRefreshToken(userID int64, username, email string) (string, error) {
	return GenerateTokenWithExpire(userID, username, email, "refresh", RefreshTokenExpiration)
}

// GenerateTokenWithExpire 生成指定过期时间的 JWT token
func GenerateTokenWithExpire(userID int64, username, email, tokenType string, expire time.Duration) (string, error) {
	claims := &Claims{
		UserID:    userID,
		Username:  username,
		Email:     email,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expire)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Subject:   "user token",
		},
	}

	// 根据 token 类型选择密钥
	var secretKey string
	if tokenType == "refresh" {
		secretKey = refreshSecret
	} else {
		secretKey = secret
	}

	// 创建 token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 签名并获取完整的编码后的字符串
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ParseToken 解析 Access Token
func ParseToken(tokenString string) (*Claims, error) {
	return parseTokenWithSecret(tokenString, secret, "access")
}

// ParseRefreshToken 解析 Refresh Token
func ParseRefreshToken(tokenString string) (*Claims, error) {
	return parseTokenWithSecret(tokenString, refreshSecret, "refresh")
}

// parseTokenWithSecret 使用指定密钥解析 token
func parseTokenWithSecret(tokenString, secretKey, expectedType string) (*Claims, error) {
	// 解析token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("意外的签名方法: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		var ve *jwt.ValidationError
		if errors.As(err, &ve) {
			if ve.Errors&jwt.ValidationErrorExpired != 0 {
				return nil, errors.New("token 已过期")
			}
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				return nil, errors.New("token 格式错误")
			}
		}
		return nil, errors.New("token 无效")
	}

	// 获取声明
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		// 验证 token 类型
		if claims.TokenType != expectedType {
			return nil, fmt.Errorf("token 类型错误: 期望 %s，实际 %s", expectedType, claims.TokenType)
		}
		return claims, nil
	}

	return nil, errors.New("token 无效")
}

// GetSecret 获取 Access Token 密钥（用于中间件）
func GetSecret() string {
	return secret
}

package jwt

import (
	"github.com/golang-jwt/jwt/v4"
	"github.com/jovian1994/cxh-1207-be-interview/apps/translation/config"
	"github.com/jovian1994/cxh-1207-be-interview/pkg/logger"
	"time"
)

type ITokenVerify interface {
	GenerateJWT(username string) (string, error)
	ValidateJWT(tokenString string) (string, error)
}

type tokenVerify struct {
}

func NewTokenVerify() ITokenVerify {
	return &tokenVerify{}
}

var defaultJwtSecret = "your-secret-key"

func (t *tokenVerify) GenerateJWT(username string) (string, error) {
	claims := jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(24 * time.Hour).Unix(), // 设置过期时间为 24 小时
		"iat":      time.Now().Unix(),                     // 签发时间
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	jwtKey := config.GetConfig().JwtKey
	if jwtKey == "" {
		logger.Warn("jwt key is empty")
		jwtKey = defaultJwtSecret
	}
	return token.SignedString([]byte(jwtKey))
}

// ValidateJWT 验证 JWT 的有效性
func (t *tokenVerify) ValidateJWT(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		jwtSecret := config.GetConfig().JwtKey
		if jwtSecret == "" {
			logger.Warn("jwt key is empty")
			jwtSecret = defaultJwtSecret
		}
		return jwtSecret, nil
	})
	if err != nil {
		return "", err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if username, ok := claims["username"].(string); ok {
			return username, nil
		}
		return "", jwt.ErrInvalidKeyType
	}
	return "", jwt.ErrSignatureInvalid
}

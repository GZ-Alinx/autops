package middleware

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/GZ-Alinx/autops/internal/config"
	"github.com/GZ-Alinx/autops/internal/logger"
	"github.com/GZ-Alinx/autops/internal/response"

	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

// JWTClaims JWT自定义声明
type JWTClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// JWTMiddleware JWT认证中间件
func JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取Authorization头
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			logger.Logger.Warn("JWT认证失败: 未提供认证信息")
			response.Fail(c, http.StatusUnauthorized, errors.New("未提供认证信息"))
			c.Abort()
			return
		}

		// 检查格式
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			logger.Logger.Warn("JWT认证失败: 认证信息格式错误", zap.String("authHeader", authHeader))
			response.Fail(c, http.StatusUnauthorized, errors.New("认证信息格式错误"))
			c.Abort()
			return
		}

		logger.Logger.Info("开始解析JWT令牌")

		// 解析token
		claims := &JWTClaims{}
		token, err := jwt.ParseWithClaims(parts[1], claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(config.AppConfig.JWT.Secret), nil
		})

		// 验证token
		if err != nil {
			logger.Logger.Error("JWT解析失败", zap.Error(err), zap.String("token", parts[1]))
			response.Fail(c, http.StatusUnauthorized, errors.New("无效的token或token已过期"))
			c.Abort()
			return
		}

		if !token.Valid {
			logger.Logger.Warn("JWT令牌无效", zap.String("token", parts[1]))
			response.Fail(c, http.StatusUnauthorized, errors.New("无效的token或token已过期"))
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)

		logger.Logger.Info("JWT认证成功", zap.String("username", claims.Username), zap.String("userID", claims.UserID))

		c.Next()
	}
}

// GenerateToken 生成JWT令牌
func GenerateToken(userID, username string) (string, error) {
	logger.Logger.Info("开始生成JWT令牌", zap.String("username", username), zap.String("userID", userID))

	// 设置过期时间
	expirationTime := time.Now().Add(config.AppConfig.JWT.ExpiresHour * time.Hour)
	logger.Logger.Info("JWT令牌过期时间", zap.Time("expirationTime", expirationTime))

	// 创建声明
	claims := &JWTClaims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    config.AppConfig.App.Name,
		},
	}

	// 创建token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 签名token
	tokenString, err := token.SignedString([]byte(config.AppConfig.JWT.Secret))
	if err != nil {
		logger.Logger.Error("生成JWT令牌失败", zap.String("username", username), zap.Error(err))
		return "", err
	}

	logger.Logger.Info("生成JWT令牌成功", zap.String("username", username))
	return tokenString, nil
}

package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/GZ-Alinx/autops/internal/config"
)

// CorsMiddleware 跨域中间件
func CorsMiddleware() gin.HandlerFunc {
	// 从配置获取允许的源，如果未配置则默认允许所有
	allowOrigins := []string{"*"}
	if len(config.AppConfig.Cors.AllowOrigins) > 0 {
		allowOrigins = config.AppConfig.Cors.AllowOrigins
	}

	return cors.New(cors.Config{
		AllowOrigins:     allowOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: config.AppConfig.Cors.AllowCredentials,
		MaxAge:           time.Duration(config.AppConfig.Cors.MaxAge) * time.Hour,
	})
}
package routes

import (
	"github.com/GZ-Alinx/autops/internal/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册所有路由
func RegisterRoutes(router *gin.Engine) {
	// 全局中间件
	router.Use(middleware.CorsMiddleware())
	router.Use(gin.Recovery())

	// 公开路由
	registerPublicRoutes(router)

	// API路由组
	api := router.Group("/api/v1")
	api.Use(middleware.JWTMiddleware()) // 应用JWT认证中间件
	registerAPIRoutes(api)
}

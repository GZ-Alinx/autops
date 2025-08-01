package routes

import (
	"github.com/GZ-Alinx/autops/business/controllers"
	"github.com/GZ-Alinx/autops/business/repositories"
	"github.com/GZ-Alinx/autops/business/services"
	"github.com/GZ-Alinx/autops/internal/response"
	"github.com/gin-gonic/gin"
)

// registerPublicRoutes 注册公开路由
func registerPublicRoutes(router *gin.Engine) {
	// 健康检查接口
	router.GET("/health", func(c *gin.Context) {
		response.Success(c, "OK")
	})

	// 用户登录路由
	userRepo := repositories.NewUserRepository()
	userService := services.NewUserService(userRepo)
	userController := controllers.NewUserController(userService)
	router.POST("/api/v1/user/login", userController.Login)

}

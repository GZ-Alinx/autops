package routes

import (
	"github.com/GZ-Alinx/autops/business/controllers"
	"github.com/GZ-Alinx/autops/business/repositories"
	"github.com/GZ-Alinx/autops/business/services"
	"github.com/GZ-Alinx/autops/internal/middleware"
	"github.com/gin-gonic/gin"
)

// registerAPIRoutes 注册需要认证的API路由
func registerAPIRoutes(api *gin.RouterGroup) {
	userRepo := repositories.NewUserRepository()
	userService := services.NewUserService(userRepo)
	userController := controllers.NewUserController(userService)

	// 初始化用户控制器
	permController := controllers.NewPermissionController()

	// 用户需要权限检查的接口
	userProtected := api.Group("/users")
	userProtected.Use(middleware.CasbinMiddleware())
	{
		userProtected.POST("/register", userController.Register)
		userProtected.GET("/:id", userController.GetUser)
		userProtected.GET("/", userController.ListUsers)
		userProtected.PUT("/:id", userController.UpdateUser)
		userProtected.PUT("/:id/password", userController.UpdatePassword)
		userProtected.DELETE("/:id", userController.DeleteUser)
	}

	// 权限管理接口
	perm := api.Group("/permissions")
	perm.Use(middleware.CasbinMiddleware())
	{
		perm.POST("/policy", permController.AddPolicy)
		perm.DELETE("/policy", permController.RemovePolicy)
		perm.GET("/policies", permController.GetPolicies)
		perm.PUT("/user-role", permController.UpdateUserRole)
	}

	// 角色管理接口
	role := api.Group("/roles")
	role.Use(middleware.CasbinMiddleware())
	{
		role.POST("/", permController.CreateRole)
		role.GET("/", permController.GetAllRoles)
		role.GET("/:id", permController.GetRoleByID)
		role.PUT("/", permController.UpdateRole)
		role.DELETE("/:id", permController.DeleteRole)
	}
	// 可以根据实际业务需求修改
	example := api.Group("/test")
	{
		example.GET("", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "API数据"})
		})
	}
}

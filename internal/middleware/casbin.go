package middleware

import (
	"fmt"

	"github.com/GZ-Alinx/autops/business/models"
	"github.com/GZ-Alinx/autops/internal/database"
	"github.com/GZ-Alinx/autops/internal/global"
	"github.com/GZ-Alinx/autops/internal/logger"
	"github.com/GZ-Alinx/autops/internal/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// CasbinMiddleware 权限检查中间件
func CasbinMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取当前用户
		var err error
		username, exists := c.Get("username")
		if !exists {
			logger.Logger.Warn("权限检查失败: 未登录")
			response.Unauthorized(c, fmt.Errorf("未登录"))
			c.Abort()
			return
		}

		logger.Logger.Info("开始权限检查", zap.String("username", username.(string)))

		// 查询用户角色
		var user models.User
		result := database.DB.Where("username = ?", username.(string)).First(&user)
		if result.Error != nil {
			logger.Logger.Error("查询用户失败", zap.String("username", username.(string)), zap.Error(result.Error))
			response.Unauthorized(c, fmt.Errorf("用户不存在"))
			c.Abort()
			return
		}

		logger.Logger.Info("查询用户成功", zap.String("username", username.(string)), zap.Int("userID", int(user.ID)))

		// 预加载角色信息
		database.DB.Model(&user).Association("Roles").Find(&user.Roles)
		if user.ID == 0 {
			logger.Logger.Warn("用户不存在", zap.String("username", username.(string)))
			response.Unauthorized(c, fmt.Errorf("用户不存在"))
			c.Abort()
			return
		}

		// 记录用户角色
		var roleNames []string
		for _, role := range user.Roles {
			roleNames = append(roleNames, role.Name)
		}
		logger.Logger.Info("获取用户角色成功", zap.String("username", username.(string)), zap.Strings("roles", roleNames))

		// 获取请求路径和方法
		path := c.Request.URL.Path
		method := c.Request.Method
		logger.Logger.Info("请求信息", zap.String("path", path), zap.String("method", method))

		// 检查权限：遍历用户所有角色
		ok := false
		for _, role := range user.Roles {
			ok, err = global.Enforcer.Enforce(role.Name, path, method)
			if err != nil {
				logger.Logger.Error("权限检查出错", zap.String("role", role.Name), zap.String("path", path), zap.String("method", method), zap.Error(err))
				break
			}
			if ok {
				logger.Logger.Info("权限检查通过", zap.String("role", role.Name), zap.String("path", path), zap.String("method", method))
				break
			}
			logger.Logger.Info("角色权限不足", zap.String("role", role.Name), zap.String("path", path), zap.String("method", method))
		}
		if err != nil {
			logger.Logger.Error("权限检查失败", zap.String("username", username.(string)), zap.Error(err))
			response.Forbidden(c, fmt.Errorf("权限检查失败: %v", err))
			c.Abort()
			return
		}
		if !ok {
			logger.Logger.Warn("没有操作权限", zap.String("username", username.(string)), zap.String("path", path), zap.String("method", method))
			response.Forbidden(c, fmt.Errorf("没有操作权限"))
			c.Abort()
			return
		}

		logger.Logger.Info("权限检查通过", zap.String("username", username.(string)), zap.String("path", path), zap.String("method", method))
		c.Next()
	}
}

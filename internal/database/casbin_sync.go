package database

import (
	"github.com/GZ-Alinx/autops/business/models"
	"github.com/GZ-Alinx/autops/internal/global"
	"github.com/GZ-Alinx/autops/internal/logger"
	"go.uber.org/zap"
)

// SyncCasbinPolicy 同步Casbin策略（包括用户-角色和角色-权限关联）
func SyncCasbinPolicy() error {
	// 1. 清除现有策略
	global.Enforcer.ClearPolicy()

	// 2. 同步用户-角色关联 (g策略)
	var userRoles []models.UserRole
	if err := DB.Preload("User").Preload("Role").Find(&userRoles).Error; err != nil {
		logger.Logger.Error("查询用户角色关联关系失败", zap.Error(err))
		return err
	}

	for _, ur := range userRoles {
		// Casbin g策略格式: g, 用户, 角色
		ok, err := global.Enforcer.AddGroupingPolicy(ur.User.Username, ur.Role.Name)
		if err != nil {
			logger.Logger.Error("添加用户角色关联策略失败", zap.String("user", ur.User.Username), zap.String("role", ur.Role.Name), zap.Error(err))
			return err
		}
		if !ok {
			logger.Logger.Warn("用户角色关联策略已存在", zap.String("user", ur.User.Username), zap.String("role", ur.Role.Name))
		}
	}

	// 3. 同步角色-权限关联 (p策略)
	var rolePermissions []models.RolePermission
	if err := DB.Preload("Role").Preload("Permission").Find(&rolePermissions).Error; err != nil {
		logger.Logger.Error("查询角色权限关联关系失败", zap.Error(err))
		return err
	}

	for _, rp := range rolePermissions {
		// Casbin p策略格式: p, 角色, 资源, 动作
		ok, err := global.Enforcer.AddPolicy(rp.Role.Name, rp.Permission.Resource, rp.Permission.Action)
		if err != nil {
			logger.Logger.Error("添加角色权限策略失败", zap.String("role", rp.Role.Name), zap.String("resource", rp.Permission.Resource), zap.String("action", rp.Permission.Action), zap.Error(err))
			return err
		}
		if !ok {
			logger.Logger.Warn("角色权限策略已存在", zap.String("role", rp.Role.Name), zap.String("resource", rp.Permission.Resource), zap.String("action", rp.Permission.Action))
		}
	}

	// 4. 保存策略
	if err := global.Enforcer.SavePolicy(); err != nil {
		logger.Logger.Error("保存Casbin策略失败", zap.Error(err))
		return err
	}

	logger.Logger.Info("Casbin策略同步成功",
		zap.Int("user_role_count", len(userRoles)),
		zap.Int("role_permission_count", len(rolePermissions)))
	return nil
}

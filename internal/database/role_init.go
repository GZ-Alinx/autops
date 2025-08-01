package database

import (
	"errors"

	"github.com/GZ-Alinx/autops/business/models"
	"github.com/GZ-Alinx/autops/internal/global"
	"github.com/GZ-Alinx/autops/internal/logger"
	"go.uber.org/zap"
)

// InitAdminRole 初始化管理员角色并赋予所有权限
func InitAdminRole() error {
	// 检查admin角色是否已存在
	var adminRole models.Role
	result := DB.Where("name = ?", "admin").First(&adminRole)
	if result.Error == nil {
		// admin角色已存在，更新权限
		logger.Logger.Info("admin角色已存在，更新权限")
		return updateAdminRolePermissions(adminRole.ID)
	}

	// 创建admin角色
	adminRole = models.Role{
		Name:        "admin",
		Description: "系统管理员角色，拥有所有权限",
	}

	if err := DB.Create(&adminRole).Error; err != nil {
		logger.Logger.Error("创建admin角色失败", zap.Error(err))
		return errors.New("创建admin角色失败: " + err.Error())
	}

	// 赋予所有权限
	if err := updateAdminRolePermissions(adminRole.ID); err != nil {
		return err
	}

	logger.Logger.Info("admin角色初始化成功并赋予所有权限")
	return nil
}

// updateAdminRolePermissions 更新admin角色的权限为所有权限
func updateAdminRolePermissions(roleID uint) error {
	// 获取所有权限
	var permissions []models.Permission
	if err := DB.Find(&permissions).Error; err != nil {
		logger.Logger.Error("查询所有权限失败", zap.Error(err))
		return errors.New("查询所有权限失败: " + err.Error())
	}

	// 先删除admin角色的所有现有权限（使用Unscoped永久删除）
	if err := DB.Unscoped().Where("role_id = ?", roleID).Delete(&models.RolePermission{}).Error; err != nil {
		logger.Logger.Error("删除admin角色现有权限失败", zap.Error(err))
		return errors.New("删除admin角色现有权限失败: " + err.Error())
	}

	// 批量添加所有权限
	for _, permission := range permissions {
		rolePermission := models.RolePermission{
			RoleID:       roleID,
			PermissionID: permission.ID,
		}

		if err := DB.Create(&rolePermission).Error; err != nil {
			logger.Logger.Error("添加权限到admin角色失败", zap.Uint("permissionID", permission.ID), zap.Error(err))
			return errors.New("添加权限到admin角色失败: " + err.Error())
		}

		// 同步到Casbin
		if _, err := global.Enforcer.AddPolicy("admin", permission.Resource, permission.Action); err != nil {
			logger.Logger.Error("同步权限到Casbin失败", zap.String("resource", permission.Resource), zap.String("action", permission.Action), zap.Error(err))
		}
	}

	// 保存Casbin策略
	if err := global.Enforcer.SavePolicy(); err != nil {
		logger.Logger.Error("保存Casbin策略失败", zap.Error(err))
		return errors.New("保存Casbin策略失败: " + err.Error())
	}

	return nil
}

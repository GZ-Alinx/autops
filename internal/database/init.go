package database

import (
	"fmt"
	"os"

	"github.com/GZ-Alinx/autops/business/models"
	"github.com/GZ-Alinx/autops/internal/global"
	"github.com/GZ-Alinx/autops/internal/logger"
	"github.com/casbin/casbin/v2"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"go.uber.org/zap"
)

// InitCasbinAndPermissions 初始化Casbin和角色权限
func InitCasbinAndPermissions() error {
	// 自动迁移表结构
	if err := DB.AutoMigrate(
		&models.User{},
		&models.Role{},
		&models.Permission{},
		&models.UserRole{},
		&models.RolePermission{},
	); err != nil {
		return fmt.Errorf("表结构迁移失败: %w", err)
	}

	// 初始化Casbin
	if err := initCasbin(); err != nil {
		return fmt.Errorf("Casbin初始化失败: %w", err)
	}

	// 初始化角色和权限
	if err := InitRolesAndPermissions(); err != nil {
		logger.Logger.Error("初始化角色和权限失败", zap.Error(err))
	}

	return nil
}

// initCasbin 初始化Casbin权限控制
func initCasbin() error {
	// 创建GORM适配器
	adapter, err := gormadapter.NewAdapterByDB(DB)
	if err != nil {
		return fmt.Errorf("创建Casbin适配器失败: %w", err)
	}

	// 从文件加载模型配置
	modelPath := "configs/casbin_model.conf"
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		return fmt.Errorf("Casbin模型配置文件不存在: %s", modelPath)
	}

	// 创建Casbin执行者
	enforcer, err := casbin.NewEnforcer(modelPath, adapter)
	if err != nil {
		return fmt.Errorf("创建Casbin执行者失败: %w", err)
	}

	// 加载策略
	if err := enforcer.LoadPolicy(); err != nil {
		return fmt.Errorf("加载Casbin策略失败: %w", err)
	}

	global.Enforcer = enforcer
	return nil
}

// InitRolesAndPermissions 初始化角色和权限
func InitRolesAndPermissions() error {
	// 1. 创建预设角色
	roles := []models.Role{
		{Name: "admin", Description: "超级管理员"},
		{Name: "user", Description: "普通用户"},
	}

	for _, role := range roles {
		var existingRole models.Role
		if err := DB.Where("name = ?", role.Name).First(&existingRole).Error; err != nil && err.Error() != "record not found" {
			return fmt.Errorf("查询角色失败: %w", err)
		} else if err == nil {
			// 角色已存在，更新描述
			existingRole.Description = role.Description
			if err := DB.Save(&existingRole).Error; err != nil {
				return fmt.Errorf("更新角色失败: %w", err)
			}
		} else {
			// 角色不存在，创建新角色
			if err := DB.Create(&role).Error; err != nil {
				return fmt.Errorf("创建角色失败: %w", err)
			}
		}
	}

	// 2. 创建预设权限
	permissions := []models.Permission{
		{Resource: "/api/v1/users/", Action: "GET", Description: "查看用户列表"},
		{Resource: "/api/v1/users/register", Action: "POST", Description: "创建用户"},
		{Resource: "/api/v1/users/*", Action: "GET", Description: "查看用户详情"},
		{Resource: "/api/v1/users/*", Action: "PUT", Description: "更新用户信息"},
		{Resource: "/api/v1/users/*", Action: "DELETE", Description: "删除用户"},
		{Resource: "/api/v1/users/:id/password", Action: "PUT", Description: "更新用户密码"},
		{Resource: "/api/v1/roles/*", Action: "GET", Description: "查看角色列表"},
		{Resource: "/api/v1/roles/*", Action: "POST", Description: "创建角色"},
		{Resource: "/api/v1/roles/*", Action: "GET", Description: "查看角色详情"},
		{Resource: "/api/v1/roles/*", Action: "PUT", Description: "更新角色信息"},
		{Resource: "/api/v1/roles/*", Action: "DELETE", Description: "删除角色"},
		{Resource: "/api/v1/permissions/policy", Action: "POST", Description: "添加权限策略"},
		{Resource: "/api/v1/permissions/policy", Action: "DELETE", Description: "删除权限策略"},
		{Resource: "/api/v1/permissions/policies", Action: "GET", Description: "获取所有权限策略"},
		{Resource: "/api/v1/permissions/user-role", Action: "PUT", Description: "更新用户角色"},
		{Resource: "/api/v1/test", Action: "GET", Description: "示例接口"},
	}

	for _, permission := range permissions {
		var existingPermission models.Permission
		if err := DB.Where("resource = ? AND action = ?", permission.Resource, permission.Action).First(&existingPermission).Error; err != nil && err.Error() != "record not found" {
			return fmt.Errorf("查询权限失败: %w", err)
		} else if err == nil {
			// 权限已存在，更新描述
			existingPermission.Description = permission.Description
			if err := DB.Save(&existingPermission).Error; err != nil {
				return fmt.Errorf("更新权限失败: %w", err)
			}
		} else {
			// 权限不存在，创建新权限
			if err := DB.Create(&permission).Error; err != nil {
				return fmt.Errorf("创建权限失败: %w", err)
			}
		}
	}

	// 3. 关联角色和权限
	// 获取admin角色
	var adminRole models.Role
	if err := DB.Where("name = ?", "admin").First(&adminRole).Error; err != nil {
		return fmt.Errorf("获取admin角色失败: %w", err)
	}

	// 获取所有权限
	var allPermissions []models.Permission
	if err := DB.Find(&allPermissions).Error; err != nil {
		return fmt.Errorf("获取所有权限失败: %w", err)
	}

	// 为admin角色分配所有权限
	if err := DB.Model(&adminRole).Association("Permissions").Replace(allPermissions); err != nil {
		return fmt.Errorf("为admin角色分配权限失败: %w", err)
	}

	// 获取user角色
	var userRole models.Role
	if err := DB.Where("name = ?", "user").First(&userRole).Error; err != nil {
		return fmt.Errorf("获取user角色失败: %w", err)
	}

	// 为user角色分配部分权限
	var userPermissions []models.Permission
	if err := DB.Where("resource IN (?, ?)", "/api/v1/users", "/api/v1/users/:id").Where("action = ?", "GET").Find(&userPermissions).Error; err != nil {
		return fmt.Errorf("获取用户权限失败: %w", err)
	}

	if err := DB.Model(&userRole).Association("Permissions").Replace(userPermissions); err != nil {
		return fmt.Errorf("为user角色分配权限失败: %w", err)
	}

	// 4. 刷新Casbin策略
	if err := global.Enforcer.LoadPolicy(); err != nil {
		return fmt.Errorf("刷新Casbin策略失败: %w", err)
	}

	logger.Logger.Info("角色和权限初始化成功")
	return nil
}

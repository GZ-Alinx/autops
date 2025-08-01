package database

import (
	"github.com/GZ-Alinx/autops/business/models"
	"github.com/GZ-Alinx/autops/internal/logger"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// InitAdminUser 初始化管理员用户并分配角色
func InitAdminUser() error {
	logger.Logger.Info("开始初始化管理员用户")

	// 检查用户是否已存在
	var user models.User
	result := DB.Where("username = ?", "admin").First(&user)
	if result.Error == nil && user.ID > 0 {
		logger.Logger.Warn("管理员用户已存在")
	} else {
		// 创建admin用户，密码123456
		password := "123456"
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			logger.Logger.Error("密码加密失败", zap.Error(err))
			return err
		}

		phone := ""
		user = models.User{
			Username: "admin",
			Password: string(hashedPassword),
			Email:    "admin@example.com",
			Phone:    &phone,
			Status:   1,
		}

		// 保存用户到数据库
		if err := DB.Create(&user).Error; err != nil {
			logger.Logger.Error("创建管理员用户失败", zap.Error(err))
			return err
		}

		logger.Logger.Info("管理员用户创建成功", zap.Uint("userID", user.ID))
	}

	// 为管理员用户分配admin角色
	var adminRole models.Role
	if err := DB.Where("name = ?", "admin").First(&adminRole).Error; err != nil {
		logger.Logger.Error("获取admin角色失败", zap.Error(err))
		return err
	}

	// 检查用户是否已经拥有admin角色
	var userRole models.UserRole
	result = DB.Where("user_id = ? AND role_id = ?", user.ID, adminRole.ID).First(&userRole)
	if result.Error == nil && userRole.UserID > 0 {
		logger.Logger.Warn("管理员用户已拥有admin角色，跳过分配")
	} else {
		// 建立用户和角色的关联
		if err := DB.Model(&user).Association("Roles").Append(&adminRole); err != nil {
			logger.Logger.Error("为管理员用户分配角色失败", zap.Error(err))
			return err
		}
		logger.Logger.Info("为管理员用户分配角色成功")
	}

	// 同步Casbin策略
	if err := SyncCasbinPolicy(); err != nil {
		logger.Logger.Error("同步Casbin策略失败", zap.Error(err))
		return err
	}

	return nil
}

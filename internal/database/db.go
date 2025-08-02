package database

import (
	"fmt"

	"github.com/GZ-Alinx/autops/business/models"
	"github.com/GZ-Alinx/autops/internal/config"
	"github.com/GZ-Alinx/autops/internal/logger"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	glog "gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDB 初始化数据库连接
func InitDB() error {
	logger.Logger.Info("开始初始化数据库连接")
	mysqlConfig := config.AppConfig.MySQL

	// 构建DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		mysqlConfig.Username,
		mysqlConfig.Password,
		mysqlConfig.Host,
		mysqlConfig.Port,
		mysqlConfig.Database,
		mysqlConfig.Charset,
	)

	// 设置GORM日志模式
	logLevel := glog.Silent
	if config.AppConfig.App.Env == "development" {
		logLevel = glog.Info
	}

	// 连接数据库
	var err error
	// logger.Logger.Info("正在连接数据库: " + dsn)
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: glog.Default.LogMode(logLevel),
	})
	if err != nil {
		return err
	}

	// 获取底层sql.DB并设置连接池
	// logger.Logger.Info("数据库连接成功，正在设置连接池参数")
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}

	// 设置连接池参数
	sqlDB.SetMaxOpenConns(mysqlConfig.MaxOpenConns)
	// logger.Logger.Info(fmt.Sprintf("设置最大打开连接数为: %d", mysqlConfig.MaxOpenConns))
	sqlDB.SetMaxIdleConns(mysqlConfig.MaxIdleConns)
	// logger.Logger.Info(fmt.Sprintf("设置最大空闲连接数为: %d", mysqlConfig.MaxIdleConns))
	sqlDB.SetConnMaxLifetime(mysqlConfig.ConnMaxLife)
	// logger.Logger.Info(fmt.Sprintf("设置连接最大生存时间为: %v", mysqlConfig.ConnMaxLife))

	// 自动迁移数据表
	if err := DB.AutoMigrate(&models.User{}, &models.Role{}, &models.UserRole{}, &models.Permission{}, &models.RolePermission{}); err != nil {
		logger.Logger.Error("数据表迁移失败", zap.Error(err))
		return err
	}

	logger.Logger.Info("数据库连接成功")
	return nil
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return DB
}

// CloseDB 关闭数据库连接
func CloseDB() {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			logger.Logger.Error("获取数据库实例失败", zap.Error(err))
			return
		}
		if err := sqlDB.Close(); err != nil {
			logger.Logger.Error("关闭数据库连接失败", zap.Error(err))
			return
		}
		logger.Logger.Info("数据库连接已关闭")
	}
}

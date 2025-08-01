package logger

import (
	"fmt"

	"github.com/GZ-Alinx/autops/internal/config"
	"go.uber.org/zap"
)

// Logger 全局日志实例
var Logger *zap.Logger

// logManager 全局日志管理器实例
var logManager *LogManagers

// InitLogger 初始化日志模块
func InitLogger(cfg *config.Config) error {
	var err error
	// 初始化日志管理器
	logManager, err = NewLogManager(cfg)
	if err != nil {
		return fmt.Errorf("日志管理器初始化失败: %v", err)
	}

	// 设置全局日志对象为业务日志
	Logger = logManager.BusinessLogger
	zap.ReplaceGlobals(Logger)

	Logger.Info("日志模块初始化成功", zap.String("日志级别", cfg.Logger.Level), zap.String("输出路径", cfg.Logger.Output))
	return nil
}

// Sync 刷新日志缓冲区
func Sync() error {
	if logManager != nil {
		return logManager.Sync()
	}
	return nil
}

// GetBusinessLogger 获取业务日志器
func GetBusinessLogger() *zap.Logger {
	if logManager != nil {
		return logManager.BusinessLogger
	}
	return Logger
}

// GetRouterLogger 获取路由日志器
func GetRouterLogger() *zap.Logger {
	if logManager != nil {
		return logManager.RouterLogger
	}
	return Logger
}

// GetGinLogger 获取Gin框架日志器
func GetGinLogger() *zap.Logger {
	if logManager != nil {
		return logManager.GinLogger
	}
	return Logger
}

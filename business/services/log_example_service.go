package services

import (
	"errors"

	"github.com/GZ-Alinx/autops/internal/logger"
	"go.uber.org/zap"
)

// LogExampleService 日志示例服务
type LogExampleService struct {
}

// NewLogExampleService 创建日志示例服务实例
func NewLogExampleService() *LogExampleService {
	return &LogExampleService{}
}

// DoBusiness 执行业务操作
func (s *LogExampleService) DoBusiness(param string) error {
	// 获取业务日志器
	businessLogger := logger.GetBusinessLogger()
	businessLogger.Info("开始执行业务操作", zap.String("param", param))

	// 模拟业务逻辑
	// ...

	// 记录业务信息
	businessLogger.Debug("业务操作进行中", zap.String("status", "processing"))

	// 模拟潜在错误
	if param == "error" {
		businessLogger.Error("业务操作失败", zap.String("reason", "参数错误"))
		return errors.New("业务错误: 参数错误")
	}

	// 业务完成
	businessLogger.Info("业务操作完成", zap.String("result", "success"))
	return nil
}

// LogRouteInfo 记录路由信息
func (s *LogExampleService) LogRouteInfo(path string, method string) {
	// 获取路由日志器
	routerLogger := logger.GetRouterLogger()
	routerLogger.Info("路由访问记录", zap.String("path", path), zap.String("method", method))
}

// LogGinInfo 记录Gin框架信息
func (s *LogExampleService) LogGinInfo(message string) {
	// 获取Gin日志器
	ginLogger := logger.GetGinLogger()
	ginLogger.Info("Gin框架信息", zap.String("message", message))
}
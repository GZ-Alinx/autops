# 日志配置使用文档

## 日志配置系统概述

本项目采用 `zap` 日志库结合 `lumberjack` 实现了一个功能完善的日志配置系统。该系统支持：

- 多级别日志（debug、info、warn、error、fatal）
- 多种日志格式（JSON、普通文本）
- 多种输出方式（控制台、文件、同时输出到两者）
- 日志文件轮转（基于大小、时间）
- 分类日志（业务日志、路由日志、Gin框架日志）

## 日志配置

### 配置文件

日志配置位于 `config.yaml` 文件中的 `logger` 部分：

```yaml
logger:
  level: info         # 日志级别: debug, info, warn, error, fatal
  format: json        # 日志格式: json, plain
  output: logs/       # 日志输出目录
  max_size: 100       # 单个日志文件最大大小(MB)
  max_backups: 10     # 保留的最大备份文件数
  max_age: 7          # 日志文件最大保留天数
  compress: true      # 是否压缩旧日志文件
```

### 日志级别

- `debug`: 详细的调试信息
- `info`: 一般信息
- `warn`: 警告信息，不影响程序运行
- `error`: 错误信息，需要关注
- `fatal`: 致命错误，程序通常会退出

### 日志格式

- `json`: JSON格式日志，包含时间戳、日志级别、调用位置等信息
- `plain`: 普通文本格式，易于阅读

### 日志输出

- `console`: 只输出到控制台
- `file`: 只输出到文件
- `both`: 同时输出到控制台和文件

## 代码中使用日志

### 获取日志器

系统提供了三种类型的日志器：

```go
// 获取业务日志器
businessLogger := logger.GetBusinessLogger()

// 获取路由日志器
routerLogger := logger.GetRouterLogger()

// 获取Gin框架日志器
ginLogger := logger.GetGinLogger()

// 全局通用日志器
logger.Logger
```

### 记录日志

```go
// 记录信息日志
businessLogger.Info("业务操作开始", zap.String("param", value))

// 记录调试日志
businessLogger.Debug("变量值", zap.Int("id", id), zap.String("name", name))

// 记录警告日志
businessLogger.Warn("数据不完整", zap.String("field", "email"))

// 记录错误日志
businessLogger.Error("业务操作失败", zap.Error(err))

// 记录致命错误
logger.Logger.Fatal("数据库连接失败", zap.Error(err))
```

### 日志字段

可以通过 `zap` 提供的字段函数添加自定义字段：

- `zap.String(key, value)`: 字符串字段
- `zap.Int(key, value)`: 整数字段
- `zap.Bool(key, value)`: 布尔字段
- `zap.Error(err)`: 错误字段
- `zap.Time(key, time.Now())`: 时间字段
- 更多字段类型请参考 [zap 文档](https://pkg.go.dev/go.uber.org/zap)

## 示例代码

```go
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
```

## 中间件日志

系统已实现以下日志中间件：

1. `RequestIDMiddleware`: 生成请求ID
2. `AccessLoggerMiddleware`: 记录请求访问日志
3. `ErrorLoggerMiddleware`: 记录请求错误日志
4. `GinLoggerToZap`: 将Gin默认日志重定向到Zap

这些中间件已在 `main.go` 中自动注册，无需手动添加。

## 日志文件

日志文件默认保存在 `logs/` 目录下：

- `app.log`: 主日志文件
- `business.log`: 业务日志文件
- `router.log`: 路由日志文件
- `gin.log`: Gin框架日志文件

日志文件会根据配置自动轮转，避免单个文件过大。

## 最佳实践

1. 根据日志重要性选择合适的日志级别
2. 为日志添加足够的上下文信息
3. 错误日志务必包含错误对象
4. 避免在生产环境启用debug级别日志
5. 定期清理旧日志文件

希望本文档能帮助您更好地使用和配置项目中的日志系统。
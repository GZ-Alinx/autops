package middleware

import (
	"context"
	"time"

	"github.com/GZ-Alinx/autops/internal/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"github.com/google/uuid"
)

// RequestIDMiddleware 生成请求ID中间件
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 生成请求ID
		requestID := uuid.New().String()
		c.Set("requestID", requestID)
		// 添加到响应头
		c.Writer.Header().Set("X-Request-ID", requestID)
		c.Next()
	}
}

// AccessLoggerMiddleware 访问日志中间件
func AccessLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 开始时间
		startTime := time.Now()
		// 请求ID
		requestID, _ := c.Get("requestID")
		// 客户端IP
		clientIP := c.ClientIP()
		// 请求方法
		method := c.Request.Method
		// 请求路径
		path := c.Request.URL.Path
		// 请求参数
		params := c.Request.URL.RawQuery
		// 请求头
		headers := make(map[string]string)
		for name, values := range c.Request.Header {
			headers[name] = values[0]
		}

		// 处理请求
		c.Next()

		// 结束时间
		endTime := time.Now()
		// 响应时间
		duration := endTime.Sub(startTime)
		// 响应状态码
		statusCode := c.Writer.Status()
		// 响应大小
		responseSize := c.Writer.Size()

		// 记录路由日志
		routerLogger := logger.GetRouterLogger()
		routerLogger.Info("请求访问日志",
			zap.String("requestID", requestID.(string)),
			zap.String("clientIP", clientIP),
			zap.String("method", method),
			zap.String("path", path),
			zap.String("params", params),
			zap.Int("statusCode", statusCode),
			zap.Int64("responseSize", int64(responseSize)),
			zap.Duration("duration", duration),
		)
	}
}

// ErrorLoggerMiddleware 错误日志中间件
func ErrorLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 检查是否有错误
		if len(c.Errors) > 0 {
			// 请求ID
			requestID, _ := c.Get("requestID")
			// 客户端IP
			clientIP := c.ClientIP()
			// 请求方法
			method := c.Request.Method
			// 请求路径
			path := c.Request.URL.Path

			// 获取业务日志器
			businessLogger := logger.GetBusinessLogger()

			for _, err := range c.Errors {
				businessLogger.Error("请求错误日志",
					zap.String("requestID", requestID.(string)),
					zap.String("clientIP", clientIP),
					zap.String("method", method),
					zap.String("path", path),
					zap.Error(err.Err),
				)
			}
		}
	}
}

// GinLoggerToZap 将Gin默认日志重定向到Zap
func GinLoggerToZap() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 替换Gin默认日志为Gin专用日志器
		ginLogger := logger.GetGinLogger()
		c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), "gin-logger", ginLogger))
		c.Next()
	}
}
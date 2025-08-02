package main

// @title Autops API
// @version 1.0
// @description 后台管理API接口文档
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @BasePath /api/v1
// @host localhost:8080

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"go.uber.org/zap"

	"github.com/GZ-Alinx/autops/internal/config"
	"github.com/GZ-Alinx/autops/internal/database"
	"github.com/GZ-Alinx/autops/internal/logger"
	"github.com/GZ-Alinx/autops/internal/middleware"

	"github.com/GZ-Alinx/autops/business/routes"
)

func main() {
	// 加载配置
	if err := config.LoadConfig(); err != nil {
		panic(fmt.Sprintf("加载配置失败: %v", err))
	}

	// 初始化日志
	if err := logger.InitLogger(&config.AppConfig); err != nil {
		panic(fmt.Sprintf("日志初始化失败: %v", err))
	}
	defer logger.Logger.Sync()
	logger.Logger.Info("日志系统初始化成功")

	// 初始化数据库
	if err := database.InitDB(); err != nil {
		logger.Logger.Fatal("数据库初始化失败", zap.Error(err))
	}
	defer database.CloseDB()

	// 初始化Casbin和角色权限
	if err := database.InitCasbinAndPermissions(); err != nil {
		logger.Logger.Fatal("角色权限初始化失败", zap.Error(err))
	}

	// 初始化管理员用户
	if err := database.InitAdminUser(); err != nil {
		logger.Logger.Error("初始化管理员用户失败", zap.Error(err))
	}

	logger.Logger.Info("权限控制和角色权限初始化成功")

	// 设置Gin模式
	if config.AppConfig.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
		// 禁用Gin默认控制台日志
		gin.DisableConsoleColor()
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// 配置Gin日志输出
	gin.DefaultWriter = os.Stdout
	gin.DefaultErrorWriter = os.Stderr

	// 创建Gin引擎
	router := gin.New()

	// 添加日志中间件
	router.Use(middleware.RequestIDMiddleware())
	router.Use(middleware.AccessLoggerMiddleware())
	router.Use(middleware.ErrorLoggerMiddleware())
	router.Use(middleware.GinLoggerToZap())

	// 路由注册
	// 提供静态文件访问
	router.Static("/docs", "/Users/alinx/code/autops/docs")
	// API文档
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("/docs/swagger.json")))

	routes.RegisterRoutes(router)

	// 创建HTTP服务器
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.AppConfig.App.Port),
		Handler: router,
	}

	// 启动服务器（非阻塞）
	go func() {
		logger.Logger.Info(fmt.Sprintf("服务器启动成功，监听端口: %d", config.AppConfig.App.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Logger.Fatal("服务器启动失败", zap.Error(err))
		}
	}()

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Logger.Info("正在关闭服务器...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Logger.Fatal("服务器关闭失败", zap.Error(err))
	}

	logger.Logger.Info("服务器已关闭")
}

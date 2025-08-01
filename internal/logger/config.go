package logger

import (
	"os"
	"path/filepath"

	"github.com/GZ-Alinx/autops/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// LogLevel 定义日志级别类型
type LogLevel string

// 日志级别常量
const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
	LogLevelFatal LogLevel = "fatal"
)

// LogFormat 定义日志格式类型
type LogFormat string

// 日志格式常量
const (
	LogFormatJSON  LogFormat = "json"
	LogFormatPlain LogFormat = "plain"
)

// LogOutput 定义日志输出类型
type LogOutput string

// 日志输出常量
const (
	LogOutputConsole LogOutput = "console"
	LogOutputFile    LogOutput = "file"
	LogOutputBoth    LogOutput = "both"
)

// BusinessLogConfig 业务日志配置
type BusinessLogConfig struct {
	Level      LogLevel  `json:"level"`
	Format     LogFormat `json:"format"`
	Output     LogOutput `json:"output"`
	File       string    `json:"file"`
	MaxSize    int       `json:"max_size"`    // MB
	MaxBackups int       `json:"max_backups"` // 个
	MaxAge     int       `json:"max_age"`     // 天
	Compress   bool      `json:"compress"`    // 是否压缩
}

// RouterLogConfig 路由日志配置
type RouterLogConfig struct {
	Level      LogLevel  `json:"level"`
	Format     LogFormat `json:"format"`
	Output     LogOutput `json:"output"`
	File       string    `json:"file"`
	MaxSize    int       `json:"max_size"`    // MB
	MaxBackups int       `json:"max_backups"` // 个
	MaxAge     int       `json:"max_age"`     // 天
	Compress   bool      `json:"compress"`    // 是否压缩
}

// GinLogConfig Gin框架日志配置
type GinLogConfig struct {
	Level      LogLevel  `json:"level"`
	Format     LogFormat `json:"format"`
	Output     LogOutput `json:"output"`
	File       string    `json:"file"`
	MaxSize    int       `json:"max_size"`    // MB
	MaxBackups int       `json:"max_backups"` // 个
	MaxAge     int       `json:"max_age"`     // 天
	Compress   bool      `json:"compress"`    // 是否压缩
}

// LogManager 日志管理器
type LogManagers struct {
	BusinessLogger *zap.Logger
	RouterLogger   *zap.Logger
	GinLogger      *zap.Logger
}

// NewLogManager 创建日志管理器
func NewLogManager(appConfig *config.Config) (*LogManagers, error) {
	// 创建业务日志
	businessLogger, err := createLogger(
		appConfig.Logger.Level,
		appConfig.Logger.Format,
		appConfig.Logger.Output,
		appConfig.Logger.MaxSize,
		appConfig.Logger.MaxBackups,
		appConfig.Logger.MaxAge,
		appConfig.Logger.Compress,
		"business.log",
	)
	if err != nil {
		return nil, err
	}

	// 创建路由日志
	routerLogger, err := createLogger(
		appConfig.Logger.Level,
		appConfig.Logger.Format,
		appConfig.Logger.Output,
		appConfig.Logger.MaxSize,
		appConfig.Logger.MaxBackups,
		appConfig.Logger.MaxAge,
		appConfig.Logger.Compress,
		"router.log",
	)
	if err != nil {
		return nil, err
	}

	// 创建Gin框架日志
	ginLogger, err := createLogger(
		appConfig.Logger.Level,
		appConfig.Logger.Format,
		appConfig.Logger.Output,
		appConfig.Logger.MaxSize,
		appConfig.Logger.MaxBackups,
		appConfig.Logger.MaxAge,
		appConfig.Logger.Compress,
		"gin.log",
	)
	if err != nil {
		return nil, err
	}

	return &LogManagers{
		BusinessLogger: businessLogger,
		RouterLogger:   routerLogger,
		GinLogger:      ginLogger,
	}, nil
}

// createLogger 创建日志实例
func createLogger(
	levelStr string,
	formatStr string,
	outputStr string,
	maxSize int,
	maxBackups int,
	maxAge int,
	compress bool,
	fileName string,
) (*zap.Logger, error) {
	// 设置日志级别
	level, err := zapcore.ParseLevel(levelStr)
	if err != nil {
		return nil, err
	}

	// 设置日志编码器
	var encoder zapcore.Encoder
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	if formatStr == "plain" {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	// 设置日志输出
	var writeSyncer zapcore.WriteSyncer
	if outputStr == "console" {
		writeSyncer = zapcore.AddSync(os.Stdout)
	} else if outputStr == "file" {
		// 确保日志目录存在
		if err := os.MkdirAll(config.AppConfig.Logger.Output, 0755); err != nil {
			return nil, err
		}

		// 创建日志文件
		writer := &lumberjack.Logger{
			Filename:   filepath.Join(config.AppConfig.Logger.Output, fileName),
			MaxSize:    maxSize,
			MaxBackups: maxBackups,
			MaxAge:     maxAge,
			Compress:   compress,
		}
		writeSyncer = zapcore.AddSync(writer)
	} else {
		// 同时输出到控制台和文件
		// 确保日志目录存在
		if err := os.MkdirAll(config.AppConfig.Logger.Output, 0755); err != nil {
			return nil, err
		}

		// 创建日志文件
		writer := &lumberjack.Logger{
			Filename:   filepath.Join(config.AppConfig.Logger.Output, fileName),
			MaxSize:    maxSize,
			MaxBackups: maxBackups,
			MaxAge:     maxAge,
			Compress:   compress,
		}
		writeSyncer = zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(writer))
	}

	// 创建核心日志配置
	core := zapcore.NewCore(encoder, writeSyncer, level)

	// 创建日志实例
	logger := zap.New(core, zap.AddCaller())

	return logger, nil
}

// Sync 刷新所有日志缓冲区
func (lm *LogManagers) Sync() error {
	if err := lm.BusinessLogger.Sync(); err != nil {
		return err
	}
	if err := lm.RouterLogger.Sync(); err != nil {
		return err
	}
	if err := lm.GinLogger.Sync(); err != nil {
		return err
	}
	return nil
}

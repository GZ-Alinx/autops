package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// AppConfig 应用配置
type AppConfigs struct {
	Name    string        `mapstructure:"name"`
	Env     string        `mapstructure:"env"`
	Port    int           `mapstructure:"port"`
	Timeout time.Duration `mapstructure:"timeout"`
}

// LoggerConfig 日志配置
type LoggerConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	Output     string `mapstructure:"output"`
	MaxSize    int    `mapstructure:"max_size"`    // MB
	MaxBackups int    `mapstructure:"max_backups"` // 个
	MaxAge     int    `mapstructure:"max_age"`     // 天
	Compress   bool   `mapstructure:"compress"`   // 是否压缩
}

// MySQLConfig MySQL数据库配置
type MySQLConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	Username     string        `mapstructure:"username"`
	Password     string        `mapstructure:"password"`
	Database     string        `mapstructure:"database"`
	Charset      string        `mapstructure:"charset"`
	MaxOpenConns int           `mapstructure:"max_open_conns"`
	MaxIdleConns int           `mapstructure:"max_idle_conns"`
	ConnMaxLife  time.Duration `mapstructure:"conn_max_lifetime"`
}

// CorsConfig CORS配置
type CorsConfig struct {
	AllowOrigins     []string `mapstructure:"allow_origins"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
	MaxAge           int      `mapstructure:"max_age"` // 小时
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret      string        `mapstructure:"secret"`
	ExpiresHour time.Duration `mapstructure:"expires_hours"`
}

// Config 应用总配置
type Config struct {
	App    AppConfigs   `mapstructure:"app"`
	Logger LoggerConfig `mapstructure:"logger"`
	MySQL  MySQLConfig  `mapstructure:"mysql"`
	JWT    JWTConfig    `mapstructure:"jwt"`
	Cors   CorsConfig   `mapstructure:"cors"`
}

// AppConfig 全局配置实例
var AppConfig Config

// LoadConfig 加载配置文件
func LoadConfig() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()
	fmt.Println("正在读取配置文件: config.yaml")

	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	if err := viper.Unmarshal(&AppConfig); err != nil {
		return err
	}

	fmt.Println("配置加载成功，应用配置已初始化")
	return nil
}

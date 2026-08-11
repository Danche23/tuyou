package config

import (
	"time"
)

// Config 应用配置结构体
type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Database DatabaseConfig `mapstructure:"database"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Log      LogConfig      `mapstructure:"log"`
	CORS     CORSConfig     `mapstructure:"cors"`
	SMS      SMSConfig      `mapstructure:"sms"`
	Storage  StorageConfig  `mapstructure:"storage"`
}

// AppConfig 应用配置
type AppConfig struct {
	Name    string `mapstructure:"name"`
	Version string `mapstructure:"version"`
	Mode    string `mapstructure:"mode"` // debug, release, test
	Port    int    `mapstructure:"port"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	MySQL MySQLConfig `mapstructure:"mysql"`
	Redis RedisConfig `mapstructure:"redis"`
}

// MySQLConfig MySQL 配置
type MySQLConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Username     string `mapstructure:"username"`
	Password     string `mapstructure:"password"`
	Database     string `mapstructure:"database"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
}

// RedisConfig Redis 配置
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

// JWTConfig JWT 配置
type JWTConfig struct {
	Secret      string        `mapstructure:"secret"`
	ExpireHours time.Duration `mapstructure:"expire_hours"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string `mapstructure:"level"`       // debug, info, warn, error
	Filename   string `mapstructure:"filename"`    // 日志文件路径
	MaxSize    int    `mapstructure:"max_size"`    // 单个日志文件最大大小(MB)
	MaxBackups int    `mapstructure:"max_backups"` // 保留的旧日志文件数量
	MaxAge     int    `mapstructure:"max_age"`     // 保留旧日志文件的最大天数
	Compress   bool   `mapstructure:"compress"`    // 是否压缩
}

// CORSConfig CORS 配置
type CORSConfig struct {
	Enabled          bool     `mapstructure:"enabled"`
	AllowOrigins     []string `mapstructure:"allow_origins"`
	AllowMethods     []string `mapstructure:"allow_methods"`
	AllowHeaders     []string `mapstructure:"allow_headers"`
	ExposeHeaders    []string `mapstructure:"expose_headers"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
	MaxAge           int      `mapstructure:"max_age"`
}

// SMSConfig 短信配置
type SMSConfig struct {
	Driver           string `mapstructure:"driver"`             // 短信驱动： log | aliyun
	AccessKeyId      string `mapstructure:"access_key_id"`      // 阿里云 AccessKeyId
	AccessKeySecret  string `mapstructure:"access_key_secret"`  // 阿里云 AccessKeySecret
	SignName         string `mapstructure:"sign_name"`          // 短信签名
	TemplateCode     string `mapstructure:"template_code"`      // 模板编码
	TemplateParam    string `mapstructure:"template_param"`     // 模板参数 JSON（如 {"code":"##code##"}）
	RegionId         string `mapstructure:"region_id"`          // 地域 ID
	SchemeName       string `mapstructure:"scheme_name"`        // 方案名称
	CountryCode      string `mapstructure:"country_code"`       // 国家码，默认 "86"
	CodeLength       int64  `mapstructure:"code_length"`        // 验证码长度
	ValidTime        int64  `mapstructure:"valid_time"`         // 验证码有效期（分钟）
	DuplicatePolicy  int64  `mapstructure:"duplicate_policy"`   // 同号策略：0-不限制 1-拒绝
	Interval         int64  `mapstructure:"interval"`           // 发送间隔（秒）
	CodeType         int64  `mapstructure:"code_type"`          // 验证码类型
	ReturnVerifyCode bool   `mapstructure:"return_verify_code"` // 是否返回验证码
	AutoRetry        int64  `mapstructure:"auto_retry"`         // 自动重试
	CaseAuthPolicy   int64  `mapstructure:"case_auth_policy"`   // 认证策略
}

// StorageConfig 文件存储配置
type StorageConfig struct {
	Driver string           `mapstructure:"driver"` // 存储驱动：oss（local 预留）
	OSS    OSSStorageConfig `mapstructure:"oss"`
}

// OSSStorageConfig 阿里云 OSS 配置
type OSSStorageConfig struct {
	AccessKeyID     string `mapstructure:"access_key_id"`
	AccessKeySecret string `mapstructure:"access_key_secret"`
	Endpoint        string `mapstructure:"endpoint"` // OSS 区域名（如 cn-beijing），SDK v2 WithRegion 直接吃区域名
	Bucket          string `mapstructure:"bucket"`
}

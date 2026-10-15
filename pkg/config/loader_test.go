package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// envKeys 是 loader 中支持覆盖的敏感环境变量名。
// 注意：这些名字故意不带 TUYOU_ 前缀——带前缀的是 viper 的 AutomaticEnv 机制，
// 不带前缀的由 loader.go 里显式的 os.Getenv 处理。两种机制并存。
var envKeys = []string{
	"MYSQL_PASSWORD",
	"REDIS_PASSWORD",
	"JWT_SECRET",
	"SMS_ACCESS_KEY_ID",
	"SMS_ACCESS_KEY_SECRET",
	"OSS_ACCESS_KEY_ID",
	"OSS_ACCESS_KEY_SECRET",
}

// writeTempConfig 在临时目录写一份最小可用配置，返回其路径。
// 所有敏感字段统一填 "from-file"，便于区分「来自文件」还是「来自环境变量」。
func writeTempConfig(t *testing.T) string {
	t.Helper()

	content := `
app:
  name: tuyou-api
  version: 1.0.0
  mode: release
  port: 9300

database:
  mysql:
    host: 127.0.0.1
    port: 3306
    username: root
    password: "from-file"
    database: tuyou
    max_idle_conns: 10
    max_open_conns: 100
  redis:
    host: 127.0.0.1
    port: 6379
    password: "from-file"
    db: 0
    pool_size: 10

jwt:
  secret: "from-file"
  expire_hours: 24h

log:
  level: info
  filename: logs/app.log

sms:
  driver: aliyun
  access_key_id: "from-file"
  access_key_secret: "from-file"
  sign_name: "test"
  template_code: "100001"

storage:
  driver: oss
  oss:
    access_key_id: "from-file"
    access_key_secret: "from-file"
    endpoint: cn-beijing
    bucket: tuyou-test
`

	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("写入临时配置失败: %v", err)
	}
	return path
}

// TestLoad_EnvOverrideSensitiveFields 断言敏感字段可以被环境变量覆盖。
//
// 这是生产部署的关键能力：密钥只存在于部署机的环境变量里，
// 仓库中的配置文件（连同 config.example.yaml）只保留占位符。
func TestLoad_EnvOverrideSensitiveFields(t *testing.T) {
	path := writeTempConfig(t)

	want := map[string]string{
		"MYSQL_PASSWORD":        "env-mysql-pwd",
		"REDIS_PASSWORD":        "env-redis-pwd",
		"JWT_SECRET":            "env-jwt-secret",
		"SMS_ACCESS_KEY_ID":     "env-sms-id",
		"SMS_ACCESS_KEY_SECRET": "env-sms-secret",
		"OSS_ACCESS_KEY_ID":     "env-oss-id",
		"OSS_ACCESS_KEY_SECRET": "env-oss-secret",
	}
	for k, v := range want {
		t.Setenv(k, v)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	cases := []struct {
		field string
		got   string
	}{
		{"MySQL.Password", cfg.Database.MySQL.Password},
		{"Redis.Password", cfg.Database.Redis.Password},
		{"JWT.Secret", cfg.JWT.Secret},
		{"SMS.AccessKeyId", cfg.SMS.AccessKeyId},
		{"SMS.AccessKeySecret", cfg.SMS.AccessKeySecret},
		{"Storage.OSS.AccessKeyID", cfg.Storage.OSS.AccessKeyID},
		{"Storage.OSS.AccessKeySecret", cfg.Storage.OSS.AccessKeySecret},
	}

	for _, c := range cases {
		if c.got == "from-file" {
			t.Errorf("字段 %s 未被环境变量覆盖，仍为配置文件里的值", c.field)
		}
	}

	// 逐项校验真的取到了对应环境变量的值，而不只是「变了」
	if cfg.Database.MySQL.Password != want["MYSQL_PASSWORD"] {
		t.Errorf("MySQL.Password = %q, want %q", cfg.Database.MySQL.Password, want["MYSQL_PASSWORD"])
	}
	if cfg.SMS.AccessKeyId != want["SMS_ACCESS_KEY_ID"] {
		t.Errorf("SMS.AccessKeyId = %q, want %q", cfg.SMS.AccessKeyId, want["SMS_ACCESS_KEY_ID"])
	}
	if cfg.SMS.AccessKeySecret != want["SMS_ACCESS_KEY_SECRET"] {
		t.Errorf("SMS.AccessKeySecret = %q, want %q", cfg.SMS.AccessKeySecret, want["SMS_ACCESS_KEY_SECRET"])
	}
	if cfg.Storage.OSS.AccessKeyID != want["OSS_ACCESS_KEY_ID"] {
		t.Errorf("Storage.OSS.AccessKeyID = %q, want %q", cfg.Storage.OSS.AccessKeyID, want["OSS_ACCESS_KEY_ID"])
	}
	if cfg.Storage.OSS.AccessKeySecret != want["OSS_ACCESS_KEY_SECRET"] {
		t.Errorf("Storage.OSS.AccessKeySecret = %q, want %q", cfg.Storage.OSS.AccessKeySecret, want["OSS_ACCESS_KEY_SECRET"])
	}
}

// TestLoad_NoEnvKeepsFileValues 断言「没设环境变量时，配置文件的值必须原样保留」。
// 这条守卫的是本地开发行为不变——覆盖机制不能反过来把文件里的值清掉。
func TestLoad_NoEnvKeepsFileValues(t *testing.T) {
	path := writeTempConfig(t)

	// 显式置空，避免开发机上恰好存在同名环境变量导致测试不稳定
	for _, k := range envKeys {
		t.Setenv(k, "")
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	got := map[string]string{
		"MySQL.Password":              cfg.Database.MySQL.Password,
		"Redis.Password":              cfg.Database.Redis.Password,
		"JWT.Secret":                  cfg.JWT.Secret,
		"SMS.AccessKeyId":             cfg.SMS.AccessKeyId,
		"SMS.AccessKeySecret":         cfg.SMS.AccessKeySecret,
		"Storage.OSS.AccessKeyID":     cfg.Storage.OSS.AccessKeyID,
		"Storage.OSS.AccessKeySecret": cfg.Storage.OSS.AccessKeySecret,
	}
	for field, v := range got {
		if v != "from-file" {
			t.Errorf("未设置环境变量时 %s 应保留文件原值，实际为 %q", field, v)
		}
	}
}

// TestExampleConfig_LoadsAndContainsNoSecrets 守卫仓库里的 config.example.yaml。
//
// 两道防线：
//  1. 样例必须能被真实解析——否则别人 clone 后照抄会直接启动失败（样例就是文档）；
//  2. 样例里绝不能出现真实凭据——防止有人图省事，
//     直接把自己的 config.yaml 覆盖成 config.example.yaml 提交上去。
//
// 第 2 条是这个仓库最容易踩、后果最重的一种失误，所以用测试钉死。
func TestExampleConfig_LoadsAndContainsNoSecrets(t *testing.T) {
	for _, k := range envKeys {
		t.Setenv(k, "")
	}

	path := filepath.Join("..", "..", "configs", "config.example.yaml")

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("样例配置无法解析，说明样例本身不可用: %v", err)
	}
	if cfg.App.Name == "" {
		t.Error("样例配置缺少 app.name")
	}
	if cfg.App.Port != 9300 {
		t.Errorf("样例配置 app.port = %d，期望 9300", cfg.App.Port)
	}
	if cfg.Database.MySQL.Database == "" {
		t.Error("样例配置缺少 database.mysql.database")
	}
	if cfg.Storage.OSS.Endpoint == "" {
		t.Error("样例配置缺少 storage.oss.endpoint")
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("读取样例配置失败: %v", err)
	}
	text := string(raw)

	// 阿里云 AccessKey ID 固定以 LTAI 开头，出现即代表真实凭据被写进来了
	if strings.Contains(text, "LTAI") {
		t.Error("样例配置中检出真实 AccessKey（LTAI 前缀），请改为占位符")
	}
	if strings.Contains(text, "123456") {
		t.Error("样例配置中检出弱默认密码 123456，请改为占位符")
	}
	if strings.Contains(text, "change-in-production") && !strings.Contains(text, "change-me") {
		t.Error("样例配置疑似残留默认密钥占位")
	}
}

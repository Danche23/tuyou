package utils

import (
	"time"
	_ "time/tzdata" // 内置时区数据，兼容未安装 tzdata 的运行环境。
)

// BeijingLocation 是项目统一使用的北京时间时区。
var BeijingLocation = func() *time.Location {
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		panic(err)
	}
	return location
}()

// InitTimezone 在应用启动时调用，统一业务时间、日志和按日任务的本地时区。
func InitTimezone() {
	time.Local = BeijingLocation
}

// GetCurrentTimestamp 获取当前时间戳（秒）
func GetCurrentTimestamp() int64 {
	return time.Now().Unix()
}

// GetCurrentMilliTimestamp 获取当前时间戳（毫秒）
func GetCurrentMilliTimestamp() int64 {
	return time.Now().UnixMilli()
}

// FormatTime 格式化时间
func FormatTime(t time.Time) string {
	return t.In(BeijingLocation).Format("2006-01-02 15:04:05")
}

// ParseTime 解析时间字符串
func ParseTime(s string) (time.Time, error) {
	return time.ParseInLocation("2006-01-02 15:04:05", s, BeijingLocation)
}

package sms

import (
	"context"
	"fmt"
	"time"

	"tuyou/pkg/config"
	"tuyou/pkg/database"
)

// sceneKeyPrefix 场景锁 key 前缀
const sceneKeyPrefix = "sms:scene:"

// sceneLockTTL 场景锁有效期 = 验证码有效期 + 60 秒缓冲。
// 避免出现"阿里云的验证码仍有效、但场景锁已过期"的窗口。
func sceneLockTTL() time.Duration {
	validTime := config.Get().SMS.ValidTime // 分钟
	return time.Duration(validTime)*time.Minute + 60*time.Second
}

// sceneKey 生成场景锁 key
func sceneKey(phone string) string {
	return sceneKeyPrefix + phone
}

// storeScene 验证码发送成功后写入场景锁
func storeScene(phone string, scene Scene) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	key := sceneKey(phone)
	if err := database.GetRedis().Set(ctx, key, string(scene), sceneLockTTL()).Err(); err != nil {
		return fmt.Errorf("写入短信验证码场景锁失败: %w", err)
	}
	return nil
}

// getScene 读取场景锁
func getScene(phone string) (Scene, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	val, err := database.GetRedis().Get(ctx, sceneKey(phone)).Result()
	if err != nil {
		return "", err // 由调用方决定如何处理（fail closed）
	}
	return Scene(val), nil
}

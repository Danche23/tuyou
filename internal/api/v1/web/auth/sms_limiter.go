package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
	appErrors "tuyou/pkg/errors"
)

const smsLimitScript = `
if redis.call('EXISTS', KEYS[1]) == 1 then return 1 end
if tonumber(redis.call('GET', KEYS[2]) or '0') >= tonumber(ARGV[1]) then return 2 end
if tonumber(redis.call('GET', KEYS[3]) or '0') >= tonumber(ARGV[2]) then return 3 end
redis.call('SET', KEYS[1], '1', 'EX', ARGV[3])
if redis.call('INCR', KEYS[2]) == 1 then redis.call('EXPIRE', KEYS[2], ARGV[4]) end
if redis.call('INCR', KEYS[3]) == 1 then redis.call('EXPIRE', KEYS[3], ARGV[4]) end
return 0
`

var smsLimit = redis.NewScript(smsLimitScript)

func reserveSMS(ctx context.Context, client *redis.Client, phone string) error {
	if client == nil {
		return errors.New("短信限流不可用")
	}

	hash := sha256.Sum256([]byte(phone))
	phoneKey := hex.EncodeToString(hash[:])
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	result, err := smsLimit.Run(ctx, client, []string{
		"sms:cooldown:" + phoneKey,
		"sms:daily:phone:" + phoneKey,
		"sms:daily:global",
	}, 5, 100, 60, 86400).Int()
	if err != nil {
		return err
	}
	switch result {
	case 0:
		return nil
	case 1:
		return appErrors.ErrSmsSendFrequently
	case 2, 3:
		return appErrors.ErrSmsDailyLimit
	default:
		return errors.New("短信限流返回未知结果")
	}
}

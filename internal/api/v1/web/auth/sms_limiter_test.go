package auth

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	appErrors "tuyou/pkg/errors"
)

func TestReserveSMSPhoneLimits(t *testing.T) {
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()
	phone := "13800138000"

	for i := 0; i < 5; i++ {
		if err := reserveSMS(ctx, client, phone); err != nil {
			t.Fatalf("reservation %d: %v", i+1, err)
		}
		if err := reserveSMS(ctx, client, phone); err != appErrors.ErrSmsSendFrequently {
			t.Fatalf("cooldown: got %v", err)
		}
		server.FastForward(time.Minute)
	}
	if err := reserveSMS(ctx, client, phone); err != appErrors.ErrSmsDailyLimit {
		t.Fatalf("phone daily limit: got %v", err)
	}
}

func TestReserveSMSGlobalLimitAndRedisFailure(t *testing.T) {
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()

	for i := 0; i < 100; i++ {
		if err := reserveSMS(ctx, client, fmt.Sprintf("1%010d", i)); err != nil {
			t.Fatalf("global reservation %d: %v", i+1, err)
		}
	}
	if err := reserveSMS(ctx, client, "19999999999"); err != appErrors.ErrSmsDailyLimit {
		t.Fatalf("global daily limit: got %v", err)
	}
	if err := reserveSMS(ctx, nil, "13800138000"); err == nil {
		t.Fatal("nil Redis client must reject SMS")
	}
	server.Close()
	if err := reserveSMS(ctx, client, "13800138000"); err == nil {
		t.Fatal("Redis failure must reject SMS")
	}
}

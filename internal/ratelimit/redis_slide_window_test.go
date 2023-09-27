package ratelimit

import (
	"context"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestRedisSlidingWindowLimiter_Limit(t *testing.T) {
	r := initLimiter()
	testCases := []struct {
		name     string
		ctx      context.Context
		key      string
		interval time.Duration
		want     bool
		wantErr  error
	}{
		{
			name: "正常通过！",
			ctx:  context.Background(),
			key:  "xxx",
			want: false,
		},
		{
			name: "另一个key正常通过！",
			ctx:  context.Background(),
			key:  "yyy",
			want: false,
		},
		{
			name:     "触发限流",
			ctx:      context.Background(),
			key:      "xxx",
			interval: 300 * time.Millisecond,
			want:     true,
		},
		{
			name:     "窗口有空余正常通过",
			ctx:      context.Background(),
			key:      "xxx",
			interval: 510 * time.Millisecond,
			want:     false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			<-time.After(tc.interval)
			isLimited, err := r.Limit(tc.ctx, tc.key)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.want, isLimited)
		})
	}
}

func initLimiter() *RedisSlidingWindowLimiter {
	return &RedisSlidingWindowLimiter{
		Cmd:      initRedis(),
		Interval: 500 * time.Millisecond,
		Rate:     1,
	}
}

func initRedis() redis.Cmdable {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:16379",
	})
	return redisClient
}

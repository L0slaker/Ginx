package ratelimit

import (
	"context"
	_ "embed"
	"github.com/redis/go-redis/v9"
	"time"
)

//go:embed slide_window.lua
var luaSlideWindowLimiter string

type RedisSlidingWindowLimiter struct {
	Cmd redis.Cmdable
	// 窗口大小
	Interval time.Duration
	// 阈值
	Rate int
}

func (r *RedisSlidingWindowLimiter) Limit(ctx context.Context, key string) (bool, error) {
	return r.Cmd.Eval(ctx, luaSlideWindowLimiter, []string{key},
		r.Interval.Milliseconds(), r.Rate, time.Now().UnixMilli()).Bool()
}

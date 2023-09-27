package ratelimit

import (
	"ginx/internal/ratelimit"
	"github.com/redis/go-redis/v9"
	"time"
)

func NewRedisSlidingWindowLimiter(cmd redis.Cmdable,
	interval time.Duration, rate int) ratelimit.Limiter {
	return &ratelimit.RedisSlidingWindowLimiter{
		Cmd:      cmd,
		Interval: interval,
		Rate:     rate,
	}
}

package redislimit

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/atomic"
	"net/http"
)

type RedisActiveLimit struct {
	cmd redis.Cmdable
	// 记录连接数
	key string
	// 最大限流数量
	maxActive *atomic.Int64
	logFn     func(msg any, args ...any)
}

func NewRedisActiveLimit(cmd redis.Cmdable, maxAcitve int64, key string) *RedisActiveLimit {
	return &RedisActiveLimit{
		cmd:       cmd,
		key:       key,
		maxActive: atomic.NewInt64(maxAcitve),
		logFn: func(msg any, args ...any) {
			fmt.Println(fmt.Sprintf("%v detail info %v", msg, args))
		},
	}
}

func (limit *RedisActiveLimit) SetMaxActive(maxActive int64) *RedisActiveLimit {
	limit.maxActive.Store(maxActive)
	return limit
}

func (limit *RedisActiveLimit) SetLogFunc(fun func(msg any, args ...any)) *RedisActiveLimit {
	limit.logFn = fun
	return limit
}

func (limit *RedisActiveLimit) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		current, err := limit.cmd.Incr(ctx, limit.key).Result()
		if err != nil {
			// 记录日志
			limit.logFn("redis + 1", err)
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		defer func() {
			if err = limit.cmd.Decr(ctx, limit.key).Err(); err != nil {
				limit.logFn("redis - 1", err)
				return
			}
		}()
		if current <= limit.maxActive.Load() {
			ctx.Next()
		} else {
			// 执行限流
			limit.logFn("web server->", "执行限流中")
			ctx.AbortWithStatus(http.StatusTooManyRequests)
			return
		}
	}
}

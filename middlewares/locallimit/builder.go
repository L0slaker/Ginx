package locallimit

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/atomic"
	"net/http"
)

type LocalActiveLimit struct {
	// 最大限流数量
	maxActive *atomic.Int64
	// 当前活跃数量
	countActive *atomic.Int64
}

func NewLocalActiveLimit(maxActive int64) *LocalActiveLimit {
	return &LocalActiveLimit{
		maxActive:   atomic.NewInt64(maxActive),
		countActive: atomic.NewInt64(0),
	}
}

func (limit *LocalActiveLimit) SetMaxActive(maxActive int64) *LocalActiveLimit {
	limit.maxActive.Store(maxActive)
	return limit
}

func (limit *LocalActiveLimit) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		current := limit.countActive.Add(1)
		defer func() {
			limit.countActive.Sub(1)
		}()
		if current <= limit.maxActive.Load() {
			ctx.Next()
		} else {
			// 执行限流
			ctx.AbortWithStatus(http.StatusTooManyRequests)
		}
		return
	}
}

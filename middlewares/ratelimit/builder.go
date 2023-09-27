package ratelimit

import (
	"ginx/internal/ratelimit"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strings"
)

type Builder struct {
	limiter ratelimit.Limiter
	// genKeyFn 默认使用 IP 限流
	genKeyFn func(ctx *gin.Context) string
	// logFn 默认使用 log.Println()
	logFn func(msg any, args ...any)
}

func NewBuilder(limiter ratelimit.Limiter) *Builder {
	return &Builder{
		limiter: limiter,
		genKeyFn: func(ctx *gin.Context) string {
			var b strings.Builder
			b.WriteString("ip-limiter")
			b.WriteString(":")
			b.WriteString(ctx.ClientIP())
			return b.String()
		},
		logFn: func(msg any, args ...any) {
			v := make([]any, 0, len(args)-1)
			v = append(v, msg)
			v = append(v, args...)
			log.Println(v...)
		},
	}
}

func (b *Builder) SetKeyGenFunc(fn func(ctx *gin.Context) string) *Builder {
	b.genKeyFn = fn
	return b
}

func (b *Builder) SetLogFunc(fn func(msg any, args ...any)) *Builder {
	b.logFn = fn
	return b
}

func (b *Builder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		isLimited, err := b.limit(ctx)
		if err != nil {
			b.logFn(err)
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		if isLimited {
			ctx.AbortWithStatus(http.StatusTooManyRequests)
			return
		}
		ctx.Next()
	}
}

func (b *Builder) limit(ctx *gin.Context) (bool, error) {
	return b.limiter.Limit(ctx, b.genKeyFn(ctx))
}

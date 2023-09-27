package integration

import (
	"context"
	"fmt"
	limit "ginx/middlewares/ratelimit"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestBuilder_e2e_RateLimit(t *testing.T) {
	const (
		ip       = "127.0.0.1"
		limitURL = "/limit"
	)
	rdb := initRedis()
	server := initWebServer(rdb)
	RegisterRoutes(server)

	testCases := []struct {
		name     string
		before   func(t *testing.T)
		after    func(t *testing.T)
		wantCode int
	}{
		{
			name: "不限流",
			before: func(t *testing.T) {

			},
			after: func(t *testing.T) {
				rdb.Del(context.Background(), fmt.Sprintf("ip-limiter:%s", ip))
			},
			wantCode: http.StatusOK,
		},
		{
			name: "限流",
			before: func(t *testing.T) {
				req, err := http.NewRequest(http.MethodGet, limitURL, nil)
				req.RemoteAddr = ip + ":80"
				assert.NoError(t, err)
				resp := httptest.NewRecorder()
				server.ServeHTTP(resp, req)
			},
			after: func(t *testing.T) {
				rdb.Del(context.Background(), fmt.Sprintf("ip-limiter:%s", ip))
			},
			wantCode: http.StatusTooManyRequests,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer tc.after(t)
			tc.before(t)
			req, err := http.NewRequest(http.MethodGet, limitURL, nil)
			req.RemoteAddr = ip + ":80"
			assert.NoError(t, err)

			resp := httptest.NewRecorder()
			server.ServeHTTP(resp, req)

			code := resp.Code
			assert.Equal(t, tc.wantCode, code)
		})
	}
}

func RegisterRoutes(server *gin.Engine) {
	server.GET("/limit", func(ctx *gin.Context) {
		ctx.Status(http.StatusOK)
	})
}

func initRedis() redis.Cmdable {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:16379",
	})
	return redisClient
}

func initWebServer(cmd redis.Cmdable) *gin.Engine {
	server := gin.Default()
	limiter := limit.NewRedisSlidingWindowLimiter(cmd, 500*time.Millisecond, 1)
	server.Use(limit.NewBuilder(limiter).Build())
	return server
}

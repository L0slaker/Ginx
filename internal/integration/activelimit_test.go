package integration

import (
	"context"
	"fmt"
	"ginx/middlewares/redislimit"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestBuilder_e2e_ActiveRedisLimit(t *testing.T) {
	// var redisClient *redis.Client
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:16379",
		Password: "",
		DB:       0,
	})
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	e := redisClient.Ping(ctx).Err()
	if e != nil {
		panic("redislimit 连接失败")
	}
	defer func() {
		_ = redisClient.Close()
	}()

	testCases := []struct {
		name      string
		key       string
		maxActive int64
		interval  time.Duration
		before    func(server *gin.Engine, key string)
		after     func(str string) (int64, error)

		wantCode int
		//检查退出的时候redis 状态
		afterCount int64
		afterErr   error
	}{
		{
			name:      "开启限流，限流器正常操作",
			key:       "test",
			maxActive: 1,
			interval:  time.Millisecond * 10,
			before: func(server *gin.Engine, key string) {

			},
			after: func(key string) (int64, error) {
				return redisClient.Get(context.Background(), key).Int64()
			},
			wantCode:   http.StatusOK,
			afterCount: 0,
			afterErr:   nil,
		},
		{
			name:      "开启限流，有一个请求超时未退出，引发限流",
			key:       "test",
			maxActive: 1,
			interval:  time.Millisecond * 50,
			before: func(server *gin.Engine, key string) {
				req, err := http.NewRequest(http.MethodGet, "/activelimit3", nil)
				require.NoError(t, err)
				resp := httptest.NewRecorder()
				server.ServeHTTP(resp, req)
				assert.Equal(t, http.StatusOK, resp.Code)
			},
			after: func(key string) (int64, error) {
				return redisClient.Get(context.Background(), key).Int64()
			},
			wantCode:   http.StatusTooManyRequests,
			afterCount: 1,
			afterErr:   nil,
		},
		{
			name:      "开启限流，有一个请求超时未退出，等待前一个请求退出后，正常请求",
			key:       "test",
			maxActive: 1,
			interval:  time.Millisecond * 200,
			before: func(server *gin.Engine, key string) {
				req, err := http.NewRequest(http.MethodGet, "/activelimit3", nil)
				require.NoError(t, err)
				resp := httptest.NewRecorder()
				server.ServeHTTP(resp, req)
				assert.Equal(t, http.StatusOK, resp.Code)
			},
			after: func(key string) (int64, error) {
				return redisClient.Get(context.Background(), key).Int64()
			},
			wantCode:   http.StatusOK,
			afterCount: 0,
			afterErr:   nil,
		},
	}
	gin.SetMode(gin.ReleaseMode)
	for _, tc := range testCases {
		time.Sleep(time.Millisecond * 100)
		redisClient.Del(context.Background(), tc.key)
		fmt.Println(redisClient.Get(context.Background(), tc.key).Int64())
		t.Run(tc.name, func(t *testing.T) {
			server := gin.Default()
			middleware := redislimit.NewRedisActiveLimit(redisClient, tc.maxActive, tc.key).Build()
			server.Use(middleware)
			server.GET("/activelimit", func(ctx *gin.Context) {
				ctx.Status(http.StatusOK)
			})
			server.GET("/activelimit3", func(ctx *gin.Context) {
				time.Sleep(time.Millisecond * 100)
				ctx.Status(http.StatusOK)
			})

			req, err := http.NewRequest(http.MethodGet, "/activelimit", nil)
			require.NoError(t, err)
			resp := httptest.NewRecorder()

			go func() {
				tc.before(server, tc.key)
			}()
			time.Sleep(tc.interval)
			server.ServeHTTP(resp, req)

			assert.Equal(t, tc.wantCode, resp.Code)
			afterCount, errs := tc.after(tc.key)
			assert.Equal(t, tc.afterCount, afterCount)
			assert.Equal(t, tc.afterErr, errs)
		})
	}
}

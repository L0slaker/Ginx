package redislimit

import (
	"context"
	"errors"
	"ginx/middlewares/redislimit/redismocks"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRedisActiveLimit_Build(t *testing.T) {
	testCases := []struct {
		name          string
		maxCount      int64
		key           string
		interval      time.Duration
		mock          func(ctrl *gomock.Controller, key string) redis.Cmdable
		setMiddleware func(redisClient redis.Cmdable) gin.HandlerFunc
		before        func(server *gin.Engine, key string)
		after         func(str string) (int64, error)

		wantCode int
	}{
		{
			name:     "开启限流，限流器运行正常",
			maxCount: 1,
			key:      "test",
			mock: func(ctrl *gomock.Controller, key string) redis.Cmdable {
				redisClient := redismocks.NewMockCmdable(ctrl)
				res1 := redis.NewIntCmd(context.Background())
				res1.SetErr(nil)
				res1.SetVal(int64(1))
				redisClient.EXPECT().Incr(gomock.Any(), key).Return(res1)

				res2 := redis.NewIntCmd(context.Background())
				res2.SetErr(nil)
				res2.SetVal(int64(0))
				redisClient.EXPECT().Decr(gomock.Any(), key).Return(res2)
				return redisClient
			},
			setMiddleware: func(redisClient redis.Cmdable) gin.HandlerFunc {
				return NewRedisActiveLimit(redisClient, 1, "test").Build()
			},
			before: func(server *gin.Engine, key string) {

			},
			wantCode: http.StatusOK,
		},
		{
			name:     "开启限流，正常操作，但 -1 操作异常",
			maxCount: 1,
			key:      "test",
			mock: func(ctrl *gomock.Controller, key string) redis.Cmdable {
				redisClient := redismocks.NewMockCmdable(ctrl)
				res1 := redis.NewIntCmd(context.Background())
				res1.SetErr(nil)
				res1.SetVal(int64(1))
				redisClient.EXPECT().Incr(gomock.Any(), key).Return(res1)

				res2 := redis.NewIntCmd(context.Background())
				res2.SetErr(errors.New("-1 操作异常"))
				res2.SetVal(int64(0))
				redisClient.EXPECT().Decr(gomock.Any(), key).Return(res2)
				return redisClient
			},
			setMiddleware: func(redisClient redis.Cmdable) gin.HandlerFunc {
				return NewRedisActiveLimit(redisClient, 1, "test").Build()
			},
			before: func(server *gin.Engine, key string) {

			},
			wantCode: http.StatusOK,
		},
		{
			name:     "开启限流，有一个请求超时未退出，导致限流",
			maxCount: 1,
			key:      "test",
			interval: time.Millisecond * 20,
			mock: func(ctrl *gomock.Controller, key string) redis.Cmdable {
				redisClient := redismocks.NewMockCmdable(ctrl)
				res1 := redis.NewIntCmd(context.Background())
				res1.SetErr(nil)
				res1.SetVal(int64(1))
				redisClient.EXPECT().Incr(gomock.Any(), key).Return(res1)

				res2 := redis.NewIntCmd(context.Background())
				res2.SetErr(nil)
				res2.SetVal(int64(2))
				redisClient.EXPECT().Incr(gomock.Any(), key).Return(res2)

				res3 := redis.NewIntCmd(context.Background())
				res3.SetErr(nil)
				res3.SetVal(int64(1))
				redisClient.EXPECT().Decr(gomock.Any(), key).Return(res3).AnyTimes()
				return redisClient
			},
			setMiddleware: func(redisClient redis.Cmdable) gin.HandlerFunc {
				return NewRedisActiveLimit(redisClient, 1, "test").Build()
			},
			before: func(server *gin.Engine, key string) {
				req, err := http.NewRequest(http.MethodGet, "/activelimit3", nil)
				require.NoError(t, err)
				resp := httptest.NewRecorder()
				server.ServeHTTP(resp, req)
				assert.Equal(t, 200, resp.Code)
			},
			wantCode: http.StatusTooManyRequests,
		},
		{
			name:     "系统异常",
			maxCount: 1,
			key:      "test",
			mock: func(ctrl *gomock.Controller, key string) redis.Cmdable {
				redisClient := redismocks.NewMockCmdable(ctrl)
				res1 := redis.NewIntCmd(context.Background())
				res1.SetErr(errors.New("redis 异常"))
				res1.SetVal(int64(1))
				redisClient.EXPECT().Incr(gomock.Any(), key).Return(res1)
				return redisClient
			},
			setMiddleware: func(redisClient redis.Cmdable) gin.HandlerFunc {
				return NewRedisActiveLimit(redisClient, 1, "test").Build()
			},
			before: func(server *gin.Engine, key string) {

			},
			wantCode: http.StatusInternalServerError,
		},
	}
	gin.SetMode(gin.ReleaseMode)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			server := gin.Default()
			server.Use(tc.setMiddleware(tc.mock(ctrl, tc.key)))
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
		})
	}
}

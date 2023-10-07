package locallimit

import (
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestLocalActiveLimit_Build(t *testing.T) {
	testCases := []struct {
		name     string
		maxCount int64
		// 窗口大小
		interval time.Duration
		before   func(server *gin.Engine)
		after    func()
		// 响应的code
		wantCode int
	}{
		{
			name:     "开启限流，限流器运行正常",
			maxCount: 1,
			before:   func(server *gin.Engine) {},
			after:    func() {},
			wantCode: http.StatusOK,
		},
		{
			name:     "开启限流，有请求一个很久没出来,新请求被限流",
			maxCount: 1,
			before: func(server *gin.Engine) {
				req, err := http.NewRequest(http.MethodGet, "/activelimit3", nil)
				require.NoError(t, err)
				resp := httptest.NewRecorder()
				server.ServeHTTP(resp, req)
				assert.Equal(t, http.StatusOK, resp.Code)
			},
			after:    func() {},
			wantCode: http.StatusTooManyRequests,
		},
		{
			name:     "开启限流，有一个请求很久没出来,等待前面的请求退出后,成功通过",
			maxCount: 1,
			interval: time.Millisecond * 600,
			before: func(server *gin.Engine) {
				req, err := http.NewRequest(http.MethodGet, "/activelimit3", nil)
				require.NoError(t, err)
				resp := httptest.NewRecorder()
				server.ServeHTTP(resp, req)
				assert.Equal(t, http.StatusOK, resp.Code)
			},
			after:    func() {},
			wantCode: http.StatusOK,
		},
	}
	gin.SetMode(gin.ReleaseMode)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := gin.Default()
			limiter := NewLocalActiveLimit(tc.maxCount)
			middleware := limiter.Build()
			server.Use(middleware)

			server.GET("/activelimit", func(ctx *gin.Context) {
				ctx.Status(http.StatusOK)
			})
			server.GET("/activelimit3", func(ctx *gin.Context) {
				time.Sleep(time.Millisecond * 300)
				ctx.Status(http.StatusOK)
			})

			req, err := http.NewRequest(http.MethodGet, "/activelimit", nil)
			require.NoError(t, err)
			resp := httptest.NewRecorder()

			go func() {
				tc.before(server)
			}()
			//加延时保证 tc.before 执行
			time.Sleep(time.Millisecond * 10)
			time.Sleep(tc.interval)
			server.ServeHTTP(resp, req)
			assert.Equal(t, tc.wantCode, resp.Code)

			tc.after()
		})
	}
}

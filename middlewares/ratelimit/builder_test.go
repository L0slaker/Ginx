package ratelimit

import (
	"errors"
	"ginx/internal/ratelimit"
	limitmocks "ginx/internal/ratelimit/mocks"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBuilder_SetKeyGenFunc(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	testCases := []struct {
		name       string
		reqBuilder func(t *testing.T) *http.Request
		fn         func(ctx *gin.Context) string
		want       string
	}{
		{
			name: "设置key成功！",
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "", nil)
				if err != nil {
					t.Fatal(err)
				}
				req.RemoteAddr = "127.0.0.1:80"
				return req
			},
			fn: func(ctx *gin.Context) string {
				return "test"
			},
			want: "test",
		},
		{
			name: "默认key",
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "", nil)
				if err != nil {
					t.Fatal(err)
				}
				req.RemoteAddr = "127.0.0.1:80"
				return req
			},
			want: "ip-limiter:127.0.0.1",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b := NewBuilder(nil)
			if tc.fn != nil {
				b.SetKeyGenFunc(tc.fn)
			}

			resp := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(resp)
			req := tc.reqBuilder(t)
			ctx.Request = req

			assert.Equal(t, tc.want, b.genKeyFn(ctx))
		})
	}
}

func TestBuilder_Build(t *testing.T) {
	const limitURL = "/limit"
	testCases := []struct {
		name       string
		mock       func(ctrl *gomock.Controller) ratelimit.Limiter
		reqBuilder func(t *testing.T) *http.Request
		wantCode   int
	}{
		{
			name: "不限流",
			mock: func(ctrl *gomock.Controller) ratelimit.Limiter {
				limiter := limitmocks.NewMockLimiter(ctrl)
				limiter.EXPECT().Limit(gomock.Any(), gomock.Any()).
					Return(false, nil)
				return limiter
			},
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, limitURL, nil)
				if err != nil {
					t.Fatal(err)
				}
				return req
			},
			wantCode: http.StatusOK,
		},
		{
			name: "限流",
			mock: func(ctrl *gomock.Controller) ratelimit.Limiter {
				limiter := limitmocks.NewMockLimiter(ctrl)
				limiter.EXPECT().Limit(gomock.Any(), gomock.Any()).
					Return(true, nil)
				return limiter
			},
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, limitURL, nil)
				if err != nil {
					t.Fatal(err)
				}
				return req
			},
			wantCode: http.StatusTooManyRequests,
		},
		{
			name: "系统错误！",
			mock: func(ctrl *gomock.Controller) ratelimit.Limiter {
				limiter := limitmocks.NewMockLimiter(ctrl)
				limiter.EXPECT().Limit(gomock.Any(), gomock.Any()).
					Return(false, errors.New("模拟系统错误"))
				return limiter
			},
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, limitURL, nil)
				if err != nil {
					t.Fatal(err)
				}
				return req
			},
			wantCode: http.StatusInternalServerError,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc := NewBuilder(tc.mock(ctrl))
			server := gin.Default()
			server.Use(svc.Build())
			svc.RegisterRoutes(server)

			req := tc.reqBuilder(t)
			resp := httptest.NewRecorder()
			server.ServeHTTP(resp, req)

			assert.Equal(t, tc.wantCode, resp.Code)
		})
	}
}

func TestBuilder_limit(t *testing.T) {
	testCases := []struct {
		name       string
		mock       func(ctrl *gomock.Controller) ratelimit.Limiter
		reqBuilder func(t *testing.T) *http.Request
		want       bool
		wantErr    error
	}{
		{
			name: "不限流",
			mock: func(ctrl *gomock.Controller) ratelimit.Limiter {
				limiter := limitmocks.NewMockLimiter(ctrl)
				limiter.EXPECT().Limit(gomock.Any(), gomock.Any()).
					Return(false, nil)
				return limiter
			},
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "", nil)
				if err != nil {
					t.Fatal(err)
				}
				req.RemoteAddr = "127.0.0.1:80"
				return req
			},
			want: false,
		},
		{
			name: "限流",
			mock: func(ctrl *gomock.Controller) ratelimit.Limiter {
				limiter := limitmocks.NewMockLimiter(ctrl)
				limiter.EXPECT().Limit(gomock.Any(), gomock.Any()).
					Return(true, nil)
				return limiter
			},
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "", nil)
				if err != nil {
					t.Fatal(err)
				}
				req.RemoteAddr = "127.0.0.1:80"
				return req
			},
			want: true,
		},
		{
			name: "限流代码出错",
			mock: func(ctrl *gomock.Controller) ratelimit.Limiter {
				limiter := limitmocks.NewMockLimiter(ctrl)
				limiter.EXPECT().Limit(gomock.Any(), gomock.Any()).
					Return(false, errors.New("模拟系统错误"))
				return limiter
			},
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "", nil)
				if err != nil {
					t.Fatal(err)
				}
				req.RemoteAddr = "127.0.0.1:80"
				return req
			},
			want:    false,
			wantErr: errors.New("模拟系统错误"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			limiter := tc.mock(ctrl)
			b := NewBuilder(limiter)

			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			req := tc.reqBuilder(t)
			ctx.Request = req

			got, err := b.limit(ctx)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func (b *Builder) RegisterRoutes(server *gin.Engine) {
	server.GET("/limit", func(ctx *gin.Context) {
		ctx.Status(http.StatusOK)
	})
}

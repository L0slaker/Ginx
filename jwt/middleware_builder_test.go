package jwt

import (
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestMiddlewareBuilder_Build(t *testing.T) {
	type testCase[T any] struct {
		name        string
		manager     *Management[T]
		reqBuilder  func(t *testing.T) *http.Request
		isUseIgnore bool
		wantCode    int
	}
	testCases := []testCase[data]{
		{
			name:    "验证失败",
			manager: NewManagement[data](defaultOption),
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/", nil)
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Add("authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkYXRhIjp7ImZvbyI6IjEifSwiZXhwIjoxNjk1NTcxODAwLCJpYXQiOjE2OTU1NzEyMDB9.RMpM5YNgxl9OtCy4lt_JRxv6k8s6plCkthnAV-vbXEQ")
				return req
			},
			wantCode: http.StatusUnauthorized,
		},
		{
			name: "验证通过",
			manager: NewManagement[data](defaultOption,
				WithNowFunc[data](func() time.Time {
					return time.UnixMilli(1695571500000)
				}),
			),
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/", nil)
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Add("authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkYXRhIjp7ImZvbyI6IjEifSwiZXhwIjoxNjk1NTcxODAwLCJpYXQiOjE2OTU1NzEyMDB9.RMpM5YNgxl9OtCy4lt_JRxv6k8s6plCkthnAV-vbXEQ")
				return req
			},
			wantCode: http.StatusOK,
		},
		{
			name:    "提取 token 失败",
			manager: NewManagement[data](defaultOption),
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/", nil)
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Add("authorization", "Bearer ")
				return req
			},
			wantCode: http.StatusUnauthorized,
		},
		{
			name:    "无需验证，直接通过",
			manager: NewManagement[data](defaultOption),
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/login", nil)
				if err != nil {
					t.Fatal(err)
				}
				return req
			},
			isUseIgnore: true,
			wantCode:    http.StatusOK,
		},
		{
			//
			name:    "未使用忽略选项则进行拦截",
			manager: NewManagement[data](defaultOption),
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/login", nil)
				if err != nil {
					t.Fatal(err)
				}
				return req
			},
			wantCode: http.StatusUnauthorized,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := gin.Default()
			m := tc.manager.MiddlewareBuilder()
			if tc.isUseIgnore {
				m = m.IgnorePath("/login")
			}
			server.Use(m.Build())
			tc.manager.registerRoutes(server)

			req := tc.reqBuilder(t)
			resp := httptest.NewRecorder()

			server.ServeHTTP(resp, req)
			assert.Equal(t, tc.wantCode, resp.Code)
		})
	}
}

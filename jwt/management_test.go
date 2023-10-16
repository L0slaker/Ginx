package jwt

import (
	"fmt"
	"github.com/ecodeclub/ekit/bean/option"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type data struct {
	Foo string `json:"foo"`
}

var (
	defaultExpire        = 10 * time.Minute
	defaultEncryptionKey = "sign key"
	defaultMethod        = jwt.SigningMethodHS256
	// 2023-11-27 13:20:00 UTC
	now           = time.UnixMilli(1695571200000)
	defaultOption = NewOptions(defaultExpire, defaultEncryptionKey)
	defaultClaims = RegisteredClaims[data]{
		Data: data{Foo: "1"},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(defaultExpire)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	defaultManagement = NewManagement[data](defaultOption,
		WithNowFunc[data](func() time.Time {
			return now
		}),
	)
)

func TestManagement_Refresh(t *testing.T) {
	type testCase[T any] struct {
		name             string
		manager          *Management[T]
		reqBuilder       func(t *testing.T) *http.Request
		wantCode         int
		wantAccessToken  string
		wantRefreshToken string
	}
	testCases := []testCase[data]{
		{
			name: "更新资源令牌并轮换刷新令牌",
			manager: NewManagement[data](defaultOption,
				WithRefreshJWTOptions[data](NewOptions(24*60*time.Minute, "refresh sign key")),
				WithRotateRefreshToken[data](true),
				WithNowFunc[data](func() time.Time {
					return time.UnixMilli(1695623000000)
				}),
			),
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/refresh", nil)
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Add("authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkYXRhIjp7ImZvbyI6IjEifSwiZXhwIjoxNjk1NjU3NjAwLCJpYXQiOjE2OTU1NzEyMDB9.y2AQ98i0le5AbmJFgYCAfCVAphd_9NecmHdhtehMSZE")
				return req
			},
			wantCode:         http.StatusNoContent,
			wantAccessToken:  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkYXRhIjp7ImZvbyI6IjEifSwiZXhwIjoxNjk1NTcxODAwLCJpYXQiOjE2OTU1NzEyMDB9.RMpM5YNgxl9OtCy4lt_JRxv6k8s6plCkthnAV-vbXEQ",
			wantRefreshToken: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkYXRhIjp7ImZvbyI6IjEifSwiZXhwIjoxNjk1NTcxODAwLCJpYXQiOjE2OTU1NzEyMDB9.-ZFWX_iSA3i9JvS8FB4fwtQPOezM-NaRp5_BjP_VQV0",
		},
		{
			name: "更新资源令牌但轮换刷新令牌生成失败",
			manager: NewManagement[data](defaultOption,
				WithRefreshJWTOptions[data](NewOptions(
					24*60*time.Minute,
					"refresh sign key",
					WithMethod(jwt.SigningMethodRS256),
				)),
				WithRotateRefreshToken[data](true),
				WithNowFunc[data](func() time.Time {
					return time.UnixMilli(1695623000000)
				}),
			),
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/refresh", nil)
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Add("authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkYXRhIjp7ImZvbyI6IjEifSwiZXhwIjoxNjk1NjU3NjAwLCJpYXQiOjE2OTU1NzEyMDB9.y2AQ98i0le5AbmJFgYCAfCVAphd_9NecmHdhtehMSZE")
				return req
			},
			wantCode: http.StatusInternalServerError,
		},
		{
			name: "更新资源令牌",
			manager: NewManagement[data](defaultOption,
				WithRefreshJWTOptions[data](NewOptions(24*60*time.Minute, "refresh sign key")),
				WithNowFunc[data](func() time.Time {
					return time.UnixMilli(1695623000000)
				}),
			),
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/refresh", nil)
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Add("authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkYXRhIjp7ImZvbyI6IjEifSwiZXhwIjoxNjk1NjU3NjAwLCJpYXQiOjE2OTU1NzEyMDB9.y2AQ98i0le5AbmJFgYCAfCVAphd_9NecmHdhtehMSZE")
				return req
			},
			wantCode:        http.StatusNoContent,
			wantAccessToken: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkYXRhIjp7ImZvbyI6IjEifSwiZXhwIjoxNjk1NTcxODAwLCJpYXQiOjE2OTU1NzEyMDB9.RMpM5YNgxl9OtCy4lt_JRxv6k8s6plCkthnAV-vbXEQ",
		},
		{
			name: "更新资源令牌失败",
			manager: NewManagement[data](
				Options{
					Expire:        10 * time.Minute,
					EncryptionKey: defaultEncryptionKey,
					DecryptKey:    defaultEncryptionKey,
					Method:        jwt.SigningMethodRS256,
					genIDFn:       func() string { return "" },
				},
				WithRefreshJWTOptions[data](NewOptions(24*60*time.Minute, "refresh sign key")),
				WithNowFunc[data](func() time.Time {
					return time.UnixMilli(1695623000000)
				}),
			),
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/refresh", nil)
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Add("authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkYXRhIjp7ImZvbyI6IjEifSwiZXhwIjoxNjk1NjU3NjAwLCJpYXQiOjE2OTU1NzEyMDB9.y2AQ98i0le5AbmJFgYCAfCVAphd_9NecmHdhtehMSZE")
				return req
			},
			wantCode: http.StatusInternalServerError,
		},
		{
			name: "刷新令牌认证失败",
			manager: NewManagement[data](defaultOption,
				WithRefreshJWTOptions[data](NewOptions(24*60*time.Minute, "refresh sign key")),
				WithNowFunc[data](func() time.Time {
					// 已经过期，认证失败
					return time.UnixMilli(1695723000000)
				}),
			),
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/refresh", nil)
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Add("authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkYXRhIjp7ImZvbyI6IjEifSwiZXhwIjoxNjk1NjU3NjAwLCJpYXQiOjE2OTU1NzEyMDB9.y2AQ98i0le5AbmJFgYCAfCVAphd_9NecmHdhtehMSZE")
				return req
			},
			wantCode: http.StatusUnauthorized,
		},
		{
			name: "没有设置刷新令牌选项",
			manager: NewManagement[data](defaultOption,
				WithNowFunc[data](func() time.Time {
					return time.UnixMilli(1695723000000)
				}),
			),
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/refresh", nil)
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Add("authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkYXRhIjp7ImZvbyI6IjEifSwiZXhwIjoxNjk1NjU3NjAwLCJpYXQiOjE2OTU1NzEyMDB9.y2AQ98i0le5AbmJFgYCAfCVAphd_9NecmHdhtehMSZE")
				return req
			},
			wantCode: http.StatusInternalServerError,
		},
	}
	gin.SetMode(gin.ReleaseMode)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := gin.Default()
			tc.manager.registerRoutes(server)

			req := tc.reqBuilder(t)
			resp := httptest.NewRecorder()

			server.ServeHTTP(resp, req)
			assert.Equal(t, tc.wantCode, resp.Code)
			if resp.Code != http.StatusOK {
				return
			}
			assert.Equal(t, tc.wantAccessToken, resp.Header().Get("x-access-token"))
			assert.Equal(t, tc.wantRefreshToken, resp.Header().Get("x-refresh-token"))
		})
	}
}

func TestManagement_GenerateAccessToken(t *testing.T) {
	m := defaultManagement
	type testCase[T any] struct {
		name    string
		data    T
		want    string
		wantErr error
	}
	testCases := []testCase[data]{
		{
			name: "生成 token 成功",
			data: data{Foo: "1"},
			want: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkYXRhIjp7ImZvbyI6IjEifSwiZXhwIjoxNjk1NTcxODAwLCJpYXQiOjE2OTU1NzEyMDB9.RMpM5YNgxl9OtCy4lt_JRxv6k8s6plCkthnAV-vbXEQ",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			token, err := m.GenerateAccessToken(tc.data)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.want, token)
		})
	}
}

func TestManagement_VerifyAccessToken(t *testing.T) {
	type testCase[T any] struct {
		name    string
		manager *Management[T]
		token   string
		want    RegisteredClaims[T]
		wantErr error
	}
	testCases := []testCase[data]{
		{
			name:    "校验 token 成功",
			manager: defaultManagement,
			token:   "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkYXRhIjp7ImZvbyI6IjEifSwiZXhwIjoxNjk1NTcxODAwLCJpYXQiOjE2OTU1NzEyMDB9.RMpM5YNgxl9OtCy4lt_JRxv6k8s6plCkthnAV-vbXEQ",
			want:    defaultClaims,
		},
		{
			// token 过期了
			name: "token_expired",
			manager: NewManagement[data](defaultOption, WithNowFunc[data](func() time.Time {
				return time.UnixMilli(1695671200000)
			})),
			token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkYXRhIjp7ImZvbyI6IjEifSwiZXhwIjoxNjk1NTcxODAwLCJpYXQiOjE2OTU1NzEyMDB9.RMpM5YNgxl9OtCy4lt_JRxv6k8s6plCkthnAV-vbXEQ",
			wantErr: fmt.Errorf("验证失败: %v",
				fmt.Errorf("%v: %v", jwt.ErrTokenInvalidClaims, jwt.ErrTokenExpired)),
		},
		{
			name: "token 签名错误",
			manager: NewManagement[data](defaultOption, WithNowFunc[data](func() time.Time {
				return now
			})),
			token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkYXRhIjp7ImZvbyI6IjEifSwiZXhwIjoxNjk1NTcxODAwLCJpYXQiOjE2OTU1NzEyMDB9.pnP991l48s_j4fkiZnmh48gjgDGult9Or_wLChHvYp0",
			wantErr: fmt.Errorf("验证失败: %v",
				fmt.Errorf("%v: %v", jwt.ErrTokenSignatureInvalid, jwt.ErrSignatureInvalid)),
		},
		{
			name:    "错误的 token",
			manager: defaultManagement,
			token:   "wrong token",
			wantErr: fmt.Errorf("验证失败: %v: token contains an invalid number of segments",
				jwt.ErrTokenMalformed),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			token, err := tc.manager.VerifyAccessToken(tc.token, jwt.WithTimeFunc(tc.manager.nowFunc))
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.want, token)
		})
	}
}

func TestManagement_GenerateRefreshToken(t *testing.T) {
	m := defaultManagement
	type testCase[T any] struct {
		name              string
		refreshJWTOptions *Options
		data              T
		want              string
		wantErr           error
	}
	testsCases := []testCase[data]{
		{
			name: "生成刷新 token 成功",
			refreshJWTOptions: func() *Options {
				opt := NewOptions(24*60*time.Minute, "refresh sign key")
				return &opt
			}(),
			data: data{Foo: "1"},
			want: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkYXRhIjp7ImZvbyI6IjEifSwiZXhwIjoxNjk1NjU3NjAwLCJpYXQiOjE2OTU1NzEyMDB9.y2AQ98i0le5AbmJFgYCAfCVAphd_9NecmHdhtehMSZE",
		},
		{
			name:    "mistake",
			data:    data{Foo: "1"},
			want:    "",
			wantErr: errEmptyRefreshOpts,
		},
	}
	for _, tc := range testsCases {
		t.Run(tc.name, func(t *testing.T) {
			m.refreshJWTOptions = tc.refreshJWTOptions
			token, err := m.GenerateRefreshToken(tc.data)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.want, token)
		})
	}
}

func TestManagement_VerifyRefreshToken(t *testing.T) {
	defaultOpts := Options{
		Expire:        24 * 60 * time.Minute,
		EncryptionKey: "refresh sign key",
		DecryptKey:    "refresh sign key",
		Method:        jwt.SigningMethodHS256,
	}
	type testCase[T any] struct {
		name    string
		m       *Management[T]
		token   string
		want    RegisteredClaims[T]
		wantErr error
	}
	tests := []testCase[data]{
		{
			name: "验证刷新 token 成功",
			m: NewManagement[data](defaultOption,
				WithNowFunc[data](func() time.Time {
					return time.UnixMilli(1695601200000)
				}),
				WithRefreshJWTOptions[data](defaultOpts),
			),
			token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkYXRhIjp7ImZvbyI6IjEifSwiZXhwIjoxNjk1NjU3NjAwLCJpYXQiOjE2OTU1NzEyMDB9.y2AQ98i0le5AbmJFgYCAfCVAphd_9NecmHdhtehMSZE",
			want: RegisteredClaims[data]{
				Data: data{Foo: "1"},
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: jwt.NewNumericDate(now.Add(24 * 60 * time.Minute)),
					IssuedAt:  jwt.NewNumericDate(now),
				},
			},
		},
		{
			name: "token 过期了",
			m: NewManagement[data](defaultOption,
				WithNowFunc[data](func() time.Time {
					return time.UnixMilli(1695701200000)
				}),
				WithRefreshJWTOptions[data](defaultOpts),
			),
			token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkYXRhIjp7ImZvbyI6IjEifSwiZXhwIjoxNjk1NjU3NjAwLCJpYXQiOjE2OTU1NzEyMDB9.y2AQ98i0le5AbmJFgYCAfCVAphd_9NecmHdhtehMSZE",
			wantErr: fmt.Errorf("验证失败: %v",
				fmt.Errorf("%v: %v", jwt.ErrTokenInvalidClaims, jwt.ErrTokenExpired)),
		},
		{
			name: "token 签名错误",
			m: NewManagement[data](defaultOption,
				WithNowFunc[data](func() time.Time {
					return time.UnixMilli(1695601200000)
				}),
				WithRefreshJWTOptions[data](defaultOpts),
			),
			token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkYXRhIjp7ImZvbyI6IjEifSwiZXhwIjoxNjk1NjU3NjAwLCJpYXQiOjE2OTU1NzEyMDB9.yZ_ZlD1jE-0b3qd0bicTDLSdwGsenv6tRmOEqMCM2uw",
			wantErr: fmt.Errorf("验证失败: %v",
				fmt.Errorf("%v: %v", jwt.ErrTokenSignatureInvalid, jwt.ErrSignatureInvalid)),
		},
		{
			name: "错误的 token",
			m: NewManagement[data](defaultOption,
				WithNowFunc[data](func() time.Time {
					return time.UnixMilli(1695601200000)
				}),
				WithRefreshJWTOptions[data](defaultOpts),
			),
			token: "bad_token",
			wantErr: fmt.Errorf("验证失败: %v: token contains an invalid number of segments",
				jwt.ErrTokenMalformed),
		},
		{
			name: "没有刷新 token 的配置",
			m: NewManagement[data](defaultOption,
				WithNowFunc[data](func() time.Time {
					return time.UnixMilli(1695601200000)
				}),
			),
			token:   "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkYXRhIjp7ImZvbyI6IjEifSwiZXhwIjoxNjk1NjU3NjAwLCJpYXQiOjE2OTU1NzEyMDB9.y2AQ98i0le5AbmJFgYCAfCVAphd_9NecmHdhtehMSZE",
			wantErr: errEmptyRefreshOpts,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.m.VerifyRefreshToken(tt.token,
				jwt.WithTimeFunc(tt.m.nowFunc))
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestManagement_SetClaims(t *testing.T) {
	m := defaultManagement
	type testCase[T any] struct {
		name    string
		claims  RegisteredClaims[T]
		want    RegisteredClaims[T]
		wantErr error
	}
	tests := []testCase[data]{
		{
			name:    "设置成功！",
			claims:  defaultClaims,
			want:    defaultClaims,
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
			m.SetClaims(ctx, tt.claims)
			v, ok := ctx.Get("claims")
			if !ok {
				t.Errorf("claims not found")
			}
			clm, ok := v.(RegisteredClaims[data])
			if !ok {
				t.Errorf("claims type error")
			}
			assert.Equal(t, tt.want, clm)
		})
	}
}

func TestManagement_extractTokenString(t *testing.T) {
	m := defaultManagement
	type header struct {
		key   string
		value string
	}
	type testCase[T any] struct {
		name   string
		header header
		want   string
	}
	testCases := []testCase[data]{
		{
			name: "设置成功",
			header: header{
				key:   "authorization",
				value: "Bearer token",
			},
			want: "token",
		},
		{
			name: "前缀有误",
			header: header{
				key:   "authorization",
				value: "bearer token",
			},
		},
		{
			name: "没有 token 的请求头",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(resp)
			req, err := http.NewRequest(http.MethodGet, "", nil)
			req.Header.Add(tc.header.key, tc.header.value)
			if err != nil {
				t.Fatal(err)
			}
			ctx.Request = req
			token := m.extractTokenString(ctx)
			assert.Equal(t, tc.want, token)
		})
	}
}

func TestNewManagement(t *testing.T) {
	type testCase[T any] struct {
		name             string
		accessJWTOptions Options
		wantPanic        bool
	}
	testCases := []testCase[data]{
		{
			name:             "创建成功",
			accessJWTOptions: defaultOption,
			wantPanic:        false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if err := recover(); err != nil {
					if !tc.wantPanic {
						t.Errorf("期望出现panic，但没发生")
					}
				}
			}()
			NewManagement[data](tc.accessJWTOptions)
		})
	}
}

func TestAllowTokenHeader(t *testing.T) {
	type testCase[T any] struct {
		name string
		fn   func() option.Option[Management[T]]
		want string
	}
	testCases := []testCase[data]{
		{
			name: "默认值",
			fn: func() option.Option[Management[data]] {
				return nil
			},
			want: "authorization",
		},
		{
			name: "设置新的请求头",
			fn: func() option.Option[Management[data]] {
				return WithAllowTokenHeader[data]("jwt")
			},
			want: "jwt",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var header string
			if tc.fn() == nil {
				header = NewManagement[data](
					defaultOption,
				).allowTokenHeader
			} else {
				header = NewManagement[data](
					defaultOption,
					tc.fn(),
				).allowTokenHeader
			}
			assert.Equal(t, tc.want, header)
		})
	}
}

func TestWithExposeAccessHeader(t *testing.T) {
	type testCase[T any] struct {
		name string
		fn   func() option.Option[Management[T]]
		want string
	}
	testCases := []testCase[data]{
		{
			name: "默认值",
			fn: func() option.Option[Management[data]] {
				return nil
			},
			want: "x-access-token",
		},
		{
			name: "设置新的请求头",
			fn: func() option.Option[Management[data]] {
				return WithExposeAccessHeader[data]("data-token")
			},
			want: "data-token",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var header string
			if tc.fn() == nil {
				header = NewManagement[data](
					defaultOption,
				).exposeAccessHeader
			} else {
				header = NewManagement[data](
					defaultOption,
					tc.fn(),
				).exposeAccessHeader
			}
			assert.Equal(t, tc.want, header)
		})
	}
}

func TestWithExposeRefreshHeader(t *testing.T) {
	type testCase[T any] struct {
		name string
		fn   func() option.Option[Management[T]]
		want string
	}
	testCases := []testCase[data]{
		{
			name: "默认值",
			fn: func() option.Option[Management[data]] {
				return nil
			},
			want: "x-refresh-token",
		},
		{
			name: "设置新的请求头",
			fn: func() option.Option[Management[data]] {
				return WithExposeRefreshHeader[data]("data-token")
			},
			want: "data-token",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var header string
			if tc.fn() == nil {
				header = NewManagement[data](
					defaultOption,
				).exposeRefreshHeader
			} else {
				header = NewManagement[data](
					defaultOption,
					tc.fn(),
				).exposeRefreshHeader
			}
			assert.Equal(t, tc.want, header)
		})
	}
}

func TestWithRotateRefreshToken(t *testing.T) {
	type testCase[T any] struct {
		name string
		fn   func() option.Option[Management[T]]
		want bool
	}
	testCases := []testCase[data]{
		{
			name: "默认值",
			fn: func() option.Option[Management[data]] {
				return nil
			},
			want: false,
		},
		{
			name: "设置轮换刷新令牌",
			fn: func() option.Option[Management[data]] {
				return WithRotateRefreshToken[data](true)
			},
			want: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var rotate bool
			if tc.fn() == nil {
				rotate = NewManagement[data](
					defaultOption,
				).rotateRefreshToken
			} else {
				rotate = NewManagement[data](
					defaultOption,
					tc.fn(),
				).rotateRefreshToken
			}
			assert.Equal(t, tc.want, rotate)
		})
	}
}

func TestWithNowFunc(t *testing.T) {
	type testCase[T any] struct {
		name string
		fn   func() option.Option[Management[T]]
		want time.Time
	}
	testCases := []testCase[data]{
		{
			name: "默认值",
			fn: func() option.Option[Management[data]] {
				return nil
			},
			want: time.Now(),
		},
		{
			name: "设置新的时间",
			fn: func() option.Option[Management[data]] {
				return WithNowFunc[data](func() time.Time {
					return now
				})
			},
			want: now,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var currentTime time.Time
			if tc.fn() == nil {
				currentTime = NewManagement[data](
					defaultOption,
				).nowFunc()
			} else {
				currentTime = NewManagement[data](
					defaultOption,
					tc.fn(),
				).nowFunc()
			}
			assert.Equal(t, tc.want.Unix(), currentTime.Unix())
		})
	}
}

func TestWithRefreshJWTOptions(t *testing.T) {
	var genIDFn func() string
	type testCase[T any] struct {
		name string
		fn   func() option.Option[Management[T]]
		want *Options
	}
	testCases := []testCase[data]{
		{
			name: "默认值",
			fn: func() option.Option[Management[data]] {
				return nil
			},
			want: nil,
		},
		{
			name: "新的 jwt 配置",
			fn: func() option.Option[Management[data]] {
				return WithRefreshJWTOptions[data](NewOptions(
					24*60*time.Minute,
					"refresh sign key",
					WithGenIDFunc(genIDFn)))
			},
			want: &Options{
				Expire:        24 * 60 * time.Minute,
				EncryptionKey: "refresh sign key",
				DecryptKey:    "refresh sign key",
				Method:        jwt.SigningMethodHS256,
				genIDFn:       genIDFn,
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var opt *Options
			if tc.fn() == nil {
				opt = NewManagement[data](
					defaultOption,
				).refreshJWTOptions
			} else {
				opt = NewManagement[data](
					defaultOption,
					tc.fn(),
				).refreshJWTOptions
			}
			assert.Equal(t, tc.want, opt)
		})
	}
}

func (m *Management[T]) registerRoutes(server *gin.Engine) {
	server.GET("/", func(ctx *gin.Context) {
		ctx.Status(http.StatusOK)
	})
	server.GET("/login", func(ctx *gin.Context) {
		ctx.Status(http.StatusOK)
	})
	server.GET("/refresh", m.Refresh)
}

func Test_GenerateJWTToken(t *testing.T) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, defaultClaims)
	key := []byte(defaultEncryptionKey)
	tokenString, _ := token.SignedString(key)
	fmt.Println(tokenString)
}

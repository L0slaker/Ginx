package jwt

import (
	"github.com/ecodeclub/ekit/bean/option"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewOptions(t *testing.T) {
	var genIDFn func() string
	testCases := []struct {
		name          string
		expire        time.Duration
		encryptionKey string
		want          Options
	}{
		{
			name:          "新建Options成功",
			expire:        15 * time.Minute,
			encryptionKey: "sign key",
			want: Options{
				Expire:        15 * time.Minute,
				EncryptionKey: "sign key",
				DecryptKey:    "sign key",
				Method:        jwt.SigningMethodHS256,
				Issuer:        "lisa",
				genIDFn:       genIDFn,
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			opts := NewOptions(tc.expire, tc.encryptionKey)
			opts.genIDFn = genIDFn
			opts.Issuer = "lisa"
			assert.Equal(t, tc.want, opts)
		})
	}
}

func TestWithDecryptKey(t *testing.T) {
	testCases := []struct {
		name string
		fn   func() option.Option[Options]
		want string
	}{
		{
			name: "默认值",
			fn: func() option.Option[Options] {
				return nil
			},
			want: defaultEncryptionKey,
		},
		{
			name: "设置新的密钥",
			fn: func() option.Option[Options] {
				return WithDecryptKey("another sign key")
			},
			want: "another sign key",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var encryptKey string
			if tc.fn() == nil {
				encryptKey = NewOptions(defaultExpire, defaultEncryptionKey).DecryptKey
			} else {
				encryptKey = NewOptions(defaultExpire, defaultEncryptionKey, tc.fn()).DecryptKey
			}
			assert.Equal(t, tc.want, encryptKey)
		})
	}
}

func TestWithMethod(t *testing.T) {
	testCases := []struct {
		name string
		fn   func() option.Option[Options]
		want jwt.SigningMethod
	}{
		{
			name: "默认值",
			fn: func() option.Option[Options] {
				return nil
			},
			want: defaultMethod,
		},
		{
			name: "设置新的方法",
			fn: func() option.Option[Options] {
				return WithMethod(jwt.SigningMethodES384)
			},
			want: jwt.SigningMethodES384,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var method jwt.SigningMethod
			if tc.fn() == nil {
				method = NewOptions(defaultExpire, defaultEncryptionKey).Method
			} else {
				method = NewOptions(defaultExpire, defaultEncryptionKey, tc.fn()).Method
			}
			assert.Equal(t, tc.want, method)
		})
	}
}

func TestWithIssuer(t *testing.T) {
	testCases := []struct {
		name string
		fn   func() option.Option[Options]
		want string
	}{
		{
			name: "默认值",
			fn: func() option.Option[Options] {
				return nil
			},
			want: "",
		},
		{
			name: "设置发行者",
			fn: func() option.Option[Options] {
				return WithIssuer("lisa")
			},
			want: "lisa",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var issuer string
			if tc.fn() == nil {
				issuer = NewOptions(defaultExpire, defaultEncryptionKey).Issuer
			} else {
				issuer = NewOptions(defaultExpire, defaultEncryptionKey, tc.fn()).Issuer
			}
			assert.Equal(t, tc.want, issuer)
		})
	}
}

func TestWithGenIDFunc(t *testing.T) {
	testCases := []struct {
		name string
		fn   func() option.Option[Options]
		want string
	}{
		{
			name: "默认值",
			fn: func() option.Option[Options] {
				return nil
			},
			want: "",
		},
		{
			name: "设置 jti",
			fn: func() option.Option[Options] {
				return WithGenIDFunc(func() string {
					return "unique id"
				})
			},
			want: "unique id",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var jti string
			if tc.fn() == nil {
				jti = NewOptions(defaultExpire, defaultEncryptionKey).genIDFn()
			} else {
				jti = NewOptions(defaultExpire, defaultEncryptionKey, tc.fn()).genIDFn()
			}
			assert.Equal(t, tc.want, jti)
		})
	}
}

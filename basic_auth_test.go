package basicauth

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vicanso/cod"
	"github.com/vicanso/hes"
)

func TestBasicAuth(t *testing.T) {
	m := NewBasicAuth(Config{
		Validate: func(account, pwd string, c *cod.Context) (bool, error) {
			if account == "tree.xie" && pwd == "password" {
				return true, nil
			}
			if account == "n" {
				return false, hes.New("account is invalid")
			}
			return false, nil
		},
	})
	req := httptest.NewRequest("GET", "https://aslant.site/", nil)

	t.Run("skip", func(t *testing.T) {
		done := false
		mSkip := NewBasicAuth(Config{
			Validate: func(account, pwd string, c *cod.Context) (bool, error) {
				return false, nil
			},
			Skipper: func(c *cod.Context) bool {
				return true
			},
		})
		d := cod.New()
		d.Use(mSkip)
		d.GET("/", func(c *cod.Context) error {
			done = true
			return nil
		})
		resp := httptest.NewRecorder()
		d.ServeHTTP(resp, req)
		if !done {
			t.Fatalf("skip fail")
		}
	})

	t.Run("no auth header", func(t *testing.T) {
		d := cod.New()
		d.Use(m)
		d.GET("/", func(c *cod.Context) error {
			return nil
		})
		resp := httptest.NewRecorder()
		d.ServeHTTP(resp, req)
		if resp.Code != http.StatusUnauthorized {
			t.Fatalf("http status code should be 401")
		}
		if resp.Header().Get(cod.HeaderWWWAuthenticate) != "basic realm=basic auth tips" {
			t.Fatalf("www authenticate header is invalid")
		}
	})

	t.Run("auth value not base64", func(t *testing.T) {
		d := cod.New()
		d.Use(m)
		d.GET("/", func(c *cod.Context) error {
			return nil
		})
		req.Header.Set(cod.HeaderAuthorization, "basic 测试")
		resp := httptest.NewRecorder()
		d.ServeHTTP(resp, req)
		if resp.Code != http.StatusBadRequest ||
			resp.Body.String() != "category=cod-basic-auth, message=illegal base64 data at input byte 0" {
			t.Fatalf("base64 decode fail error is invalid")
		}
	})

	t.Run("auth validate fail", func(t *testing.T) {
		d := cod.New()
		d.Use(m)
		d.GET("/", func(c *cod.Context) error {
			return nil
		})
		req.Header.Set(cod.HeaderAuthorization, "basic YTpi")
		resp := httptest.NewRecorder()
		d.ServeHTTP(resp, req)
		if resp.Code != http.StatusUnauthorized ||
			resp.Body.String() != "category=cod-basic-auth, message=unAuthorized" {
			t.Fatalf("validate fail error is invalid")
		}
		req.Header.Set(cod.HeaderAuthorization, "basic bjph")
		resp = httptest.NewRecorder()
		d.ServeHTTP(resp, req)
		if resp.Code != http.StatusBadRequest ||
			resp.Body.String() != "message=account is invalid" {
			t.Fatalf("validate return error is fail")
		}
	})

	t.Run("validate error", func(t *testing.T) {
		mValidateFail := NewBasicAuth(Config{
			Validate: func(account, pwd string, c *cod.Context) (bool, error) {
				return false, errors.New("abcd")
			},
		})
		d := cod.New()
		d.Use(mValidateFail)
		d.GET("/", func(c *cod.Context) error {
			return nil
		})
		resp := httptest.NewRecorder()
		d.ServeHTTP(resp, req)
		if resp.Code != http.StatusBadRequest ||
			resp.Body.String() != "category=cod-basic-auth, message=abcd" {
			t.Fatalf("validate fail should return error")
		}
	})

	t.Run("auth success", func(t *testing.T) {
		d := cod.New()
		d.Use(m)
		done := false
		d.GET("/", func(c *cod.Context) error {
			done = true
			return nil
		})
		req.Header.Set(cod.HeaderAuthorization, "basic dHJlZS54aWU6cGFzc3dvcmQ=")
		resp := httptest.NewRecorder()
		d.ServeHTTP(resp, req)
		if !done {
			t.Fatalf("auth fail")
		}
	})
}

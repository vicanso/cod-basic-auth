package basicauth

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vicanso/elton"
	"github.com/vicanso/hes"
)

func TestNoVildatePanic(t *testing.T) {
	assert := assert.New(t)
	defer func() {
		r := recover()
		assert.NotNil(r)
		assert.Equal(r.(error), errRequireValidateFunction)
	}()

	New(Config{})
}

func TestBasicAuth(t *testing.T) {
	m := New(Config{
		Validate: func(account, pwd string, c *elton.Context) (bool, error) {
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
		assert := assert.New(t)
		done := false
		mSkip := New(Config{
			Validate: func(account, pwd string, c *elton.Context) (bool, error) {
				return false, nil
			},
			Skipper: func(c *elton.Context) bool {
				return true
			},
		})
		e := elton.New()
		e.Use(mSkip)
		e.GET("/", func(c *elton.Context) error {
			done = true
			return nil
		})
		resp := httptest.NewRecorder()
		e.ServeHTTP(resp, req)
		assert.True(done)
	})

	t.Run("no auth header", func(t *testing.T) {
		assert := assert.New(t)
		e := elton.New()
		e.Use(m)
		e.GET("/", func(c *elton.Context) error {
			return nil
		})
		resp := httptest.NewRecorder()
		e.ServeHTTP(resp, req)
		assert.Equal(resp.Code, http.StatusUnauthorized)
		assert.Equal(resp.Header().Get(elton.HeaderWWWAuthenticate), `basic realm="basic auth tips"`)
	})

	t.Run("auth validate fail", func(t *testing.T) {
		assert := assert.New(t)
		e := elton.New()
		e.Use(m)
		e.GET("/", func(c *elton.Context) error {
			return nil
		})
		req.Header.Set(elton.HeaderAuthorization, "basic YTpi")
		resp := httptest.NewRecorder()
		e.ServeHTTP(resp, req)
		assert.Equal(resp.Code, http.StatusUnauthorized)
		assert.Equal(resp.Body.String(), "category=elton-basic-auth, message=unAuthorized")

		req.Header.Set(elton.HeaderAuthorization, "basic bjph")
		resp = httptest.NewRecorder()
		e.ServeHTTP(resp, req)
		assert.Equal(resp.Code, http.StatusBadRequest)
		assert.Equal(resp.Body.String(), "message=account is invalid")
	})

	t.Run("validate error", func(t *testing.T) {
		assert := assert.New(t)
		mValidateFail := New(Config{
			Validate: func(account, pwd string, c *elton.Context) (bool, error) {
				return false, errors.New("abcd")
			},
		})
		e := elton.New()
		e.Use(mValidateFail)
		e.GET("/", func(c *elton.Context) error {
			return nil
		})
		resp := httptest.NewRecorder()
		e.ServeHTTP(resp, req)
		assert.Equal(resp.Code, http.StatusBadRequest)
		assert.Equal(resp.Body.String(), "category=elton-basic-auth, message=abcd")
	})

	t.Run("auth success", func(t *testing.T) {
		assert := assert.New(t)
		e := elton.New()
		e.Use(m)
		done := false
		e.GET("/", func(c *elton.Context) error {
			done = true
			return nil
		})
		req.Header.Set(elton.HeaderAuthorization, "basic dHJlZS54aWU6cGFzc3dvcmQ=")
		resp := httptest.NewRecorder()
		e.ServeHTTP(resp, req)
		assert.True(done)
	})
}

// https://stackoverflow.com/questions/50120427/fail-unit-tests-if-coverage-is-below-certain-percentage
func TestMain(m *testing.M) {
	// call flag.Parse() here if TestMain uses flags
	rc := m.Run()

	// rc 0 means we've passed,
	// and CoverMode will be non empty if run with -cover
	if rc == 0 && testing.CoverMode() != "" {
		c := testing.Coverage()
		if c < 0.9 {
			fmt.Println("Tests passed but coverage failed at", c)
			rc = -1
		}
	}
	os.Exit(rc)
}

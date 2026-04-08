package cookiesmanager_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	cookiesmanager "github.com/theoguidoux/cookiesmanager"
)

func GetTestingCookieConfig() cookiesmanager.CookieConfig {
	value := fmt.Sprintf("foo")
	path := fmt.Sprintf("/")
	domain := fmt.Sprintf("localhost")
	expires, _ := time.Parse("2006-Jan-02", "2014-Feb-04")
	expires = expires.Add(24 * time.Hour)
	maxAge := 3600
	secure := true
	httpOnly := true
	sameSite := "strict"

	return cookiesmanager.CookieConfig{
		Name:     "test1",
		Value:    &value,
		Path:     &path,
		Domain:   &domain,
		Expires:  &expires,
		MaxAge:   &maxAge,
		Secure:   &secure,
		HttpOnly: &httpOnly,
		SameSite: &sameSite,
	}
}

func TestRemoveCookies(t *testing.T) {
	cfg := cookiesmanager.CreateConfig()
	testingCookie := GetTestingCookieConfig()
	cfg.Remover = append(cfg.Remover, testingCookie)
	ctx := context.Background()
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})

	handler, err := cookiesmanager.New(ctx, next, cfg, "test")
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost", nil)
	if err != nil {
		t.Fatal(err)
	}
	cookies := []*http.Cookie{
		{Name: "test1", Value: "value1|foo", Path: "/"},
		{Name: "test2", Value: "value2", Path: "/"},
	}
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	handler.ServeHTTP(recorder, req)

	cookies = req.Cookies()

	if len(cookies) != 2 {
		t.Errorf("there should be 2 cookies in the request, found %d", len(cookies))
	}

	if cookies[0].Value != "foo" || cookies[0].Name != "test1" {
		t.Error("the expected cookie that should be kept does not match", len(cookies))
	}

	if cookies[1].Value != "value2" || cookies[1].Name != "test2" {
		t.Error("the expected cookie that should be kept does not match", len(cookies))
	}
}

func TestAddNonExistingCookie(t *testing.T) {
	cfg := cookiesmanager.CreateConfig()
	testingCookie := GetTestingCookieConfig()
	cfg.Adder = append(cfg.Adder, testingCookie)
	ctx := context.Background()
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})

	handler, err := cookiesmanager.New(ctx, next, cfg, "test")
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost", nil)
	if err != nil {
		t.Fatal(err)
	}
	cookies := []*http.Cookie{
		{Name: "test2", Value: "value1", Path: "/"},
		{Name: "test3", Value: "value2", Path: "/"},
	}
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	handler.ServeHTTP(recorder, req)

	cookies = req.Cookies()
	if len(cookies) != 3 {
		t.Errorf("there should be 3 cookies in the request, found %d", len(cookies))
	}

	if cookies[0].Value != "value1" || cookies[0].Name != "test2" {
		t.Error("the expected cookie that should be kept does not match", len(cookies))
	}

	if cookies[1].Value != "value2" || cookies[1].Name != "test3" {
		t.Error("the expected cookie that should be kept does not match", len(cookies))
	}

	if cookies[2].Value != *testingCookie.Value || cookies[2].Name != testingCookie.Name {
		t.Error("the expected cookie that should be kept does not match", len(cookies))
	}
}

func TestUpdateCookie(t *testing.T) {
	cfg := cookiesmanager.CreateConfig()
	testingCookie := GetTestingCookieConfig()
	cfg.Adder = append(cfg.Adder, testingCookie)
	ctx := context.Background()
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})

	handler, err := cookiesmanager.New(ctx, next, cfg, "test")
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost", nil)
	if err != nil {
		t.Fatal(err)
	}
	cookies := []*http.Cookie{
		{Name: "test1", Value: "value1", Path: "/"},
		{Name: "test2", Value: "value2", Path: "/"},
	}
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	handler.ServeHTTP(recorder, req)

	cookies = req.Cookies()
	if len(cookies) != 2 {
		t.Errorf("there should be 2 cookies in the request, found %d", len(cookies))
	}

	if cookies[0].Value != "foo" {
		t.Error("the expected cookie that should be kept does not match", len(cookies))
	}

	if cookies[1].Value != "value2" || cookies[1].Name != "test2" {
		t.Error("the expected cookie that should be kept does not match", len(cookies))
	}
}

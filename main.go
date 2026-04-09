// Package cookiesmanager to add a Secure flag to the cookies
package cookiesmanager

import (
	"context"
	"net/http"
)

const setCookieHeader string = "Set-Cookie"

// Config the plugin configuration.
type Config struct {
	Secure bool `json:"secure,omitempty" toml:"secure,omitempty" yaml:"secure,omitempty"`
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{}
}

// CookieMng an plugin with a possible configuration.
type CookieMng struct {
	next   http.Handler
	name   string
	secure bool
}

// New creates new instance of the plugin.
func New(_ context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	return &CookieMng{
		name:   name,
		next:   next,
		secure: config.Secure,
	}, nil
}

func (p *CookieMng) ServeHTTP(rw http.ResponseWriter, req *http.Request) {

	_secure := p.secure

	if req.Header.Get("Upgrade") == "websocket" {
		p.next.ServeHTTP(rw, req)
		return
	}

	// if no tls, don't do anything
	if req.TLS == nil {
		p.next.ServeHTTP(rw, req)
		return
	}

	myWriter := &responseWriter{
		writer: rw,
		secure: _secure,
	}

	p.next.ServeHTTP(myWriter, req)
}

type responseWriter struct {
	writer http.ResponseWriter
	secure bool
}

func (r *responseWriter) Header() http.Header {
	return r.writer.Header()
}

func (r *responseWriter) Write(bytes []byte) (int, error) {
	return r.writer.Write(bytes)
}

func (r *responseWriter) WriteHeader(statusCode int) {
	headers := r.writer.Header()

	// Extract raw Set-Cookie headers
	rawCookies := headers.Get(setCookieHeader)

	// if set-cookie is not present, don't do anything
	if rawCookies == "" {
		r.writer.WriteHeader(statusCode)
		return
	}

	// Delete existing set-cookie headers
	headers.Del(setCookieHeader)

	/*
		Because things are not always as beautiful as we want, this function can't be used
		Because the traefik golang interpret, as of today, only supports until go 1.22, and the function ParseSetCookie was added in go 1.23
	*/
	//cookie, _ := http.ParseSetCookie(rawCookies)

	header := http.Header{}
	header.Add("Set-Cookie", rawCookies)
	req := http.Response{Header: header}
	cookie := req.Cookies()[0]

	// Modify cookie
	cookie.Secure = r.secure

	// write back the modified cookie
	http.SetCookie(r, cookie)

	r.writer.WriteHeader(statusCode)
}

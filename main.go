// Package traefik_plugin_cookie_flags a traefik plugin adding flags to cookies in the response.
package traefik_plugin_cookie_flags //nolint

import (
	"context"
	"net/http"
)

const setCookieHeader string = "Set-Cookie"

// Config the plugin configuration.
type Config struct {
	//	SameSite string `json:"sameSite,omitempty" toml:"sameSite,omitempty" yaml:"sameSite,omitempty"`
	Secure bool `json:"secure,omitempty" toml:"secure,omitempty" yaml:"secure,omitempty"`
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{}
}

// CookieMng an plugin with a possible configuration.
type CookieMng struct {
	next http.Handler
	name string
	//	sameSite string
	secure bool
}

// New creates new instance of the plugin.
func New(_ context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	return &CookieMng{
		name: name,
		next: next,
		//		sameSite: config.SameSite,
		secure: config.Secure,
	}, nil
}

func (p *CookieMng) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	/*
		_sameSite := http.SameSiteDefaultMode //nolint

		switch p.sameSite {
		case "lax":
			_sameSite = http.SameSiteLaxMode
		case "strict":
			_sameSite = http.SameSiteStrictMode
		case "none":
			_sameSite = http.SameSiteNoneMode
		default:
			_sameSite = http.SameSiteDefaultMode
		}
	*/
	_secure := p.secure

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
	cookie, _ := http.ParseSetCookie(rawCookies)

	// Modify cookie
	cookie.Secure = r.secure

	// write back the modified cookie
	http.SetCookie(r, cookie)

	r.writer.WriteHeader(statusCode)
}

// Just enable or disable the secure flag for sticky cookies
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

// CookieMng with possible configuration.
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

	/*
		websockets need to be managed with their own handlers.
		Otherwise, using the http.writer will break them.

		https://www.reddit.com/r/golang/comments/d4h7tk/how_to_use_http_server_and_websocket_server/
	*/
	if req.Header.Get("Connection") == "Upgrade" && req.Header.Get("Upgrade") == "websocket" {
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

	// Get all Set-Cookie headers
	rawCookies := headers.Values("Set-Cookie")

	// if set-cookie was not present, leave
	if len(rawCookies) == 0 {
		r.writer.WriteHeader(statusCode)
		return
	}

	// Delete existing set-cookie headers
	headers.Del(setCookieHeader)

	//cookie, _ := http.ParseSetCookie(rawCookies)
	for _, raw := range rawCookies {

		// From Go >= 1.23, we can use the function http.ParseSetCookie
		// But traefik go interpret is pinned at 1.22 as of today
		header := http.Header{}
		header.Add("Set-Cookie", raw)
		resp := http.Response{Header: header}
		cookies := resp.Cookies()

		if len(cookies) == 0 {
			continue // skip invalid cookies
		}
		cookie := cookies[0]

		// Modify cookie
		cookie.Secure = r.secure

		// write back the modified cookie
		http.SetCookie(r, cookie)
	}

	r.writer.WriteHeader(statusCode)
}

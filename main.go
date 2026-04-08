// Package cookiesmanager traefik middleware plugin.
package cookiesmanager

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type CookieConfig struct {
	Name     string     `json:"name"`
	Value    *string    `json:"value"`
	Path     *string    `json:"path,omitempty"`
	Domain   *string    `json:"domain,omitempty"`
	Expires  *time.Time `json:"expires,omitempty"`
	MaxAge   *int       `json:"maxAge,omitempty"`
	Secure   *bool      `json:"secure,omitempty"`
	HttpOnly *bool      `json:"httpOnly,omitempty"`
	SameSite *string    `json:"sameSite,omitempty"`
}

func (c *CookieConfig) String() string {
	// Only show value if it is not nil
	value := ""
	if c.Value != nil {
		value = fmt.Sprintf("value=%s", *c.Value)
	}

	// Only show path if it is not nil
	path := ""
	if c.Path != nil {
		path = fmt.Sprintf("path=%s", *c.Path)
	}

	// Only show domain if it is not nil
	domain := ""
	if c.Domain != nil {
		domain = fmt.Sprintf("domain=%s", *c.Domain)
	}

	// Only show expires if it is not nil
	expires := ""
	if c.Expires != nil {
		expires = fmt.Sprintf("expires=%s", c.Expires.String())
	}

	// Only show maxAge if it is not nil
	maxAge := ""
	if c.MaxAge != nil {
		maxAge = fmt.Sprintf("maxAge=%d", *c.MaxAge)
	}

	// Only show secure if it is not nil
	secure := ""
	if c.Secure != nil {
		secure = fmt.Sprintf("secure=%t", *c.Secure)
	}

	// Only show httpOnly if it is not nil
	httpOnly := ""
	if c.HttpOnly != nil {
		httpOnly = fmt.Sprintf("httpOnly=%t", *c.HttpOnly)
	}

	// Only show sameSite if it is not nil
	sameSite := ""
	if c.SameSite != nil {
		sameSite = fmt.Sprintf("sameSite=%s", *c.SameSite)
	}

	return fmt.Sprintf("CookieConfig{name=%s, %s, %s, %s, %s, %s, %s, %s, %s}", c.Name, value, path, domain, expires, maxAge, secure, httpOnly, sameSite)
}
func (c *CookieConfig) SamesiteFromString(s string) http.SameSite {
	switch s {
	case "lax":
		return http.SameSiteLaxMode
	case "strict":
		return http.SameSiteStrictMode
	case "none":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteDefaultMode
	}
}
func (c *CookieConfig) ToHttpCookie() http.Cookie {
	// Convert CookieConfig to http.Cookie
	// Only set property values if they are not nil

	cookie := http.Cookie{
		Name: c.Name,
	}

	if c.Value != nil {
		cookie.Value = *c.Value
	}

	if c.Path != nil {
		cookie.Path = *c.Path
	}

	if c.Domain != nil {
		cookie.Domain = *c.Domain
	}

	if c.Expires != nil {
		cookie.Expires = *c.Expires
	}

	if c.MaxAge != nil {
		cookie.MaxAge = *c.MaxAge
	}

	if c.Secure != nil {
		cookie.Secure = *c.Secure
	}

	if c.HttpOnly != nil {
		cookie.HttpOnly = *c.HttpOnly
	}

	if c.SameSite != nil {
		cookie.SameSite = c.SamesiteFromString(*c.SameSite)
	}

	return cookie
}

// Config the plugin configuration.
type Config struct {
	Adder   []CookieConfig `json:"adder,omitempty"`
	Remover []CookieConfig `json:"remover,omitempty"`
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{
		Adder:   []CookieConfig{},
		Remover: []CookieConfig{},
	}
}

type CookieManager struct {
	next   http.Handler
	Config *Config
	name   string
}

func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	return &CookieManager{
		Config: config,
		next:   next,
		name:   name,
	}, nil
}

func MergeCookies(c1 []*http.Cookie, c2 []*http.Cookie) []*http.Cookie {
	// Merge cookies with c2 overriding c1
	// If cookie does not exist in c1, add it
	// If cookie exists in c1 and c2, override each values with c2 values but keep c1 values if c2 values are empty

	// Create a map of c1 cookies
	c1Map := make(map[string]*http.Cookie)

	for _, cookie := range c1 {
		c1Map[cookie.Name] = cookie
	}

	// Merge c2 cookies into c1 cookies
	// with each property values of c2 overriding c1 property values
	for _, cookie := range c2 {
		// Check if cookie already exists
		c1Cookie, ok := c1Map[cookie.Name]

		// If cookie does not exist, add it
		if !ok {
			c1Map[cookie.Name] = cookie
			continue
		}

		// If cookie exists, override each property values with c2 property values
		if cookie.Value != "" {
			c1Cookie.Value = cookie.Value
		}
		if cookie.Path != "" {
			c1Cookie.Path = cookie.Path
		}
		if cookie.Domain != "" {
			c1Cookie.Domain = cookie.Domain
		}
		if !cookie.Expires.IsZero() {
			c1Cookie.Expires = cookie.Expires
		}
		if cookie.MaxAge != 0 {
			c1Cookie.MaxAge = cookie.MaxAge
		}
		if cookie.Secure {
			c1Cookie.Secure = cookie.Secure
		}
		if cookie.HttpOnly {
			c1Cookie.HttpOnly = cookie.HttpOnly
		}
		if cookie.SameSite != 0 {
			c1Cookie.SameSite = cookie.SameSite
		}

		// Update cookie in c1
		c1Map[cookie.Name] = c1Cookie
	}

	// Convert map to slice
	c := []*http.Cookie{}
	for _, cookie := range c1Map {
		c = append(c, cookie)
	}

	return c

}

type responseWriter struct {
	http.ResponseWriter
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	headers := rw.Header()

	cookies := headers["Set-Cookie"]

	newCookies := []string{}

	for _, c := range cookies {
		// Add Secure if missing
		if !strings.Contains(strings.ToLower(c), "secure") {
			c += "; Secure"
		}
		newCookies = append(newCookies, c)
	}

	headers["Set-Cookie"] = newCookies

	rw.ResponseWriter.WriteHeader(statusCode)
}

func (c *CookieManager) ServeHTTP(rw http.ResponseWriter, req *http.Request) {

	// Make sure adder cookies are merged with existing cookies
	// the adder cookies will override existing cookies or add new ones if does not exist
	// the remover cookies will be removed from the existing cookies
	// We should take care of the Set-Cookie header as well

	// Get existing cookies
	//cookies := req.Cookies()
	setCookieStr := req.Header.Get("Set-Cookie")
	setCookie := &http.Cookie{}

	// Convert adder and remover to http.Cookie
	adderCookies := []*http.Cookie{}
	for _, cookie := range c.Config.Adder {
		c := cookie.ToHttpCookie()
		adderCookies = append(adderCookies, &c)
	}

	removerCookies := []*http.Cookie{}
	for _, cookie := range c.Config.Remover {
		c := cookie.ToHttpCookie()
		removerCookies = append(removerCookies, &c)
	}

	if setCookieStr != "" {
		// Parse set cookies
		rawRequest := fmt.Sprintf("GET / HTTP/1.0\r\nCookie: %s\r\n\r\n", setCookieStr)
		fakeReq, _ := http.ReadRequest(bufio.NewReader(strings.NewReader(rawRequest)))
		setCookie = fakeReq.Cookies()[0]

	}

	// Merge cookies
	//mergedAddedCookies := MergeCookies(cookies, adderCookies)

	setCookie = MergeCookies([]*http.Cookie{setCookie}, adderCookies)[0]

	// Remove cookies
	//mergedRemovedCookies := MergeCookies(mergedAddedCookies, removerCookies)
	setCookie = MergeCookies([]*http.Cookie{setCookie}, removerCookies)[0]

	//req.Header.Del("Set-Cookie")
	//req.Header.Del("Cookie")

	// Set cookies
	//for _, cookie := range mergedRemovedCookies {
	//	req.AddCookie(cookie)
	//}

	// Set Set-Cookie header
	//req.Header.Set("Set-Cookie", setCookie.String())
	wrapped := &responseWriter{rw}
	// Call next handler
	c.next.ServeHTTP(wrapped, req)
}

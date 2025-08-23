package browserrequest

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/storage"
	"github.com/chromedp/chromedp"
)

type Session struct {
	debugURL  string
	timeout   time.Duration
	cookies   []*network.Cookie
	userAgent string
	sync.RWMutex
	lastFetch time.Time
	cacheTTL  time.Duration
}

// NewSession creates a new Session with the Chrome debug URL
func NewSession(debugURL string) *Session {
	if debugURL == "" {
		debugURL = "ws://localhost:9222"
	}

	return &Session{
		debugURL: debugURL,
		timeout:  30 * time.Second,
		cacheTTL: 5 * time.Minute,
	}
}

// WithTimeout sets the timeout for Chrome operations
func (s *Session) WithTimeout(timeout time.Duration) *Session {
	s.Lock()
	defer s.Unlock()
	s.timeout = timeout
	return s
}

// WithCacheTTL sets how long to cache cookies
func (s *Session) WithCacheTTL(ttl time.Duration) *Session {
	s.Lock()
	defer s.Unlock()
	s.cacheTTL = ttl
	return s
}

// SetRequest adds Chrome cookies to an http.Request for the matching domain (context-aware)
func (c *Session) SetRequest(ctx context.Context, req *http.Request) error {
	if req == nil {
		return fmt.Errorf("request is nil")
	}

	err := c.update(ctx)
	if err != nil {
		return fmt.Errorf("failed to get cookies: %w", err)
	}

	// Extract domain from request URL
	reqURL := req.URL
	if reqURL.Host == "" && req.Host != "" {
		reqURL.Host = req.Host
	}

	domain := reqURL.Hostname()
	if domain == "" {
		return fmt.Errorf("cannot determine domain from request")
	}

	// Build cookie string for matching domain
	cookieString := c.buildCookieString(c.cookies, domain)

	if cookieString != "" {
		// Append to existing cookies if any
		existingCookie := req.Header.Get("Cookie")
		if existingCookie != "" {
			cookieString = existingCookie + "; " + cookieString
		}
		req.Header.Set("Cookie", cookieString)
	}

	// log.Println(c.userAgent)
	// spew.Dump(c.cookies)
	req.Header.Set("user-agent", c.userAgent)

	return nil
}

// update returns all Chrome cookies (with caching)
func (c *Session) update(ctx context.Context) error {
	c.RLock()
	if time.Since(c.lastFetch) < c.cacheTTL && c.cookies != nil {
		c.RUnlock()
		return nil
	}
	c.RUnlock()

	// Need to refresh cookies
	c.Lock()
	defer c.Unlock()

	// Double-check after acquiring write lock
	if time.Since(c.lastFetch) < c.cacheTTL && c.cookies != nil {
		slog.Debug("browserrequest: cache still valid after write lock")
		return nil
	}

	actx, acancel := chromedp.NewRemoteAllocator(
		ctx,
		c.debugURL,
	)
	defer acancel()

	cookies, err := c.fetchCookies(actx)
	if err != nil {
		return err
	}

	agent, err := c.fetchUserAgent(actx)
	if err != nil {
		return err
	}

	c.userAgent = agent
	c.cookies = cookies
	c.lastFetch = time.Now()

	return nil
}

func (c *Session) fetchUserAgent(actx context.Context) (string, error) {
	// Get user agent from browser version info without creating a new tab
	// This avoids the browser window gaining focus
	cctx, ccancel := chromedp.NewContext(actx)
	defer ccancel()
	tctx, tcancel := context.WithTimeout(cctx, c.timeout)
	defer tcancel()

	browser, err := chromedp.FromContext(tctx).Allocator.Allocate(tctx)
	if err != nil {
		return "", fmt.Errorf("error connecting to browser: %w", err)
	}

	xctx := cdp.WithExecutor(tctx, browser)

	var version struct {
		UserAgent string `json:"userAgent"`
	}

	if err := cdp.Execute(xctx, "Browser.getVersion", nil, &version); err != nil {
		return "", err
	}

	return version.UserAgent, nil
}

// fetchCookies fetches cookies from Chrome (internal method)
func (c *Session) fetchCookies(actx context.Context) ([]*network.Cookie, error) {
	cctx, ccancel := chromedp.NewContext(actx)
	defer ccancel()

	tctx, tcancel := context.WithTimeout(cctx, c.timeout)
	defer tcancel()

	browser, err := chromedp.FromContext(tctx).Allocator.Allocate(tctx)
	if nil != err {
		log.Fatalf("Error connecting to browser: %s", err)
	}
	xctx := cdp.WithExecutor(tctx, browser)

	cookies, err := storage.GetCookies().Do(xctx)
	if nil != err {
		return nil, err
	}

	return cookies, nil
}

// buildCookieString builds a cookie string for HTTP headers
func (c *Session) buildCookieString(cookies []*network.Cookie, domain string) string {
	var cookieParts []string
	seen := make(map[string]bool)

	for _, cookie := range cookies {
		if matchesDomain(cookie.Domain, domain) {
			key := cookie.Name + "=" + cookie.Value
			if !seen[key] {
				cookieParts = append(cookieParts, key)
				seen[key] = true
			}
		}
	}

	return strings.Join(cookieParts, "; ")
}

// matchesDomain checks if a cookie domain matches the request domain
func matchesDomain(cookieDomain, requestDomain string) bool {
	// Remove leading dot from cookie domain if present
	cookieDomain = strings.TrimPrefix(cookieDomain, ".")

	// Exact match
	if cookieDomain == requestDomain {
		return true
	}

	// Subdomain match (cookie domain is parent of request domain)
	if strings.HasSuffix(requestDomain, "."+cookieDomain) {
		return true
	}

	// Check if request domain contains cookie domain
	if strings.Contains(requestDomain, cookieDomain) {
		return true
	}

	// Check if cookie domain contains request domain (for partial matches)
	if strings.Contains(cookieDomain, requestDomain) {
		return true
	}

	return false
}

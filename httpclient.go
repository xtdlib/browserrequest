package browserrequest

import (
	"context"
	"net/http"
)

type roundTripper struct {
	base          http.RoundTripper
	chromeSession *Session
	debug         bool
}

func (rt *roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// slog.Info("RoundTrip called", "url", req.URL.String())
	ctx := req.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	if err := rt.chromeSession.SetRequest(ctx, req); err != nil {
		return nil, err
	}

	return rt.base.RoundTrip(req)
}

// NewHTTPClient creates an http.Client that injects Chrome cookies
func NewHTTPClient(debugURL string) *http.Client {
	session := NewSession(debugURL)

	return &http.Client{
		Transport: &roundTripper{
			base:          http.DefaultTransport,
			chromeSession: session,
		},
	}
}

// NewHTTPClientWithTransport creates an http.Client with custom transport and Chrome cookies
func NewHTTPClientWithTransport(debugURL string, base http.RoundTripper) *http.Client {
	if base == nil {
		base = http.DefaultTransport
	}

	session := NewSession(debugURL)

	return &http.Client{
		Transport: &roundTripper{
			base:          base,
			chromeSession: session,
		},
	}
}

// NewHTTPClientWithSession creates an http.Client with a pre-configured Session
func NewHTTPClientWithSession(session *Session, base http.RoundTripper) *http.Client {
	if base == nil {
		base = http.DefaultTransport
	}

	return &http.Client{
		Transport: &roundTripper{
			base:          base,
			chromeSession: session,
		},
	}
}

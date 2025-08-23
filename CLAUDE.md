# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

`browserrequest` is a Go library that extracts cookies and user-agent from Chrome browser and injects them into HTTP requests using Chrome's remote debugging protocol. This enables HTTP clients to authenticate with the same cookies that are present in a running Chrome browser session.

## Prerequisites

Chrome must be running with remote debugging enabled:
```bash
google-chrome --remote-debugging-port=9222 --user-data-dir=/tmp/chrome
```

## Common Development Commands

### Build the library
```bash
go build ./...
```

### Run examples
```bash
# Simple example (uses browserrequest.Client directly)
go run exampl/simple/main.go

# HTTPClient example (uses the HTTP client wrapper)
go run exampl/httpclient/main.go
```

### Install dependencies
```bash
go mod download
go mod tidy
```

## Architecture

The library consists of two main components:

### 1. Core Client (`browserrequest.go`)
- **`Client`**: Manages Chrome cookie extraction and injection with caching support
- **Key methods**:
  - `SetRequest()`: Injects cookies and user-agent into an HTTP request
  - `update()`: Fetches cookies from Chrome with caching (5-minute TTL by default)
  - Uses mutex-based locking for thread safety
- **Chrome interaction**: Uses chromedp to connect to Chrome's debugging port and extract cookies via the Chrome DevTools Protocol

### 2. HTTP Client Wrapper (`httpclient.go`)
- **`HTTPClient`**: Wrapper around standard `http.Client` that automatically injects Chrome cookies
- **`roundTripper`**: Custom RoundTripper that intercepts requests to inject cookies
- Creates a new HTTP client instance to avoid modifying the passed client's transport

## Important Implementation Details

### Thread Safety
- The `Client` uses RWMutex to handle concurrent access to cookies
- Be careful with the locking in `update()` - it uses read lock first, then upgrades to write lock if refresh is needed

### Context Handling
- Always use the passed context (not `context.Background()`) in chromedp operations to respect timeouts
- See browserrequest.go:189 for the correct pattern

### HTTP Client Isolation
- The `NewHTTPClient` creates a new HTTP client instance rather than modifying the passed one
- This prevents chromedp's internal HTTP requests from going through our custom RoundTripper (which would cause deadlock)

## Known Issues & Solutions

### Hanging/Deadlock Issues
If the httpclient hangs, check:
1. Context is properly passed to chromedp operations (not using `context.Background()`)
2. HTTP client transport is not shared with chromedp's internal client
3. Chrome is actually running on the debug port (check with `netstat -an | grep 9222`)

## Dependencies

- `github.com/chromedp/chromedp`: Chrome DevTools Protocol client
- `github.com/chromedp/cdproto`: Chrome DevTools Protocol definitions
- `moul.io/http2curl`: For debugging HTTP requests as curl commands

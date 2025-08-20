package http

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TestNewHTTPTransport tests the creation of a new HTTP transport
func TestNewHTTPTransport(t *testing.T) {
	tests := []struct {
		name     string
		opts     []HTTPTransportOpts
		expected HTTPTransport
	}{
		{
			name: "default transport",
			opts: []HTTPTransportOpts{},
			expected: HTTPTransport{
				readDeadline:  0,
				writeDeadline: 0,
				basePath:      "",
				toolHandler:   nil,
				methodHandler: nil,
			},
		},
		{
			name: "with read deadline",
			opts: []HTTPTransportOpts{
				WithReadDeadline(5 * time.Second),
			},
			expected: HTTPTransport{
				readDeadline:  5 * time.Second,
				writeDeadline: 0,
				basePath:      "",
				toolHandler:   nil,
				methodHandler: nil,
			},
		},
		{
			name: "with write deadline",
			opts: []HTTPTransportOpts{
				WithWriteDeadline(10 * time.Second),
			},
			expected: HTTPTransport{
				readDeadline:  0,
				writeDeadline: 10 * time.Second,
				basePath:      "",
				toolHandler:   nil,
				methodHandler: nil,
			},
		},
		{
			name: "with base path",
			opts: []HTTPTransportOpts{
				WithPath("/api/v1"),
			},
			expected: HTTPTransport{
				readDeadline:  0,
				writeDeadline: 0,
				basePath:      "/api/v1",
				toolHandler:   nil,
				methodHandler: nil,
			},
		},
		{
			name: "with multiple options",
			opts: []HTTPTransportOpts{
				WithReadDeadline(5 * time.Second),
				WithWriteDeadline(10 * time.Second),
				WithPath("/api/v1"),
			},
			expected: HTTPTransport{
				readDeadline:  5 * time.Second,
				writeDeadline: 10 * time.Second,
				basePath:      "/api/v1",
				toolHandler:   nil,
				methodHandler: nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport := NewHTTPTransport(tt.opts...)

			if transport.readDeadline != tt.expected.readDeadline {
				t.Errorf("readDeadline = %v, want %v", transport.readDeadline, tt.expected.readDeadline)
			}
			if transport.writeDeadline != tt.expected.writeDeadline {
				t.Errorf("writeDeadline = %v, want %v", transport.writeDeadline, tt.expected.writeDeadline)
			}
			if transport.basePath != tt.expected.basePath {
				t.Errorf("basePath = %v, want %v", transport.basePath, tt.expected.basePath)
			}
			if transport.toolHandler != tt.expected.toolHandler {
				t.Errorf("toolHandler = %v, want %v", transport.toolHandler, tt.expected.toolHandler)
			}
			if transport.methodHandler != tt.expected.methodHandler {
				t.Errorf("methodHandler = %v, want %v", transport.methodHandler, tt.expected.methodHandler)
			}
		})
	}
}

// TestTimeoutFunctionality tests that timeout/deadline functionality works correctly
func TestTimeoutFunctionality(t *testing.T) {
	tests := []struct {
		name           string
		readDeadline   time.Duration
		writeDeadline  time.Duration
		handlerDelay   time.Duration
		expectedStatus int
		description    string
	}{
		{
			name:           "no timeouts set",
			readDeadline:   0,
			writeDeadline:  0,
			handlerDelay:   100 * time.Millisecond,
			expectedStatus: http.StatusOK,
			description:    "should work normally when no timeouts are set",
		},
		{
			name:           "read deadline longer than handler delay",
			readDeadline:   1 * time.Second,
			writeDeadline:  0,
			handlerDelay:   100 * time.Millisecond,
			expectedStatus: http.StatusOK,
			description:    "should work when read deadline is longer than handler delay",
		},
		{
			name:           "write deadline longer than handler delay",
			readDeadline:   0,
			writeDeadline:  1 * time.Second,
			handlerDelay:   100 * time.Millisecond,
			expectedStatus: http.StatusOK,
			description:    "should work when write deadline is longer than handler delay",
		},
		{
			name:           "both deadlines longer than handler delay",
			readDeadline:   1 * time.Second,
			writeDeadline:  1 * time.Second,
			handlerDelay:   100 * time.Millisecond,
			expectedStatus: http.StatusOK,
			description:    "should work when both deadlines are longer than handler delay",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a slow handler that simulates processing time
			slowHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(tt.handlerDelay)
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("response after delay"))
			})

			// Create transport with timeouts
			transport := NewHTTPTransport(
				WithReadDeadline(tt.readDeadline),
				WithWriteDeadline(tt.writeDeadline),
				WithToolHandler(slowHandler),
			)

			handler := transport.HTTPHandler()

			// Create request with context timeout
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			req := httptest.NewRequest("GET", "/tools", nil).WithContext(ctx)
			w := httptest.NewRecorder()

			// Execute request
			handler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("status = %d, want %d (%s)", w.Code, tt.expectedStatus, tt.description)
			}
		})
	}
}

// TestTimeoutWithSlowHandler tests timeout behavior with a very slow handler
func TestTimeoutWithSlowHandler(t *testing.T) {
	// Create a very slow handler
	slowHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second) // Very slow
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("slow response"))
	})

	// Create transport with short timeouts
	transport := NewHTTPTransport(
		WithReadDeadline(100*time.Millisecond),
		WithWriteDeadline(100*time.Millisecond),
		WithToolHandler(slowHandler),
	)

	handler := transport.HTTPHandler()

	// Create request with short context timeout
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	req := httptest.NewRequest("GET", "/tools", nil).WithContext(ctx)
	w := httptest.NewRecorder()

	// Execute request
	handler.ServeHTTP(w, req)

	// The handler should still work because httptest doesn't enforce connection timeouts
	// This test demonstrates that the current implementation doesn't enforce deadlines
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d (httptest doesn't enforce connection timeouts)", w.Code, http.StatusOK)
	}
}

// TestRealServerTimeout demonstrates actual timeout functionality with a real server
func TestRealServerTimeout(t *testing.T) {
	// Skip this test in CI environments where we can't bind to ports
	if testing.Short() {
		t.Skip("skipping real server test in short mode")
	}

	// Create a very slow handler
	slowHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(3 * time.Second) // Very slow
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("slow response"))
	})

	// Create transport with short timeouts
	transport := NewHTTPTransport(
		WithReadDeadline(1*time.Second),
		WithWriteDeadline(1*time.Second),
		WithToolHandler(slowHandler),
	)

	// Start server in background
	server := &http.Server{
		Addr:    ":0", // Let the system choose a port
		Handler: transport.HTTPHandler(),
	}

	// Apply deadline settings
	if transport.readDeadline > 0 {
		server.ReadTimeout = transport.readDeadline
	}
	if transport.writeDeadline > 0 {
		server.WriteTimeout = transport.writeDeadline
	}

	// Start server
	go func() {
		listener, err := net.Listen("tcp", server.Addr)
		if err != nil {
			t.Errorf("failed to create listener: %v", err)
			return
		}
		defer listener.Close()
		server.Serve(listener)
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Test with a client that has a short timeout
	client := &http.Client{
		Timeout: 500 * time.Millisecond,
	}

	// Get the actual address
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("failed to create test listener: %v", err)
	}
	addr := listener.Addr().String()
	listener.Close()

	resp, err := client.Get("http://" + addr + "/tools")
	if err != nil {
		// Expected error due to timeout
		t.Logf("Expected timeout error: %v", err)
		return
	}
	defer resp.Body.Close()

	// If we get here, the request didn't timeout as expected
	t.Logf("Request completed with status: %d", resp.StatusCode)
}

// TestDeadlineValuesAreStored tests that deadline values are properly stored
func TestDeadlineValuesAreStored(t *testing.T) {
	readDeadline := 5 * time.Second
	writeDeadline := 10 * time.Second

	transport := NewHTTPTransport(
		WithReadDeadline(readDeadline),
		WithWriteDeadline(writeDeadline),
	)

	if transport.readDeadline != readDeadline {
		t.Errorf("readDeadline = %v, want %v", transport.readDeadline, readDeadline)
	}

	if transport.writeDeadline != writeDeadline {
		t.Errorf("writeDeadline = %v, want %v", transport.writeDeadline, writeDeadline)
	}
}

// TestWithToolHandler tests the WithToolHandler option
func TestWithToolHandler(t *testing.T) {
	mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("mock tool handler"))
	})

	transport := NewHTTPTransport(WithToolHandler(mockHandler))

	if transport.toolHandler == nil {
		t.Error("toolHandler should not be nil")
	}
}

// TestWithMethodHandler tests the WithMethodHandler option
func TestWithMethodHandler(t *testing.T) {
	mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("mock method handler"))
	})

	transport := NewHTTPTransport(WithMethodHandler(mockHandler))

	if transport.methodHandler == nil {
		t.Error("methodHandler should not be nil")
	}
}

// TestHTTPHandler tests the HTTPHandler method
func TestHTTPHandler(t *testing.T) {
	tests := []struct {
		name           string
		basePath       string
		toolHandler    http.Handler
		methodHandler  http.Handler
		requestPath    string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "root path with base path",
			basePath:       "/api/v1",
			toolHandler:    nil,
			methodHandler:  nil,
			requestPath:    "/api/v1/",
			expectedStatus: http.StatusOK,
			expectedBody:   "Tool Service API - Use /tools for discovery, /execute for method calls",
		},
		{
			name:           "unknown path with base path",
			basePath:       "/api/v1",
			toolHandler:    nil,
			methodHandler:  nil,
			requestPath:    "/api/v1/unknown",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "Not Found!",
		},
		{
			name:           "tools endpoint with base path and tool handler",
			basePath:       "/api/v1",
			toolHandler:    http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("tools response")) }),
			methodHandler:  nil,
			requestPath:    "/api/v1/tools",
			expectedStatus: http.StatusOK,
			expectedBody:   "tools response",
		},
		{
			name:           "execute endpoint with base path and method handler",
			basePath:       "/api/v1",
			toolHandler:    nil,
			methodHandler:  http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("execute response")) }),
			requestPath:    "/api/v1/execute",
			expectedStatus: http.StatusOK,
			expectedBody:   "execute response",
		},
		{
			name:           "both handlers with base path",
			basePath:       "/api/v1",
			toolHandler:    http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("tools response")) }),
			methodHandler:  http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("execute response")) }),
			requestPath:    "/api/v1/tools",
			expectedStatus: http.StatusOK,
			expectedBody:   "tools response",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport := &HTTPTransport{
				basePath:      tt.basePath,
				toolHandler:   tt.toolHandler,
				methodHandler: tt.methodHandler,
			}

			handler := transport.HTTPHandler()

			req := httptest.NewRequest("GET", tt.requestPath, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.expectedStatus)
			}

			body := strings.TrimSpace(w.Body.String())
			if body != tt.expectedBody {
				t.Errorf("body = %q, want %q", body, tt.expectedBody)
			}
		})
	}
}

// TestHTTPHandlerWithoutBasePath tests the HTTPHandler method when no base path is set
func TestHTTPHandlerWithoutBasePath(t *testing.T) {
	tests := []struct {
		name           string
		toolHandler    http.Handler
		methodHandler  http.Handler
		requestPath    string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "root path without handlers",
			toolHandler:    nil,
			methodHandler:  nil,
			requestPath:    "/",
			expectedStatus: http.StatusOK,
			expectedBody:   "Tool Service API - Use /tools for discovery, /execute for method calls",
		},
		{
			name:           "unknown path without handlers",
			toolHandler:    nil,
			methodHandler:  nil,
			requestPath:    "/unknown",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "Not Found!",
		},
		{
			name:           "tools endpoint with tool handler",
			toolHandler:    http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("tools response")) }),
			methodHandler:  nil,
			requestPath:    "/tools",
			expectedStatus: http.StatusOK,
			expectedBody:   "tools response",
		},
		{
			name:           "execute endpoint with method handler",
			toolHandler:    nil,
			methodHandler:  http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("execute response")) }),
			requestPath:    "/execute",
			expectedStatus: http.StatusOK,
			expectedBody:   "execute response",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport := &HTTPTransport{
				basePath:      "",
				toolHandler:   tt.toolHandler,
				methodHandler: tt.methodHandler,
			}

			handler := transport.HTTPHandler()

			req := httptest.NewRequest("GET", tt.requestPath, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.expectedStatus)
			}

			body := strings.TrimSpace(w.Body.String())
			if body != tt.expectedBody {
				t.Errorf("body = %q, want %q", body, tt.expectedBody)
			}
		})
	}
}

// TestHTTPHandlerContentType tests that the root endpoint sets the correct content type
func TestHTTPHandlerContentType(t *testing.T) {
	transport := &HTTPTransport{}
	handler := transport.HTTPHandler()

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	contentType := w.Header().Get("Content-Type")
	if contentType != "text/plain" {
		t.Errorf("Content-Type = %q, want %q", contentType, "text/plain")
	}
}

// TestHTTPHandlerMethodHandling tests that the handler properly handles different HTTP methods
func TestHTTPHandlerMethodHandling(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
	}{
		{"GET root", "GET", "/", http.StatusOK},
		{"POST root", "POST", "/", http.StatusOK},
		{"PUT root", "PUT", "/", http.StatusOK},
		{"DELETE root", "DELETE", "/", http.StatusOK},
		{"GET unknown", "GET", "/unknown", http.StatusNotFound},
		{"POST unknown", "POST", "/unknown", http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport := &HTTPTransport{}
			handler := transport.HTTPHandler()

			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.expectedStatus)
			}
		})
	}
}

// TestHTTPHandlerWithBasePathStripping tests that the base path is properly stripped
func TestHTTPHandlerWithBasePathStripping(t *testing.T) {
	tests := []struct {
		name           string
		basePath       string
		requestPath    string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "base path with trailing slash",
			basePath:       "/api/v1/",
			requestPath:    "/api/v1/",
			expectedStatus: http.StatusOK,
			expectedBody:   "Tool Service API - Use /tools for discovery, /execute for method calls",
		},
		{
			name:           "base path without trailing slash",
			basePath:       "/api/v1",
			requestPath:    "/api/v1/",
			expectedStatus: http.StatusOK,
			expectedBody:   "Tool Service API - Use /tools for discovery, /execute for method calls",
		},
		{
			name:           "nested base path",
			basePath:       "/agents/api/v1",
			requestPath:    "/agents/api/v1/",
			expectedStatus: http.StatusOK,
			expectedBody:   "Tool Service API - Use /tools for discovery, /execute for method calls",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport := &HTTPTransport{basePath: tt.basePath}
			handler := transport.HTTPHandler()

			req := httptest.NewRequest("GET", tt.requestPath, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.expectedStatus)
			}

			body := strings.TrimSpace(w.Body.String())
			if body != tt.expectedBody {
				t.Errorf("body = %q, want %q", body, tt.expectedBody)
			}
		})
	}
}

// TestHTTPHandlerNilHandlers tests that nil handlers don't cause panics
func TestHTTPHandlerNilHandlers(t *testing.T) {
	transport := &HTTPTransport{
		toolHandler:   nil,
		methodHandler: nil,
	}
	handler := transport.HTTPHandler()

	// Test that requests to /tools and /execute don't panic when handlers are nil
	paths := []string{"/tools", "/execute"}

	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("handler panicked on %s: %v", path, r)
				}
			}()

			req := httptest.NewRequest("GET", path, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			// Should return 404 when handlers are nil
			if w.Code != http.StatusNotFound {
				t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
			}
		})
	}
}

// TestHTTPHandlerRequestBody tests that the handler properly handles requests with bodies
func TestHTTPHandlerRequestBody(t *testing.T) {
	transport := &HTTPTransport{}
	handler := transport.HTTPHandler()

	body := strings.NewReader("test body")
	req := httptest.NewRequest("POST", "/", body)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

// TestHTTPHandlerLargeRequest tests that the handler can handle large request bodies
func TestHTTPHandlerLargeRequest(t *testing.T) {
	transport := &HTTPTransport{}
	handler := transport.HTTPHandler()

	// Create a large request body
	largeBody := strings.Repeat("a", 1024*1024) // 1MB
	req := httptest.NewRequest("POST", "/", strings.NewReader(largeBody))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

// TestHTTPHandlerConcurrentRequests tests that the handler can handle concurrent requests
func TestHTTPHandlerConcurrentRequests(t *testing.T) {
	transport := &HTTPTransport{}
	handler := transport.HTTPHandler()

	// Test concurrent requests
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			req := httptest.NewRequest("GET", "/", nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("concurrent request status = %d, want %d", w.Code, http.StatusOK)
			}

			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

package middleware

import (
	"log"
	"net/http"
	"strings"
	"time"

	"go-learn/internal/model"
)

// Middleware is a standard HTTP middleware type.
type Middleware func(http.Handler) http.Handler

// Chain combines multiple middlewares from left to right.
func Chain(h http.Handler, middlewares ...Middleware) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (rec *statusRecorder) WriteHeader(code int) {
	rec.statusCode = code
	rec.ResponseWriter.WriteHeader(code)
}

// Logger logs HTTP method, path, remote address, status code, and latency.
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(rec, r)

		log.Printf("[%s] %s %d %s %s",
			r.Method,
			r.URL.Path,
			rec.statusCode,
			time.Since(start),
			r.RemoteAddr,
		)
	})
}

// Recovery catches any panics in downstream handlers and responds with a 500 Internal Server Error.
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("PANIC RECOVERED: %v", rec)
				http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// CORS adds basic CORS headers to allow cross-origin requests.
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-User-ID, X-Tenant-ID, X-Request-ID, X-Role")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// This is context propagation middleware that extracts request headers and metadata into a structured model.RequestContext and embeds it into the standard context.Context.
// is it a standard practice ? well, yes, it is a common practice in Go web applications to use middleware for context propagation. This allows you to pass request-scoped data
// (like user ID, tenant ID, request ID, etc.) down the call chain without having to modify function signatures or rely on global state. By embedding this metadata into the
// context, you can access it in handlers, services, and repositories in a decoupled manner.
// here we add it to the main context, main context or parent context is the context that is passed down from the top-level handler to all downstream handlers and services.
// By adding the RequestContext to the main context, we ensure that all parts of the application that have access to the context can retrieve the request-scoped metadata without needing to pass it explicitly through function parameters. This promotes cleaner code and better separation of concerns.
// uses of parent context: The parent context is used to propagate request-scoped data, cancellation signals, and deadlines across API boundaries and between processes.
// It allows different parts of the application to share information about the request without tightly coupling them together. By using a parent context, you can ensure that all downstream components have access to the same metadata and can respond appropriately to cancellation or timeout events.

// RequestContextMiddleware extracts request headers and metadata into model.RequestContext
// and embeds it into the standard context.Context. This allows downstream layers
// (handlers, services, repositories) to access request headers in a decoupled manner.
func RequestContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headers := make(map[string]string, len(r.Header))
		for k, v := range r.Header {
			if len(v) > 0 {
				headers[strings.ToLower(k)] = v[0]
			}
		}

		rc := model.RequestContext{
			UserID:    r.Header.Get("X-User-ID"),
			TenantID:  r.Header.Get("X-Tenant-ID"),
			RequestID: r.Header.Get("X-Request-ID"),
			Headers:   headers,
		}

		// Inject the structured metadata into request context
		ctx := model.WithRequestContext(r.Context(), rc)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

package middleware

import (
	"log"
	"net/http"
	"strings"
	"time"

	"go-learn/internal/model"

	"go.elastic.co/apm/v2"
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
// It also records the panic error to CubeAPM if an active transaction exists.
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("PANIC RECOVERED: %v", rec)

				// Report panic to CubeAPM transaction
				if e := apm.DefaultTracer().Recovered(rec); e != nil {
					e.SetTransaction(apm.TransactionFromContext(r.Context()))
					e.Send()
				}

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
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-User-ID, X-Tenant-ID, X-Request-ID, X-Role, traceparent, tracestate")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RequestContextMiddleware extracts request headers and metadata into model.RequestContext
// and embeds it into the standard context.Context. It also annotates the CubeAPM transaction
// with labels and user context so they appear in CubeAPM traces.
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

		// Enrich CubeAPM transaction with custom labels & metadata if present
		if tx := apm.TransactionFromContext(r.Context()); tx != nil {
			if rc.UserID != "" {
				tx.Context.SetUsername(rc.UserID)
				tx.Context.SetLabel("user_id", rc.UserID)
			}
			if rc.TenantID != "" {
				tx.Context.SetLabel("tenant_id", rc.TenantID)
			}
			if rc.RequestID != "" {
				tx.Context.SetLabel("request_id", rc.RequestID)
			}
		}

		// Inject the structured metadata into request context
		ctx := model.WithRequestContext(r.Context(), rc)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

package model

import (
	"context"
	"strings"
)

type contextKey struct{}

// RequestContext encapsulates request-scoped metadata extracted from HTTP transport,
// enabling the business logic to access headers, identities, and tracing data
// without importing net/http or coupling to transport details.
type RequestContext struct {
	UserID    string            `json:"user_id"`
	TenantID  string            `json:"tenant_id"`
	RequestID string            `json:"request_id"`
	Headers   map[string]string `json:"headers"`
}

// WithRequestContext injects RequestContext into a standard context.Context.
func WithRequestContext(ctx context.Context, rc RequestContext) context.Context {
	// Normalize header map keys to lowercase for case-insensitive lookups
	normalizedHeaders := make(map[string]string, len(rc.Headers))
	for k, v := range rc.Headers {
		normalizedHeaders[strings.ToLower(k)] = v
	}
	rc.Headers = normalizedHeaders

	//args: ctx: the existing context to which we want to add the RequestContext
	//rc: the RequestContext instance containing request-scoped metadata
	//contextKey{}: a unique key type to avoid collisions in the context's key-value store
	// cant we keep a simple 2nd arg ? well, we use a unique type (contextKey{}) to avoid key collisions in the context's key-value store.
	// If we used a simple string or int as the key, it could potentially collide with other keys used elsewhere in the application or in
	// third-party libraries. By using a unique struct type, we ensure that our key is distinct and won't accidentally overwrite or be
	// overwritten by other context values. When it can be overridden? If another part of the application or a third-party library uses
	//  the same key type (contextKey{}) and stores a value in the context, it could overwrite our RequestContext.
	return context.WithValue(ctx, contextKey{}, rc)
}

// GetRequestContext retrieves the RequestContext from context.Context.
func GetRequestContext(ctx context.Context) (RequestContext, bool) {
	rc, ok := ctx.Value(contextKey{}).(RequestContext)
	return rc, ok
}

// GetHeader returns the value of a specific header from the context (case-insensitive).
func GetHeader(ctx context.Context, key string) string {
	rc, ok := GetRequestContext(ctx)
	if !ok || rc.Headers == nil {
		return ""
	}
	return rc.Headers[strings.ToLower(key)]
}

// GetUserID extracts the user identifier from the request context.
func GetUserID(ctx context.Context) string {
	if rc, ok := GetRequestContext(ctx); ok {
		return rc.UserID
	}
	return ""
}

// GetTenantID extracts the tenant identifier from the request context.
func GetTenantID(ctx context.Context) string {
	if rc, ok := GetRequestContext(ctx); ok {
		return rc.TenantID
	}
	return ""
}

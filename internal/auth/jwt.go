// Package auth provides authentication and authorization primitives for
// the Stratavore HTTP API and gRPC server.
//
// Authentication is optional: when no secret is configured the middleware
// operates in allow-all mode so existing deployments are unaffected.
package auth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// ErrUnauthorized is returned when a request lacks valid credentials.
var ErrUnauthorized = errors.New("unauthorized")

// ErrTokenExpired is returned when a JWT has passed its expiry time.
var ErrTokenExpired = errors.New("token expired")

// Claims represents the payload embedded in a Stratavore JWT.
type Claims struct {
	Subject   string    `json:"sub"`
	IssuedAt  int64     `json:"iat"`
	ExpiresAt int64     `json:"exp"`
	Scope     []string  `json:"scope,omitempty"`
	ProjectID string    `json:"project_id,omitempty"`
	issued    time.Time // parsed from IssuedAt for convenience
}

// Valid reports whether the claims are currently valid.
func (c *Claims) Valid() error {
	if c.ExpiresAt != 0 && time.Now().Unix() > c.ExpiresAt {
		return ErrTokenExpired
	}
	return nil
}

// HasScope reports whether the claims include the requested scope.
func (c *Claims) HasScope(scope string) bool {
	for _, s := range c.Scope {
		if s == scope || s == "*" {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// Validator
// ---------------------------------------------------------------------------

// Validator verifies Stratavore API tokens.
// It uses a simple HS256-style HMAC scheme over JSON payloads (not a full
// JWT library dependency) so that the binary stays light.
type Validator struct {
	secret  []byte
	enabled bool
}

// NewValidator creates a Validator using the provided HMAC secret.
// If secret is empty the validator operates in pass-through mode.
func NewValidator(secret string) *Validator {
	return &Validator{
		secret:  []byte(secret),
		enabled: secret != "",
	}
}

// Enabled reports whether authentication is enforced.
func (v *Validator) Enabled() bool { return v.enabled }

// Generate creates a signed token string for the given claims.
func (v *Validator) Generate(claims Claims) (string, error) {
	if !v.enabled {
		return "", errors.New("auth: cannot generate token: no secret configured")
	}
	now := time.Now()
	claims.IssuedAt = now.Unix()
	if claims.ExpiresAt == 0 {
		claims.ExpiresAt = now.Add(24 * time.Hour).Unix()
	}

	payload, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("auth: marshal claims: %w", err)
	}

	b64 := base64.RawURLEncoding.EncodeToString(payload)
	sig := v.sign(b64)
	return b64 + "." + sig, nil
}

// Validate parses and verifies a token, returning the embedded Claims.
func (v *Validator) Validate(token string) (*Claims, error) {
	if !v.enabled {
		// Pass-through: return synthetic superuser claims.
		return &Claims{Subject: "anonymous", Scope: []string{"*"}}, nil
	}

	parts := strings.SplitN(token, ".", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("%w: malformed token", ErrUnauthorized)
	}

	b64, sig := parts[0], parts[1]
	if expected := v.sign(b64); !hmac.Equal([]byte(sig), []byte(expected)) {
		return nil, fmt.Errorf("%w: invalid signature", ErrUnauthorized)
	}

	raw, err := base64.RawURLEncoding.DecodeString(b64)
	if err != nil {
		return nil, fmt.Errorf("%w: base64 decode: %v", ErrUnauthorized, err)
	}

	var claims Claims
	if err := json.Unmarshal(raw, &claims); err != nil {
		return nil, fmt.Errorf("%w: unmarshal: %v", ErrUnauthorized, err)
	}

	if err := claims.Valid(); err != nil {
		return nil, err
	}

	return &claims, nil
}

func (v *Validator) sign(payload string) string {
	mac := hmac.New(sha256.New, v.secret)
	mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

// ---------------------------------------------------------------------------
// HTTP Middleware
// ---------------------------------------------------------------------------

type contextKey string

const claimsContextKey contextKey = "auth_claims"

// Middleware returns an HTTP middleware that validates Bearer tokens.
// If auth is disabled (no secret) it calls next unconditionally.
func Middleware(v *Validator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !v.enabled {
				next.ServeHTTP(w, r)
				return
			}

			// Allow health + metrics endpoints unauthenticated
			if r.URL.Path == "/health" || strings.HasPrefix(r.URL.Path, "/metrics") {
				next.ServeHTTP(w, r)
				return
			}

			token := extractBearerToken(r)
			if token == "" {
				// Also accept X-API-Key header for CLI convenience
				token = r.Header.Get("X-API-Key")
			}
			if token == "" {
				http.Error(w, `{"error":"missing authorization"}`, http.StatusUnauthorized)
				return
			}

			claims, err := v.Validate(token)
			if err != nil {
				http.Error(w, fmt.Sprintf(`{"error":%q}`, err.Error()), http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), claimsContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ClaimsFromContext retrieves Claims stored by Middleware.
func ClaimsFromContext(ctx context.Context) (*Claims, bool) {
	c, ok := ctx.Value(claimsContextKey).(*Claims)
	return c, ok
}

func extractBearerToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	return ""
}

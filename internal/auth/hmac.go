package auth

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

// ReplaySafeWindow is the maximum age of an HMAC-signed request that the
// daemon will accept. Requests older than this are rejected as potential
// replays.
const ReplaySafeWindow = 5 * time.Minute

// SignRequest attaches HMAC-SHA256 authentication headers to an outgoing
// HTTP request. The caller must have already set the request body (if any)
// before calling SignRequest.
//
// Headers added:
//
//	X-Stratavore-Timestamp  – Unix seconds of signing time
//	X-Stratavore-Signature  – hex(HMAC-SHA256(secret, method+"\n"+path+"\n"+ts+"\n"+body))
func SignRequest(req *http.Request, secret string) error {
	if secret == "" {
		return nil // signing disabled
	}

	ts := strconv.FormatInt(time.Now().Unix(), 10)

	// Read body for signing without consuming it
	var bodyBytes []byte
	if req.Body != nil && req.Body != http.NoBody {
		var err error
		bodyBytes, err = io.ReadAll(req.Body)
		if err != nil {
			return fmt.Errorf("hmac: read body: %w", err)
		}
		req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	}

	sig := computeSignature(secret, req.Method, req.URL.RequestURI(), ts, bodyBytes)

	req.Header.Set("X-Stratavore-Timestamp", ts)
	req.Header.Set("X-Stratavore-Signature", sig)
	return nil
}

// VerifyRequest validates the HMAC signature on an incoming request.
// Returns nil on success; returns a descriptive error on failure.
// If secret is empty the function always returns nil (verification disabled).
func VerifyRequest(req *http.Request, secret string) error {
	if secret == "" {
		return nil
	}

	tsHeader := req.Header.Get("X-Stratavore-Timestamp")
	sigHeader := req.Header.Get("X-Stratavore-Signature")

	if tsHeader == "" || sigHeader == "" {
		return fmt.Errorf("%w: missing HMAC headers", ErrUnauthorized)
	}

	// Check timestamp to prevent replay attacks
	ts, err := strconv.ParseInt(tsHeader, 10, 64)
	if err != nil {
		return fmt.Errorf("%w: invalid timestamp", ErrUnauthorized)
	}
	age := time.Since(time.Unix(ts, 0))
	if age > ReplaySafeWindow || age < -ReplaySafeWindow {
		return fmt.Errorf("%w: timestamp outside replay-safe window (age=%s)", ErrUnauthorized, age)
	}

	// Read body for verification without consuming it
	var bodyBytes []byte
	if req.Body != nil && req.Body != http.NoBody {
		bodyBytes, err = io.ReadAll(req.Body)
		if err != nil {
			return fmt.Errorf("hmac: read body: %w", err)
		}
		req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	}

	expected := computeSignature(secret, req.Method, req.URL.RequestURI(), tsHeader, bodyBytes)
	if !hmac.Equal([]byte(sigHeader), []byte(expected)) {
		return fmt.Errorf("%w: signature mismatch", ErrUnauthorized)
	}
	return nil
}

// HMACMiddleware returns an HTTP middleware that verifies HMAC request
// signatures. If secret is empty the middleware is a no-op pass-through.
func HMACMiddleware(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if secret == "" {
				next.ServeHTTP(w, r)
				return
			}
			if err := VerifyRequest(r, secret); err != nil {
				http.Error(w, fmt.Sprintf(`{"error":%q}`, err.Error()), http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// computeSignature returns hex(HMAC-SHA256(secret, payload)) where payload is:
//
//	METHOD\nPATH\nTIMESTAMP\nBODY
func computeSignature(secret, method, path, ts string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(method))
	mac.Write([]byte("\n"))
	mac.Write([]byte(path))
	mac.Write([]byte("\n"))
	mac.Write([]byte(ts))
	mac.Write([]byte("\n"))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}

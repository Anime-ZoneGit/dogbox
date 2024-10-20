package main

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

const MAX_API_KEY_LENGTH = 64

// Extracts the token from the given header in the request.
func extractToken(c *gin.Context, header string) (string, error) {
	authString := c.GetHeader(header)
	tokens := strings.SplitN(authString, " ", 2)
	if len(tokens) < 2 {
		return "", errors.New("Invalid token")
	}

	token := strings.TrimSpace(tokens[1])
	if len(token) > MAX_API_KEY_LENGTH {
		return "", errors.New("Invalid token")
	}

	return strings.TrimSpace(token), nil
}

// Verifies the raw API key by hashing it with the SHA-256 algorithm and
// doing a constant-time comparison against the server's API key. A
// constant-time comparison is necessary to prevent timing attacks: an
// adversary could guess the key by measuring how long it takes for a naive
// equality algorithm to return.
func verifyApiKey(cfg *Config, s string) bool {
	hash := sha256.Sum256([]byte(s))
	key := hash[:]

	return subtle.ConstantTimeCompare(cfg.DecodedAPIKey, key) == 1
}

func ApiKeyMiddleware(cfg *Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		key, err := extractToken(c, "Authorization")
		if err != nil {
			c.JSON(
				http.StatusUnauthorized,
				gin.H{"msg": "Could not find API key"},
			)
			c.Abort()
			return
		}

		if !verifyApiKey(cfg, key) {
			c.JSON(
				http.StatusUnauthorized,
				gin.H{"msg": "Invalid authentication"},
			)
			c.Abort()
			return
		}

		c.Next()
	}
}

func RateLimiter(r rate.Limit, b int) gin.HandlerFunc {
	limiter := rate.NewLimiter(r, b)
	return func(c *gin.Context) {
		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{"msg": "Rate Limit Exceeded"})
			c.Abort()
			return
		} else {
			c.Next()
		}
	}
}

// FIXME: broken
func Timeout(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		defer func() {
			if ctx.Err() == context.DeadlineExceeded {
				c.JSON(http.StatusGatewayTimeout, gin.H{"msg": "Timeout"})
				c.Abort()
			}

			cancel()
		}()

		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func Delay(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		time.Sleep(timeout)
		c.Next()
	}
}

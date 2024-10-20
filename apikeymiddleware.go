package main

import (
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Extracts the token from the given header in the request.
func extractToken(c *gin.Context, header string) (string, error) {
	authString := c.GetHeader(header)
	tokens := strings.SplitN(authString, " ", 2)
	if len(tokens) < 2 {
		return "", errors.New("Token with incorrect bearer format")
	}

	return strings.TrimSpace(tokens[1]), nil
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

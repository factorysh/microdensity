package claims

import (
	"context"
	"errors"
	"fmt"

	"github.com/cristalhq/jwt/v3"
)

type contextKey string

var (
	claimsKey = contextKey("claims")
)

// Claims represents all the data found in JWT Claims
type Claims struct {
	jwt.StandardClaims
	Owner string `json:"owner"`
	Admin bool   `json:"admin"`
	Path  string `json:"path"`
}

// Validate data owner struct
func (c *Claims) Validate() error {
	// FIXME: add standard claims check
	if c.Owner == "" {
		return fmt.Errorf("invalid owner name")
	}

	return nil
}

// ToCtx creates a context containing a user key
func (c *Claims) ToCtx(in context.Context) context.Context {
	return context.WithValue(in, claimsKey, *c)
}

// FromCtx extract a user from a context
func FromCtx(ctx context.Context) (*Claims, error) {
	u, ok := ctx.Value(claimsKey).(Claims)
	if !ok {
		return nil, errors.New("No user in this context")
	}

	return &u, nil
}

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

/*
{
  "namespace_id": "23",
  "namespace_path": "factory",
  "project_id": "605",
  "project_path": "factory/check-my-web",
  "user_id": "4",
  "user_login": "mlecarme",
  "user_email": "mlecarme@bearstech.com",
  "pipeline_id": "20364",
  "pipeline_source": "push",
  "job_id": "106045",
  "ref": "main",
  "ref_type": "branch",
  "ref_protected": "true",
  "jti": "fdc1d726-4238-4830-9a3f-178f9d2ba6b0",
  "iss": "gitlab.bearstech.com",
  "iat": 1643491670,
  "nbf": 1643491665,
  "exp": 1643495270,
  "sub": "job_106045"
}

*/
// Claims represents all the data found in JWT Claims
type Claims struct {
	jwt.StandardClaims
	NamespaceID   string `json:"namespace_id"`
	NamespacePath string `json:"namespace_path"`
	ProjectID     string `json:"project_id"`
	ProjectPath   string `json:"project_path"`
	UserID        string `json:"user_id"`
	UserLogin     string `json:"user_login"`
	UserEmail     string `json:"user_email"`
	PipelineID    string `json:"pipeline_id"`
	JobID         string `json:"job_id"`
	Ref           string `json:"ref"`
	RefType       string `json:"ref_type"`
	RefProtected  string `json:"ref_protected"`
}

// Validate data owner struct
func (c *Claims) Validate() error {
	// FIXME: add standard claims check
	if c.UserLogin == "" {
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

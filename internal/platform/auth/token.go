package auth

import (
	"errors"
	"strings"
)

var ErrInvalidAuthHeader = errors.New("invalid authorization header")

type TokenValidator interface {
	Validate(authorizationHeader string) (Principal, error)
}

type LocalTokenValidator struct {
	tokenPrefix string
}

func NewLocalTokenValidator(tokenPrefix string) LocalTokenValidator {
	return LocalTokenValidator{tokenPrefix: tokenPrefix}
}

func (v LocalTokenValidator) Validate(authorizationHeader string) (Principal, error) {
	if strings.TrimSpace(authorizationHeader) == "" {
		return AnonymousPrincipal(), nil
	}

	if !strings.HasPrefix(authorizationHeader, "Bearer ") {
		return AnonymousPrincipal(), ErrInvalidAuthHeader
	}

	rawToken := strings.TrimSpace(strings.TrimPrefix(authorizationHeader, "Bearer "))
	parts := strings.Split(rawToken, ":")
	if len(parts) < 2 {
		return AnonymousPrincipal(), ErrInvalidAuthHeader
	}

	if parts[0] != v.tokenPrefix {
		return AnonymousPrincipal(), ErrInvalidAuthHeader
	}

	subject := strings.TrimSpace(parts[1])
	if subject == "" {
		return AnonymousPrincipal(), ErrInvalidAuthHeader
	}

	roles := []string{}
	if len(parts) >= 3 && strings.TrimSpace(parts[2]) != "" {
		for _, role := range strings.Split(parts[2], ",") {
			trimmed := strings.TrimSpace(role)
			if trimmed != "" {
				roles = append(roles, trimmed)
			}
		}
	}

	return NewPrincipal(subject, roles), nil
}

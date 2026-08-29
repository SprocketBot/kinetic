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
	prefix := v.tokenPrefix + ":"
	if !strings.HasPrefix(rawToken, prefix) {
		return AnonymousPrincipal(), ErrInvalidAuthHeader
	}

	remainder := strings.TrimPrefix(rawToken, prefix)
	subjectPart := remainder
	rolesPart := ""
	if separator := strings.LastIndex(remainder, ":"); separator >= 0 {
		subjectPart = remainder[:separator]
		rolesPart = remainder[separator+1:]
	}

	subject := strings.TrimSpace(subjectPart)
	if subject == "" {
		return AnonymousPrincipal(), ErrInvalidAuthHeader
	}

	roles := []string{}
	if strings.TrimSpace(rolesPart) != "" {
		for _, role := range strings.Split(rolesPart, ",") {
			trimmed := strings.TrimSpace(role)
			if trimmed != "" {
				roles = append(roles, trimmed)
			}
		}
	}

	return NewPrincipal(subject, roles), nil
}

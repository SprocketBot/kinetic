package auth

import "context"

type Principal struct {
	Subject   string
	Roles     []string
	Anonymous bool
}

func AnonymousPrincipal() Principal {
	return Principal{
		Anonymous: true,
		Roles:     []string{},
	}
}

func NewPrincipal(subject string, roles []string) Principal {
	return Principal{
		Subject:   subject,
		Roles:     roles,
		Anonymous: false,
	}
}

type principalContextKey struct{}

func WithPrincipal(ctx context.Context, principal Principal) context.Context {
	return context.WithValue(ctx, principalContextKey{}, principal)
}

func PrincipalFromContext(ctx context.Context) Principal {
	principal, ok := ctx.Value(principalContextKey{}).(Principal)
	if !ok {
		return AnonymousPrincipal()
	}
	return principal
}

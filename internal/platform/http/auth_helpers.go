package http

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/sprocketbot/sprocket-v3/internal/domain/apitoken"
	"github.com/sprocketbot/sprocket-v3/internal/domain/authz"
	"github.com/sprocketbot/sprocket-v3/internal/platform/auth"
)

func parseSessionTTL(raw string) time.Duration {
	parsed, err := time.ParseDuration(strings.TrimSpace(raw))
	if err != nil || parsed <= 0 {
		return 12 * time.Hour
	}
	return parsed
}

func readSessionPrincipal(r *http.Request, cookieName, secret string) (auth.SessionPrincipal, bool) {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		return auth.SessionPrincipal{}, false
	}
	principal, err := auth.VerifySessionToken(cookie.Value, secret, time.Now().UTC())
	if err != nil {
		return auth.SessionPrincipal{}, false
	}
	return principal, true
}

// readPrincipal returns an auth.SessionPrincipal for the request. It first
// checks the session cookie; if none is present it falls back to an
// Authorization: Bearer token validated via the apitoken store. Session cookie
// auth takes priority.
func readPrincipal(r *http.Request, cookieName, secret string, tokenStore apitoken.Store) (auth.SessionPrincipal, bool) {
	if p, ok := readSessionPrincipal(r, cookieName, secret); ok {
		return p, true
	}
	if tokenStore == nil {
		return auth.SessionPrincipal{}, false
	}
	authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return auth.SessionPrincipal{}, false
	}
	rawToken := strings.TrimPrefix(authHeader, "Bearer ")
	rawToken = strings.TrimSpace(rawToken)
	if rawToken == "" {
		return auth.SessionPrincipal{}, false
	}
	result, err := tokenStore.ValidateAPIToken(r.Context(), rawToken)
	if err != nil {
		return auth.SessionPrincipal{}, false
	}
	return auth.SessionPrincipal{
		Subject:     result.Subject,
		DisplayName: result.Subject,
		Roles:       result.Scopes,
		ExpiresAt:   time.Time{},
	}, true
}

func buildAuthCallbackQuery(values url.Values, webBaseURL string) string {
	subject := strings.TrimSpace(values.Get("subject"))
	if subject == "" {
		subject = "local-player"
	}
	displayName := strings.TrimSpace(values.Get("displayName"))
	roles := strings.TrimSpace(values.Get("roles"))
	if roles == "" {
		roles = "player"
	}
	redirect := normalizeRedirectURL(values.Get("redirect"), webBaseURL)
	query := url.Values{}
	query.Set("subject", subject)
	query.Set("displayName", firstNonEmpty(displayName, subject))
	query.Set("roles", roles)
	query.Set("redirect", redirect)
	return query.Encode()
}

func normalizeRedirectURL(raw, webBaseURL string) string {
	defaultRedirect := strings.TrimRight(webBaseURL, "/") + "/app"
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return defaultRedirect
	}
	if strings.HasPrefix(trimmed, "/") {
		return strings.TrimRight(webBaseURL, "/") + trimmed
	}

	targetURL, err := url.Parse(trimmed)
	if err != nil || targetURL.Scheme == "" || targetURL.Host == "" {
		return defaultRedirect
	}

	allowedURL, err := url.Parse(webBaseURL)
	if err != nil || !strings.EqualFold(allowedURL.Host, targetURL.Host) {
		return defaultRedirect
	}
	return targetURL.String()
}

func parseRoleList(raw string) []string {
	roles := make([]string, 0)
	for _, role := range strings.Split(raw, ",") {
		trimmed := strings.TrimSpace(role)
		if trimmed != "" {
			roles = append(roles, trimmed)
		}
	}
	if len(roles) == 0 {
		return []string{"player"}
	}
	return roles
}

// checkPermission reads the requesting principal and checks whether they hold
// the required resource+action permission. Auth is resolved in order:
//  1. Session cookie
//  2. Bearer token validated by the TokenValidator (local/JWT format)
//  3. Bearer token validated by the APITokenStore (spr_ machine tokens)
//
// It writes an appropriate HTTP error and returns false if the check fails.
// Callers should return immediately when this returns false.
func checkPermission(
	w http.ResponseWriter,
	r *http.Request,
	cookieName, secret string,
	tokenValidator auth.TokenValidator,
	tokenStore apitoken.Store,
	ev authz.Evaluator,
	resource, action string,
) bool {
	var roles []string

	if sp, ok := readSessionPrincipal(r, cookieName, secret); ok {
		roles = sp.Roles
	} else if tokenValidator != nil {
		authHeader := r.Header.Get("Authorization")
		if principal, err := tokenValidator.Validate(authHeader); err == nil && !principal.Anonymous {
			roles = principal.Roles
		} else if tokenStore != nil {
			sp, ok := readAPIBearerPrincipal(r, tokenStore)
			if !ok {
				http.Error(w, "authentication required", http.StatusUnauthorized)
				return false
			}
			roles = sp.Roles
		} else {
			http.Error(w, "authentication required", http.StatusUnauthorized)
			return false
		}
	} else if tokenStore != nil {
		sp, ok := readAPIBearerPrincipal(r, tokenStore)
		if !ok {
			http.Error(w, "authentication required", http.StatusUnauthorized)
			return false
		}
		roles = sp.Roles
	} else {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return false
	}

	if ev == nil {
		// No evaluator configured — allow by default (development mode).
		return true
	}
	if !ev.AllowedInContext(roles, nil, resource, action, authz.GlobalContext()) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return false
	}
	return true
}

// readAPIBearerPrincipal extracts a principal from an Authorization: Bearer
// header using the APITokenStore (for spr_* machine tokens).
func readAPIBearerPrincipal(r *http.Request, tokenStore apitoken.Store) (auth.SessionPrincipal, bool) {
	authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return auth.SessionPrincipal{}, false
	}
	rawToken := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
	if rawToken == "" {
		return auth.SessionPrincipal{}, false
	}
	result, err := tokenStore.ValidateAPIToken(r.Context(), rawToken)
	if err != nil {
		return auth.SessionPrincipal{}, false
	}
	return auth.SessionPrincipal{
		Subject:     result.Subject,
		DisplayName: result.Subject,
		Roles:       result.Scopes,
		ExpiresAt:   time.Time{},
	}, true
}

const oauthStateCookieName = "oauth_state"

// generateOAuthState returns a cryptographically random hex string suitable
// for use as an OAuth2 state parameter.
func generateOAuthState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// setOAuthStateCookie writes a short-lived HttpOnly cookie containing the
// OAuth state value. The cookie expires in 10 minutes.
func setOAuthStateCookie(w http.ResponseWriter, state string) {
	http.SetCookie(w, &http.Cookie{
		Name:     oauthStateCookieName,
		Value:    state,
		Path:     "/",
		MaxAge:   600,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

// validateOAuthStateCookie checks that the provided state value matches the
// oauth_state cookie on the request. It returns true only when the cookie is
// present and the values match.
func validateOAuthStateCookie(r *http.Request, providedState string) bool {
	if strings.TrimSpace(providedState) == "" {
		return false
	}
	cookie, err := r.Cookie(oauthStateCookieName)
	if err != nil {
		return false
	}
	return cookie.Value == providedState
}

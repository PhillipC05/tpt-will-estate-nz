package realme

import (
	"context"
	"net/http"
	"net/url"
)

// RequireLogin returns a Chi/net/http middleware that ensures the request has
// an authenticated RealMe session (any assurance level).
// Unauthenticated requests are redirected to /auth/realme/login.
func (p *Provider) RequireLogin() func(http.Handler) http.Handler {
	return p.requireAssuranceLevel(LevelLogin, "/auth/realme/login")
}

// RequireVerified returns a Chi/net/http middleware that ensures the request
// has a RealMe Verified Identity session.
// Requests with only a basic login are redirected to /auth/realme/verified
// (the Assertion Service login endpoint).
// Unauthenticated requests are redirected to /auth/realme/login.
func (p *Provider) RequireVerified() func(http.Handler) http.Handler {
	return p.requireAssuranceLevel(LevelVerified, "/auth/realme/verified")
}

// RequireAssuranceLevel returns middleware enforcing a specific minimum level.
func (p *Provider) RequireAssuranceLevel(level AssuranceLevel) func(http.Handler) http.Handler {
	loginURL := "/auth/realme/login"
	if level >= LevelVerified {
		loginURL = "/auth/realme/verified"
	}
	return p.requireAssuranceLevel(level, loginURL)
}

func (p *Provider) requireAssuranceLevel(required AssuranceLevel, redirectTo string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, err := p.store.GetSession(r)
			if err != nil || session == nil {
				// Not authenticated at all — redirect to login.
				p.redirectToLogin(w, r, "/auth/realme/login")
				return
			}

			identity, ok := session.(*Identity)
			if !ok {
				p.redirectToLogin(w, r, "/auth/realme/login")
				return
			}

			if identity.AssuranceLevel < required {
				// Authenticated but insufficient assurance — redirect to higher assurance.
				p.redirectToLogin(w, r, redirectTo)
				return
			}

			// Inject identity into request context.
			ctx := context.WithValue(r.Context(), identityContextKey, identity)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func (p *Provider) redirectToLogin(w http.ResponseWriter, r *http.Request, loginPath string) {
	returnURL := r.URL.RequestURI()
	target := loginPath + "?return=" + url.QueryEscape(returnURL)
	http.Redirect(w, r, target, http.StatusFound)
}

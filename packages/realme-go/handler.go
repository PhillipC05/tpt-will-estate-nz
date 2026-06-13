package realme

import (
	"encoding/xml"
	"net/http"

	"github.com/crewjam/saml/samlsp"
)

// OnSuccessFunc is called after a successful RealMe authentication.
// The Identity is guaranteed to be non-nil. Use it to create your own
// application user record before redirecting.
type OnSuccessFunc func(w http.ResponseWriter, r *http.Request, identity *Identity)

// LoginHandler returns an http.HandlerFunc that initiates the RealMe SAML SSO
// flow by redirecting the user to the RealMe IdP.
func (p *Provider) LoginHandler() http.HandlerFunc {
	mw := p.middleware()
	return func(w http.ResponseWriter, r *http.Request) {
		// If already authenticated, redirect to the return URL or /.
		if session, err := p.store.GetSession(r); err == nil && session != nil {
			returnURL := r.URL.Query().Get("return")
			if returnURL == "" {
				returnURL = "/"
			}
			http.Redirect(w, r, returnURL, http.StatusFound)
			return
		}
		mw.HandleStartAuthFlow(w, r)
	}
}

// CallbackHandler returns an http.HandlerFunc that handles the RealMe SAML
// POST-binding callback (ACS URL). On success it calls onSuccess, then redirects
// to "/" or the relay state URL.
//
// onSuccess may be nil — in that case the handler simply sets the session cookie
// and redirects. Pass a non-nil onSuccess to create your application user record
// (e.g., upsert into your database) before the redirect.
func (p *Provider) CallbackHandler(onSuccess OnSuccessFunc) http.HandlerFunc {
	mw := p.middleware()
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		assertion, err := mw.ServiceProvider.ParseResponse(r, []string{})
		if err != nil {
			samlsp.DefaultOnError(w, r, err)
			return
		}

		if err := p.store.CreateSession(w, r, assertion); err != nil {
			http.Error(w, "session error", http.StatusInternalServerError)
			return
		}

		if onSuccess != nil {
			identity, _ := extractIdentity(assertion)
			if identity != nil {
				onSuccess(w, r, identity)
				return
			}
		}

		returnURL := r.Form.Get("RelayState")
		if returnURL == "" {
			returnURL = "/"
		}
		http.Redirect(w, r, returnURL, http.StatusFound)
	}
}

// LogoutHandler returns an http.HandlerFunc that clears the RealMe session cookie
// and redirects to "/" (or a provided `return` query parameter URL).
func (p *Provider) LogoutHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_ = p.store.DeleteSession(w, r)
		returnURL := r.URL.Query().Get("return")
		if returnURL == "" {
			returnURL = "/"
		}
		http.Redirect(w, r, returnURL, http.StatusFound)
	}
}

// MetadataHandler returns an http.HandlerFunc that serves the SP's SAML metadata
// XML. Register this at your EntityID path (e.g. GET /saml/metadata).
// Download the output and submit it to DIA when registering your service.
func (p *Provider) MetadataHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		meta := p.sp.Metadata()
		w.Header().Set("Content-Type", "application/xml; charset=utf-8")
		enc := xml.NewEncoder(w)
		enc.Indent("", "  ")
		if err := enc.Encode(meta); err != nil {
			http.Error(w, "metadata error", http.StatusInternalServerError)
		}
	}
}

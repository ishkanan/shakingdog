package auth

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"net/url"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
)

// we request these scopes from Okta
// they will be in the returned token
var desiredScopes = [...]string{
	"openid",
	"email",
	"groups",
}

const (
	sessionName  = "auth"
	stateCookie  = "state"
	userCookie   = "user"
	groupsCookie = "groups"
	codeQueryKey = "code"

	oktaAuthPath  = "/oauth2/v1/authorize"
	oktaTokenPath = "/oauth2/v1/token"
)

// Okta provides methods for creating authentication handlers and wrapping
// normal handlers to ensure there is an authenticated user present.
type Okta struct {
	oktaDomain *url.URL
	store      sessions.Store
	authCfg    *oauth2.Config
}

// NewOktaAuth creates a new authentication helper configured for Okta auth
//
// domain, clientID and clientSecret should come straight from your Okta org and
// the configured application.
// appBaseURL and appAuthCallbackPath allow correct URLs to be built for
// the full URL redirects to/from Okta
func NewOktaAuth(domain, clientID, clientSecret, appBaseURL, appAuthCallbackPath string) *Okta {
	// generate OAuth2 config
	authCfg := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       []string{"openid", "email", "profile", "groups", "address"},
		RedirectURL:  appBaseURL + appAuthCallbackPath,
		Endpoint: oauth2.Endpoint{
			AuthURL:  domain + oktaAuthPath,
			TokenURL: domain + oktaTokenPath,
		},
	}

	// generate a new Cookie key
	// this means users will have to log back in after a server restart
	signingKey := securecookie.GenerateRandomKey(32)
	encryptKey := securecookie.GenerateRandomKey(16)

	oktaBaseURL, err := url.Parse(domain)
	if err != nil {
		panic(err)
	}

	return &Okta{
		oktaDomain: oktaBaseURL,
		store:      sessions.NewCookieStore(signingKey, encryptKey),
		authCfg:    authCfg,
	}
}

// generates a new random value to use for CSRF protection
func randomState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

// LoginHandler returns a handler that sends the user, with a state
// cookie, to Okta for auth
func (o *Okta) LoginHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// open the encrypted cookie-based session (it's created if it doesn't exist)
		session, err := o.store.New(r, sessionName)
		if session == nil && err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// generate a non-guessable value for CSRF protection
		state := randomState()
		// store this so that the callback handler can check it later
		session.Values[stateCookie] = state
		session.Save(r, w)

		// do the redirect
		http.Redirect(w, r, o.authCfg.AuthCodeURL(state, oauth2.AccessTypeOnline), http.StatusFound)
	})
}

// AuthCallbackHandler returns a http.Handler that will process an OAuth2
// authentication callback. Saving the needed user information into a secure
// cookie based session
func (o *Okta) AuthCallbackHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		session, err := o.store.New(r, sessionName)
		if session == nil && err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// parse the form, erroring if need be
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// check the state cookie matches the query string
		state := r.FormValue("state")
		wantState, ok := session.Values[stateCookie].(string)
		delete(session.Values, stateCookie)
		if !ok || state == "" || state != wantState {
			http.Error(w, "Incorrect state value", http.StatusBadRequest)
			return
		}

		// get the token
		code := r.FormValue(codeQueryKey)
		token, err := o.authCfg.Exchange(ctx, code)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// make a client so we can get the user info and groups
		oktaClient := NewOktaClient(o.authCfg.Client(ctx, token), o.oktaDomain)

		// get the user profile
		userInfo, _, err := oktaClient.GetOpenIDUser(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// save the username and groups to the cookie
		session.Values[userCookie] = userInfo.Username
		session.Values[groupsCookie] = userInfo.Groups
		session.Save(r, w)

		// redirect back to /app
		http.Redirect(w, r, "/app", http.StatusFound)
	})
}

// SecuredHandler does 2 things, ensures the user has a valid session and
// places the user information into the request context for use within handers
func (o *Okta) SecuredHandler(handler, needAuthHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		session, _ := o.store.New(r, sessionName)

		// put the username into the context
		username, ok := session.Values[userCookie].(string)
		if ok {
			ctx = WithUsername(ctx, username)
			// then put groups list into the context
			groups, ok := session.Values[groupsCookie].([]string)
			if ok {
				ctx = WithGroups(ctx, groups)
			}
		}

		if !ok {
			needAuthHandler.ServeHTTP(w, r)
		} else {
			handler.ServeHTTP(w, r.WithContext(ctx))
		}
	})
}

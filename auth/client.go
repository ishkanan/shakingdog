package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
)

// OktaClient some helper functions for context-aware queries to Okta
type OktaClient struct {
	client  *http.Client
	BaseURL *url.URL
}

// NewOktaClient builds a new okta client using the provided http client and
// okta details.
// The provided should be authenticated, e.g. using oauth2.Config
func NewOktaClient(httpClient *http.Client, oktaDomain *url.URL) *OktaClient {
	return &OktaClient{
		client:  httpClient,
		BaseURL: oktaDomain,
	}
}

// NewRequest builds a new request and sets content-type, accept, etc
// as required
func (c *OktaClient) NewRequest(method, urlStr string, body interface{}) (*http.Request, error) {
	u, err := c.BaseURL.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	var buf io.ReadWriter
	if body != nil {
		buf = new(bytes.Buffer)
		enc := json.NewEncoder(buf)
		enc.SetEscapeHTML(false)
		err := enc.Encode(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return req, nil
}

// Do will perform a single API request to Okta, parsing the returned JSON into
// v, if v is non-nil.
// v can also be an io.Writer, in which case the response body will be copied
// into it.
func (c *OktaClient) Do(ctx context.Context, req *http.Request, v interface{}) (*http.Response, error) {
	req = req.WithContext(ctx)

	resp, err := c.client.Do(req)
	if err != nil {
		// If we got an error, and the context has been canceled,
		// the context's error is probably more useful.
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		return nil, err
	}

	defer func() {
		// Drain up to 512 bytes and close the body to let the Transport reuse the connection
		io.CopyN(ioutil.Discard, resp.Body, 512)
		resp.Body.Close()
	}()

	err = CheckResponse(resp)
	if err != nil {
		// even though there was an error, we still return the response
		// in case the caller wants to inspect it further
		return resp, err
	}

	if v != nil {
		if w, ok := v.(io.Writer); ok {
			io.Copy(w, resp.Body)
		} else {
			err = json.NewDecoder(resp.Body).Decode(v)
			if err == io.EOF {
				err = nil // ignore EOF errors caused by empty response body
			}
		}
	}

	return resp, err
}

// ErrorResponse captures the information returned from Okta in an API error
type ErrorResponse struct {
	Response     *http.Response // HTTP response that caused this error
	ErrorCode    string         `json:"errorCode"`
	ErrorSummary string         `json:"errorSummary"`
	ErrorLink    string         `json:"errorLink"`
	ErrorID      string         `json:"errorId"`
	ErrorCauses  []struct {
		ErrorSummary string `json:"errorSummary"`
	} `json:"errorCauses"`
}

func (r *ErrorResponse) Error() string {
	return fmt.Sprintf("%v %v: %d %v %+v",
		r.Response.Request.Method, r.Response.Request.URL,
		r.Response.StatusCode, r.ErrorSummary, r.ErrorSummary)
}

// CheckResponse verifies that the returned http response from Okta is valid
func CheckResponse(r *http.Response) error {
	if c := r.StatusCode; 200 <= c && c <= 299 {
		return nil
	}
	errorResponse := &ErrorResponse{Response: r}
	data, err := ioutil.ReadAll(r.Body)
	if err == nil && data != nil {
		json.Unmarshal(data, errorResponse)
	}

	return errorResponse
}

// OktaUser represents a single okta user
type OktaUser struct {
	ID      string `json:"id"`
	Status  string `json:"status"`
	Profile struct {
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
		Email     string `json:"email"`
		Login     string `json:"login"`
		Mobile    string `json:"mobilePhone"`
	} `json:"profile"`
}

// GetCurrentUser fetches the user associated with the current login
// This works with an APIKEY only, Bearer auth is not supported
func (c *OktaClient) GetCurrentUser(ctx context.Context) (*OktaUser, *http.Response, error) {
	req, err := c.NewRequest("GET", "/api/v1/users/me", nil)
	if err != nil {
		return nil, nil, err
	}

	user := new(OktaUser)
	resp, err := c.Do(ctx, req, user)
	if err != nil {
		return nil, resp, err
	}

	return user, resp, nil
}

// OktaGroup represents a single Okta Group
type OktaGroup struct {
	ID      string `json:"id"`
	Profile struct {
		Name string `json:"name"`
		Desc string `json:"description"`
	} `json:"profile"`
}

// GetUserGroups fetches the list of groups that the given user is a member of
// This works with an APIKEY only, Bearer auth is not supported
func (c *OktaClient) GetUserGroups(ctx context.Context, userID string) ([]OktaGroup, *http.Response, error) {
	var groups []OktaGroup

	path := fmt.Sprintf("/api/v1/users/%s/groups", userID)
	req, err := c.NewRequest("GET", path, nil)
	if err != nil {
		return nil, nil, err
	}

	resp, err := c.Do(ctx, req, groups)
	if err != nil {
		return nil, resp, err
	}

	return groups, resp, nil
}

// BearerUser is a single Okta user as obtainable via the OpenID UserInfo endpoint
type BearerUser struct {
	Sub        string   `json:"sub"`
	Name       string   `json:"name"`
	Locale     string   `json:"locale"`
	Email      string   `json:"email"`
	Username   string   `json:"preferred_username"`
	GivenName  string   `json:"given_name"`
	FamilyName string   `json:"family_name"`
	ZoneInfo   string   `json:"zoneinfo"`
	Groups     []string `json:"groups"`
}

// GetOpenIDUser fetches the user information via the OpenID endpoint
// This required Bearer auth, e.g. as provided by oauth2.Client
func (c *OktaClient) GetOpenIDUser(ctx context.Context) (*BearerUser, *http.Response, error) {

	req, err := c.NewRequest("GET", "/oauth2/v1/userinfo", nil)
	if err != nil {
		return nil, nil, err
	}

	user := new(BearerUser)
	resp, err := c.Do(ctx, req, user)
	if err != nil {
		return nil, resp, err
	}

	return user, resp, nil
}

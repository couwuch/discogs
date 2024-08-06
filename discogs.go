// Package discogs provides a client library for interacting with the Discogs API.
//
// The Discogs API allows developers to access and manage data related to music releases, artists, labels, and more.
// This package simplifies the process of making HTTP requests to the Discogs API by providing a structured client
// that handles authentication, parameterized routes, and response parsing. More information can be found at the
// [Discogs developers page].
//
// [Discogs developers page]: https://www.discogs.com/developers
package discogs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

const (
	BaseURL         = "https://api.discogs.com"
	DefaultAppName  = "DiscogsGo/0.1"
	AuthHeader      = "Authorization"
	UserAgentHeader = "User-Agent"
	RateLimitHeader = "X-Discogs-Ratelimit"
	RateLimitUnauth = 25
	RateLimitAuth   = 60
)

// An HTTPError provides information on an error resulting from an HTTP request, including the StatusCode
// and Message.
type HTTPError struct {
	StatusCode int
	Message    string
}

// Error provides a human digestible string of an HTTPError.
//
// Example: "HTTP 404: Item not found".
func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Message)
}

// DiscogsClient is a wrapper for http.Client that includes the Host of the API and a Config for the client.
// It also includes a rateLimiter to rate limit requests.
type DiscogsClient struct {
	*http.Client
	Host   string
	Config DiscogsConfig

	rateLimiter *rate.Limiter
	mu          sync.Mutex
}

// DiscogsConfig contains configuration options for the Discogs client.
//
// TODO: add support for AccessToken and OAuth tokens.
type DiscogsConfig struct {
	// Provided as User-Agent string to identify the application to Discogs.
	// Preferably follows [RFC 1945].
	//
	// Example: MyDiscogsClient/1.0 +http://mydiscogsclient.org
	//
	// [RFC 1945]: http://tools.ietf.org/html/rfc1945#section-3.7
	AppName        string
	ConsumerKey    *string
	ConsumerSecret *string
	AccessToken    *string
	MaxRequests    int
}

// NewDiscogsClient creates a new DiscogsClient with the provided configuration.
// If AppName is not provided in the config, it defaults to DefaultAppName.
// TODO: handle AccessToken
func NewDiscogsClient(config *DiscogsConfig) *DiscogsClient {
	if config.AppName == "" {
		config.AppName = DefaultAppName
	}

	var limiter *rate.Limiter
	if config.MaxRequests > 0 {
		limiter = rate.NewLimiter(rate.Every(time.Minute/time.Duration(config.MaxRequests)), config.MaxRequests)
	} else {
		var requestsPerMinute int
		if config.ConsumerKey != nil && config.ConsumerSecret != nil {
			requestsPerMinute = RateLimitAuth
		} else {
			requestsPerMinute = RateLimitUnauth
		}
		limiter = rate.NewLimiter(rate.Every(time.Minute/time.Duration(requestsPerMinute)), requestsPerMinute)
	}

	return &DiscogsClient{
		Client:      &http.Client{},
		Host:        BaseURL,
		Config:      *config,
		rateLimiter: limiter,
	}
}

// Get sends an HTTP GET request to the specified endpoint with the given parameters and headers,
// and unmarshals the response into the provided res interface.
func (dc *DiscogsClient) Get(ctx context.Context, endpoint string, params url.Values, headers map[string]string, res interface{}) error {
	return dc.request(ctx, http.MethodGet, endpoint, params, headers, nil, res)
}

// Post sends an HTTP POST request to the specified endpoint with the given parameters, headers, and body,
// and unmarshals the response into the provided res interface.
func (dc *DiscogsClient) Post(ctx context.Context, endpoint string, params url.Values, headers map[string]string, body, res interface{}) error {
	return dc.request(ctx, http.MethodPost, endpoint, params, headers, body, res)
}

// Put sends an HTTP PUT request to the specified endpoint with the given parameters, headers, and body,
// and unmarshals the response into the provided res interface.
func (dc *DiscogsClient) Put(ctx context.Context, endpoint string, params url.Values, headers map[string]string, body, res interface{}) error {
	return dc.request(ctx, http.MethodPut, endpoint, params, headers, body, res)
}

// Delete sends an HTTP DELETE request to the specified endpoint with the given parameters and headers,
// and unmarshals the response into the provided res interface.
func (dc *DiscogsClient) Delete(ctx context.Context, endpoint string, params url.Values, headers map[string]string, res interface{}) error {
	return dc.request(ctx, http.MethodDelete, endpoint, params, headers, nil, res)
}

// request sends an HTTP request to the specified endpoint with the given parameters, headers, and body,
// and unmarshals the response into the provided res interface. It respects the rate limit settings of the Discogs API
// and any user-defined rate limits. It handles the request creation, including setting authentication headers.
func (dc *DiscogsClient) request(ctx context.Context, method, endpoint string, params url.Values, headers map[string]string, body, res interface{}) error {
	baseURL, err := url.Parse(dc.Host + endpoint)
	if err != nil {
		return err
	}

	if params == nil {
		params = url.Values{}
	}
	baseURL.RawQuery = params.Encode()

	var reqBody io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return err
		}

		reqBody = bytes.NewReader(encoded)
	}

	req, err := http.NewRequestWithContext(ctx, method, baseURL.String(), reqBody)
	if err != nil {
		return err
	}

	// Set headers from the provided map
	for headerKey, headerValue := range headers {
		req.Header.Set(headerKey, headerValue)
	}

	// Determine authentication type based on the endpoint
	authType, err := matchRoute(endpoint, EndpointAuthMap)
	if err != nil {
		return err
	}

	// Set the User-Agent header to AppName, as requested by Discogs API
	req.Header.Set(UserAgentHeader, dc.Config.AppName)

	// Add authentication headers to the request
	if err := dc.addAuthHeaders(req, authType); err != nil {
		return err
	}

	return dc.Do(ctx, req, res)
}

// Do sends an HTTP request and unmarshals the response into the provided res interface.
// It respects the rate limits by waiting until the rate limiter allows the request.
// It also updates the rate limiter based on the X-Discogs-Ratelimit header from the API response.
// It returns an HTTPError if the response status code is not 2xx.
func (dc *DiscogsClient) Do(ctx context.Context, req *http.Request, res interface{}) error {
	err := dc.rateLimiter.Wait(ctx)
	if err != nil {
		return err
	}

	response, err := dc.Client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer response.Body.Close()

	dc.updateRateLimitFromHeader(response)

	// Read the response body
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for non-2xx status codes and return an HTTPError if necessary
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return &HTTPError{
			StatusCode: response.StatusCode,
			Message:    string(responseBody),
		}
	}

	// Unmarshal the response body into the provided res interface, if not nil
	if res != nil && len(responseBody) > 0 {
		if err := json.Unmarshal(responseBody, res); err != nil {
			return fmt.Errorf("failed to unmarshal response body: %w", err)
		}
	}

	return nil
}

// SetMaxRequests allows the user to set a custom rate limit for the DiscogsClient.
// It adjusts the rate limiter to the specified number of requests per minute.
func (dc *DiscogsClient) SetMaxRequests(requestsPerMinute int) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	if requestsPerMinute > 0 {
		dc.Config.MaxRequests = requestsPerMinute
		limit := rate.Every(time.Minute / time.Duration(requestsPerMinute))
		dc.rateLimiter.SetLimit(limit)
		dc.rateLimiter.SetBurst(requestsPerMinute)
	}
}

// Limit returns the current rate limit of the DiscogsClient.
func (dc *DiscogsClient) Limit() rate.Limit {
	return dc.rateLimiter.Limit()
}

// Tokens returns the number of tokens (available requests) currently in the rate limiter's bucket.
func (dc *DiscogsClient) Tokens() float64 {
	return dc.rateLimiter.Tokens()
}

// updateRateLimitFromHeader adjusts the rate limiter based on the X-Discogs-Ratelimit header from the API response.
// It sets the rate limit to the minimum of the user-defined limit and the Discogs API limit. The rate limiter will never
// be set to 0 (infinite requests).
//
// Note: The MaxRequests (per minute) in the config does not update based on the headers. Use SetMaxRequests to change the
// MaxRequests value along with the rate limiter.
func (dc *DiscogsClient) updateRateLimitFromHeader(res *http.Response) {
	rateLimitHeader := res.Header.Get(RateLimitHeader)

	if rateLimitHeader != "" {
		rateLimit, err := strconv.Atoi(rateLimitHeader)
		if err == nil {
			dc.mu.Lock()
			defer dc.mu.Unlock()

			// Use the minimum of user-defined limit and Discogs API limit
			var effectiveLimit int
			if dc.Config.MaxRequests > 0 {
				effectiveLimit = min(dc.Config.MaxRequests, rateLimit)
			} else {
				effectiveLimit = rateLimit
			}

			if effectiveLimit != 0 {
				dc.rateLimiter.SetLimit(rate.Every(time.Minute / time.Duration(effectiveLimit)))
				dc.rateLimiter.SetBurst(effectiveLimit)
			}
		}
	}
}

// AuthType represents the type of authentication required for an endpoint.
type AuthType string

const (
	AuthTypeUnknown   AuthType = "unknown"
	AuthTypeNone      AuthType = "none"
	AuthTypeKeySecret AuthType = "key-secret"
	AuthTypeOAuth     AuthType = "oauth"
	AuthTypePAT       AuthType = "personal-access-token"
)

// ErrMissingCredentials represents an error when credentials required for an endpoint are missing from
// the DiscogsConfig.
type ErrMissingCredentials struct {
	RequiredAuthType AuthType
	Endpoint         string
}

func (e *ErrMissingCredentials) Error() string {
	return fmt.Sprintf("missing required auth credentials of type %s for endpoint: %s", e.RequiredAuthType, e.Endpoint)
}

// addAuthHeaders adds the appropriate authentication headers to a request based on the authentication type.
// More information about the authentication process for Discogs can be found at [Discogs Auth Flow].
//
// [Discogs Auth Flow]: https://www.discogs.com/developers#page:authentication,header:authentication-discogs-auth-flow
func (dc *DiscogsClient) addAuthHeaders(req *http.Request, authType AuthType) error {
	switch authType {
	case AuthTypeNone:
		// No authentication required
	case AuthTypeKeySecret:
		if dc.Config.ConsumerKey != nil && dc.Config.ConsumerSecret != nil {
			req.Header.Set(AuthHeader, fmt.Sprintf("Discogs key=%s, secret=%s", *dc.Config.ConsumerKey, *dc.Config.ConsumerSecret))
		} else {
			return &ErrMissingCredentials{RequiredAuthType: authType, Endpoint: req.URL.Path}
		}
	case AuthTypeOAuth, AuthTypePAT:
		if dc.Config.AccessToken != nil {
			req.Header.Set(AuthHeader, fmt.Sprintf("Bearer %s", *dc.Config.AccessToken))
		} else {
			return &ErrMissingCredentials{RequiredAuthType: authType, Endpoint: req.URL.Path}
		}
	}
	return nil
}

// ErrMatchNotFound represents an error when no matching route is found for authentication.
type ErrMatchNotFound struct {
	Endpoint string
}

func (e *ErrMatchNotFound) Error() string {
	return fmt.Sprintf("no matching route found for endpoint: %s", e.Endpoint)
}

// endpointAuthMap maps API endpoints to their required authentication types.
var EndpointAuthMap = map[string]AuthType{
	"/test":                  AuthTypeNone,
	"/releases/{release_id}": AuthTypeNone,
	"/database/search":       AuthTypeKeySecret,
}

// matchRoute determines the authentication type required for a given endpoint.
func matchRoute(endpoint string, authMap map[string]AuthType) (AuthType, error) {
	for route, authType := range authMap {
		if isMatch(route, endpoint) {
			return authType, nil
		}
	}
	return AuthTypeUnknown, &ErrMatchNotFound{Endpoint: endpoint}
}

// isMatch checks if a given endpoint matches a route pattern.
func isMatch(route, endpoint string) bool {
	routeParts := strings.Split(route, "/")
	endpointParts := strings.Split(endpoint, "/")
	if len(routeParts) != len(endpointParts) {
		return false
	}
	for i, part := range routeParts {
		if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
			continue
		}
		if part != endpointParts[i] {
			return false
		}
	}
	return true
}

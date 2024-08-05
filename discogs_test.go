package discogs_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/couwuch/discogs"
	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"
)

var ctx = context.Background()
var key = "key"
var secret = "secret"

type TestClientResponse struct {
	Success bool `json:"success"`
}

type Body struct {
	Name string `json:"name"`
}

func TestNewDiscogsClient(t *testing.T) {
	type want struct {
		appName   string
		rateLimit rate.Limit
	}
	var tests = []struct {
		name string
		args *discogs.DiscogsConfig
		want want
	}{
		{
			"NewDiscogsClient with no AppName, unauthenticated rate limit",
			&discogs.DiscogsConfig{},
			want{discogs.DefaultAppName, rate.Every(time.Minute / time.Duration(discogs.RateLimitUnauth))},
		},
		{
			"NewDiscogsClient with AppName, authenticated rate limit",
			&discogs.DiscogsConfig{AppName: "Test", ConsumerKey: &key, ConsumerSecret: &secret},
			want{"Test", rate.Every(time.Minute / time.Duration(discogs.RateLimitAuth))},
		},
		{
			"NewDiscogsClient with specified rate limit",
			&discogs.DiscogsConfig{MaxRequests: 100},
			want{discogs.DefaultAppName, rate.Every(time.Minute / time.Duration(100))},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			testClient := discogs.NewDiscogsClient(tt.args)
			assert.Equal(t, tt.want.appName, testClient.Config.AppName)
			assert.Equal(t, tt.want.rateLimit, testClient.Limit())
		})
	}
}

func TestDiscogsClient_Get(t *testing.T) {
	t.Parallel()

	endpoint := "/test"
	params := url.Values{}
	params.Set("param1", "value1")
	params.Set("param2", "value2")

	response := TestClientResponse{Success: true}

	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, http.MethodGet, req.Method)

		for key, value := range params {
			assert.Equal(t, value, req.URL.Query()[key])
		}

		responseBody, err := json.Marshal(response)
		if err != nil {
			t.Fatalf("unable to marshal json response")
		}

		rw.Write(responseBody)
	}))
	defer server.Close()

	testClient := discogs.NewDiscogsClient(&discogs.DiscogsConfig{})
	testClient.Host = server.URL
	var res TestClientResponse

	err := testClient.Get(ctx, endpoint, params, nil, &res)

	assert.NoError(t, err)
	assert.Equal(t, response.Success, res.Success)
}

func TestDiscogsClient_Post(t *testing.T) {
	t.Parallel()

	endpoint := "/test"
	body := Body{Name: "test"}

	response := TestClientResponse{Success: true}

	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, http.MethodPost, req.Method)

		var b Body

		err := json.NewDecoder(req.Body).Decode(&b)
		if err != nil {
			t.Fatalf("unable to decode request body into struct")
		}

		assert.Equal(t, body.Name, b.Name)

		responseBody, err := json.Marshal(response)
		if err != nil {
			t.Fatalf("unable to marshal json response")
		}

		rw.Write(responseBody)
	}))
	defer server.Close()

	testClient := discogs.NewDiscogsClient(&discogs.DiscogsConfig{})
	testClient.Host = server.URL
	var res TestClientResponse

	err := testClient.Post(ctx, endpoint, nil, nil, body, &res)

	assert.NoError(t, err)
	assert.Equal(t, response.Success, res.Success)
}

func TestDiscogsClient_Put(t *testing.T) {
	t.Parallel()

	endpoint := "/test"
	body := Body{Name: "test"}

	response := TestClientResponse{Success: true}

	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, http.MethodPut, req.Method)

		var b Body

		err := json.NewDecoder(req.Body).Decode(&b)
		if err != nil {
			t.Fatalf("unable to decode request body into struct")
		}

		assert.Equal(t, body.Name, b.Name)

		responseBody, err := json.Marshal(response)
		if err != nil {
			t.Fatalf("unable to marshal json response")
		}

		rw.Write(responseBody)
	}))
	defer server.Close()

	testClient := discogs.NewDiscogsClient(&discogs.DiscogsConfig{})
	testClient.Host = server.URL
	var res TestClientResponse

	err := testClient.Put(ctx, endpoint, nil, nil, body, &res)

	assert.NoError(t, err)
	assert.Equal(t, response.Success, res.Success)
}

func TestDiscogsClient_Delete(t *testing.T) {
	t.Parallel()

	endpoint := "/test"

	response := TestClientResponse{Success: true}

	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, http.MethodDelete, req.Method)

		responseBody, err := json.Marshal(response)
		if err != nil {
			assert.FailNow(t, "unable to marshal json response: %v", err)
		}

		rw.Write(responseBody)
	}))
	defer server.Close()

	testClient := discogs.NewDiscogsClient(&discogs.DiscogsConfig{})
	testClient.Host = server.URL
	var res TestClientResponse

	err := testClient.Delete(ctx, endpoint, nil, nil, &res)

	assert.NoError(t, err)
	assert.Equal(t, response.Success, res.Success)
}

// Helper function to create a mock response with the given rate limit headers
func createMockResponse(rateLimit int) *http.Response {
	recorder := httptest.NewRecorder()
	recorder.Header().Set(discogs.RateLimitHeader, strconv.Itoa(rateLimit))
	return recorder.Result()
}

func TestUpdateRateLimitFromHeader(t *testing.T) {
	type args struct {
		maxRequests    int
		newMaxRequests int
	}
	tests := []struct {
		name string
		args args
		want rate.Limit
	}{
		{
			"Rate limit header value lower than config value",
			args{maxRequests: 50, newMaxRequests: discogs.RateLimitUnauth},
			rate.Every(time.Minute / time.Duration(discogs.RateLimitUnauth)),
		},
		{
			"Rate limit header value higher than config value",
			args{maxRequests: 20, newMaxRequests: discogs.RateLimitAuth},
			rate.Every(time.Minute / time.Duration(20)),
		},
		{
			"Rate limit header value 0",
			args{maxRequests: 20, newMaxRequests: 0},
			rate.Every(time.Minute / time.Duration(20)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := discogs.NewDiscogsClient(&discogs.DiscogsConfig{MaxRequests: tt.args.maxRequests})

			response := createMockResponse(tt.args.newMaxRequests)
			client.UpdateRateLimitFromHeader(response)

			assert.Equal(t, tt.want, client.Limit())
		})
	}
}

// TODO: add tests for OAuth and PAT when implemented
func TestDiscogsClient_addAuthHeaders(t *testing.T) {
	var body io.Reader
	endpoint := "/test"

	type args struct {
		config   *discogs.DiscogsConfig
		authType discogs.AuthType
	}

	type want struct {
		header string
		err    error
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			"addAuthHeaders AuthTypeNone",
			args{&discogs.DiscogsConfig{}, discogs.AuthTypeNone},
			want{"", nil},
		},
		{
			"addAuthHeaders AuthTypeKeySecret missing credentials",
			args{&discogs.DiscogsConfig{}, discogs.AuthTypeKeySecret},
			want{"", &discogs.ErrMissingCredentials{discogs.AuthTypeKeySecret, endpoint}},
		},
		{
			"addAuthHeaders AuthTypeKeySecret with credentials",
			args{&discogs.DiscogsConfig{ConsumerKey: &key, ConsumerSecret: &secret}, discogs.AuthTypeKeySecret},
			want{fmt.Sprintf("Discogs key=%s, secret=%s", key, secret), nil},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dc := discogs.NewDiscogsClient(tt.args.config)

			req, err := http.NewRequest(http.MethodGet, endpoint, body)
			if err != nil {
				assert.FailNow(t, "error creating http request: %v", err)
			}

			if err := dc.AddAuthHeaders(req, tt.args.authType); err != nil {
				assert.EqualError(t, tt.want.err, err.Error())
			} else {
				assert.NoError(t, tt.want.err)
			}
			assert.Equal(t, req.Header.Get(discogs.AuthHeader), tt.want.header)
		})
	}
}

func Test_matchRoute(t *testing.T) {
	endpointAuthMap := map[string]discogs.AuthType{
		"/none":       discogs.AuthTypeNone,
		"/key/secret": discogs.AuthTypeKeySecret,
		"/oauth":      discogs.AuthTypeOAuth,
		"/pat":        discogs.AuthTypePAT,
	}

	type want struct {
		authType discogs.AuthType
		err      error
	}
	type args struct {
		endpoint string
		authMap  map[string]discogs.AuthType
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			"matchRoute match not found",
			args{"/not/found", endpointAuthMap},
			want{discogs.AuthTypeUnknown, &discogs.ErrMatchNotFound{"/not/found"}},
		},
		{
			"matchRoute AuthTypeNone",
			args{"/none", endpointAuthMap},
			want{discogs.AuthTypeNone, nil},
		},
		{
			"matchRoute AuthTypeKeySecret",
			args{"/key/secret", endpointAuthMap},
			want{discogs.AuthTypeKeySecret, nil},
		},
		{
			"matchRoute AuthTypeOAuth",
			args{"/oauth", endpointAuthMap},
			want{discogs.AuthTypeOAuth, nil},
		},
		{
			"matchRoute AuthTypePAT",
			args{"/pat", endpointAuthMap},
			want{discogs.AuthTypePAT, nil},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := discogs.MatchRoute(tt.args.endpoint, tt.args.authMap)
			if tt.want.err != nil {
				assert.EqualError(t, tt.want.err, err.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, got, tt.want.authType)
		})
	}
}

func Test_isMatch(t *testing.T) {
	type args struct {
		route    string
		endpoint string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"isMatch no match",
			args{"/no/match", "/match"},
			false,
		},
		{
			"isMatch basic match",
			args{"/basic", "/basic"},
			true,
		},
		{
			"isMatch nested match",
			args{"/nested/match", "/nested/match"},
			true,
		},
		{
			"isMatch param match",
			args{"/param/{id}", "/param/1234"},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := discogs.IsMatch(tt.args.route, tt.args.endpoint)
			assert.Equal(t, tt.want, got)
		})
	}
}

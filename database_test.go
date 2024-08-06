package discogs_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/couwuch/discogs"
	"github.com/stretchr/testify/assert"
)

func TestDatabase_Release(t *testing.T) {
	type args struct {
		releaseID int64
		options   *discogs.ReleaseOptions
	}
	type want struct {
		res *discogs.ReleaseResponse
		err error
	}
	type mock struct {
		status int
		res    interface{}
	}
	tests := []struct {
		name string
		args args
		mock mock
		want want
	}{
		{
			"successful release fetch",
			args{1, &discogs.ReleaseOptions{discogs.CurrencyUSD}},
			mock{http.StatusOK, discogs.ReleaseResponse{Title: "Test Release", ID: 1}},
			want{&discogs.ReleaseResponse{Title: "Test Release", ID: 1}, nil},
		},
		{
			"release not found",
			args{2, nil},
			mock{http.StatusNotFound, struct {
				Message string `json:"message"`
			}{"Release not found."}},
			want{nil, &discogs.ErrReleaseNotFound{2, &discogs.HTTPError{http.StatusNotFound, `{"message":"Release not found."}`}}},
		},
		{
			"unexpected error",
			args{3, nil},
			mock{http.StatusInternalServerError, struct {
				Message string `json:"message"`
			}{"Internal server error."}},
			want{nil, &discogs.HTTPError{http.StatusInternalServerError, `{"message":"Internal server error."}`}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock server
			server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				rw.WriteHeader(tt.mock.status)

				responseBody, err := json.Marshal(tt.mock.res)
				if err != nil {
					assert.FailNow(t, "unable to marshal json response: %w", err)
				}

				if _, err := rw.Write(responseBody); err != nil {
					assert.FailNow(t, "failed to write the response body: %w", err)
				}
			}))
			defer server.Close()

			// Configure the DiscogsClient to use the mock server
			client := discogs.NewDiscogsClient(&discogs.DiscogsConfig{AppName: "DiscogsGo/0.1", ConsumerKey: &key, ConsumerSecret: &secret})
			client.Host = server.URL

			// Call the Release method
			res, err := client.Release(ctx, tt.args.releaseID, tt.args.options)

			// Check the error
			if err != nil {
				assert.EqualError(t, err, tt.want.err.Error())
			} else {
				assert.NoError(t, tt.want.err)
			}

			// Check the response
			if res != nil {
				assert.Equal(t, res, tt.want.res)
			} else {
				assert.Nil(t, tt.want.res)
			}
		})
	}
}

func TestDatabase_Search(t *testing.T) {
	type want struct {
		res *discogs.SearchResponse
		err error
	}
	type mock struct {
		status int
		res    interface{}
	}
	tests := []struct {
		name string
		args *discogs.SearchOptions
		mock mock
		want want
	}{
		{
			"successful search",
			&discogs.SearchOptions{Track: "Test Search"},
			mock{http.StatusOK, discogs.SearchResponse{
				Results: []discogs.SearchResult{{Title: "Test Search"}},
			}},
			want{&discogs.SearchResponse{
				Results: []discogs.SearchResult{{Title: "Test Search"}},
			}, nil},
		},
		{
			"search with no results",
			&discogs.SearchOptions{Track: "Test Search"},
			mock{http.StatusOK, discogs.SearchResponse{}},
			want{&discogs.SearchResponse{}, nil},
		},
		{
			"query time exceeded",
			&discogs.SearchOptions{Track: "Test Search"},
			mock{500, struct {
				Message string `json:"message"`
			}{"Query time exceeded. Please try a simpler query."}},
			want{nil, &discogs.HTTPError{500, `{"message":"Query time exceeded. Please try a simpler query."}`}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock server
			server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				rw.WriteHeader(tt.mock.status)

				responseBody, err := json.Marshal(tt.mock.res)
				if err != nil {
					assert.FailNow(t, "unable to marshal json response: %w", err)
				}

				if _, err := rw.Write(responseBody); err != nil {
					assert.FailNow(t, "failed to write the response body: %w", err)
				}
			}))
			defer server.Close()

			// Configure the DiscogsClient to use the mock server
			client := discogs.NewDiscogsClient(&discogs.DiscogsConfig{AppName: "DiscogsGo/0.1", ConsumerKey: &key, ConsumerSecret: &secret})
			client.Host = server.URL

			// Call the Release method
			res, err := client.Search(ctx, tt.args)

			// Check the error
			if err != nil {
				assert.EqualError(t, err, tt.want.err.Error())
			} else {
				assert.NoError(t, tt.want.err)
			}

			// Check the response
			if res != nil {
				assert.Equal(t, res, tt.want.res)
			} else {
				assert.Nil(t, tt.want.res)
			}
		})
	}
}

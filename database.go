package discogs

import (
	"context"
	"net/http"
	"strconv"

	"github.com/google/go-querystring/query"
)

// Release fetches detailed information about a release from the Discogs database
// by sending a GET request to the /releases/{release_id} endpoint.
// The releaseID specifies the ID of the release to fetch, and options allows for
// additional query parameters. The context.Context provides control over the request's lifecycle.
// It returns a pointer to a ReleaseResponse struct containing the release details,
// or an error if the request fails or the release is not found.
//
// Documentation: https://www.discogs.com/developers#page:database,header:database-release
func (dc *DiscogsClient) Release(ctx context.Context, releaseID int64, options *ReleaseOptions) (*ReleaseResponse, error) {
	endpoint := "/releases/" + strconv.FormatInt(releaseID, 10)
	var res ReleaseResponse

	params, err := query.Values(options)
	if err != nil {
		return nil, err
	}

	if err := dc.Get(ctx, endpoint, params, nil, &res); err != nil {
		if httpErr, ok := err.(*HTTPError); ok {
			if httpErr.StatusCode == http.StatusNotFound {
				return nil, &ErrReleaseNotFound{
					ReleaseID: int(releaseID),
					HTTPError: httpErr,
				}
			}
			return nil, httpErr
		}
		return nil, err
	}

	return &res, nil
}

// https://www.discogs.com/developers#page:database,header:database-release-rating-by-user
// GET /releases/{release_id}/rating/{username}

// PUT /releases/{release_id}/rating/{username}

// DELETE /releases/{release_id}/rating/{username}

// https://www.discogs.com/developers#page:database,header:database-community-release-rating
// GET /releases/{release_id}/rating

// https://www.discogs.com/developers#page:database,header:database-release-stats
// GET /releases/{release_id}/stats

// https://www.discogs.com/developers#page:database,header:database-master-release
// GET /masters/{master_id}

// https://www.discogs.com/developers#page:database,header:database-master-release-versions
// GET /masters/{master_id}/versions{?page,per_page}

// https://www.discogs.com/developers#page:database,header:database-artist
// GET /artists/{artist_id}

// https://www.discogs.com/developers#page:database,header:database-artist-releases
// GET /artists/{artist_id}/releases{?sort,sort_order}

// https://www.discogs.com/developers#page:database,header:database-label
// GET /labels/{label_id}

// https://www.discogs.com/developers#page:database,header:database-all-label-releases
// GET /labels/{label_id}/releases{?page,per_page}

// Search performs a search query against the Discogs database by sending a GET request
// to the /database/search endpoint. The options parameter specifies the search options,
// such as query and type. The context.Context provides control over the request's lifecycle.
// It returns a pointer to a SearchResponse struct containing the search results,
// or an error if the request fails.
func (dc *DiscogsClient) Search(ctx context.Context, options *SearchOptions) (*SearchResponse, error) {
	endpoint := "/database/search"
	var res SearchResponse

	params, err := query.Values(options)
	if err != nil {
		return nil, err
	}

	if err := dc.Get(ctx, endpoint, params, nil, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

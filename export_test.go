package discogs

import (
	"context"
	"net/http"
)

func (dc *DiscogsClient) UpdateRateLimitFromHeader(res *http.Response) {
	dc.updateRateLimitFromHeader(res)
}

func (dc *DiscogsClient) Wait(ctx context.Context) error {
	return dc.rateLimiter.Wait(ctx)
}

func (dc *DiscogsClient) AddAuthHeaders(req *http.Request, authType AuthType) error {
	return dc.addAuthHeaders(req, authType)
}

var MatchRoute = matchRoute

var IsMatch = isMatch

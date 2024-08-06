package discogs

// Pagination represents the pagination information returned by the Discogs API.
//
// See https://www.discogs.com/developers#page:home,header:home-pagination
type Pagination struct {
	Page    int64 `json:"page"`
	Pages   int64 `json:"pages"`
	Items   int64 `json:"items"`
	PerPage int64 `json:"per_page"`
	Urls    *struct {
		First string `json:"first"`
		Prev  string `json:"prev"`
		Next  string `json:"next"`
		Last  string `json:"last"`
	} `json:"urls"`
}

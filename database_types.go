package discogs

import (
	"fmt"
	"time"
)

// Currency represents a currency code used in the Discogs API.
type Currency string

// Currency constants representing various supported currencies.
const (
	CurrencyUSD Currency = "USD"
	CurrencyGBP Currency = "GBP"
	CurrencyEUR Currency = "EUR"
	CurrencyCAD Currency = "CAD"
	CurrencyAUD Currency = "AUD"
	CurrencyJPY Currency = "JPY"
	CurrencyCHF Currency = "CHF"
	CurrencyMXN Currency = "MXN"
	CurrencyBRL Currency = "BRL"
	CurrencyNZD Currency = "NZD"
	CurrencySEK Currency = "SEK"
	CurrencyZAR Currency = "ZAR"
)

// Type represents a type of entity in the Discogs database.
type Type string

// Type constants representing various entity types.
const (
	TypeRelease = "release"
	TypeMaster  = "master"
	TypeArtist  = "artist"
	TypeLabel   = "label"
)

// ErrReleaseNotFound indicates that a release with the specified ID was not found.
type ErrReleaseNotFound struct {
	ReleaseID int
	*HTTPError
}

// Error returns a formatted error message indicating that the release was not found.
func (e *ErrReleaseNotFound) Error() string {
	return fmt.Sprintf("Release ID %d not found: %s", e.ReleaseID, e.Message)
}

// PaginationParams represents the pagination parameters for API requests.
type PaginationParams struct {
	Page    *int `url:"page,omitempty"`
	PerPage *int `url:"per_page,omitempty"` // The number of items per page. Default is 50. Maximum is 100.
}

// ReleaseOptions represents the options for retrieving a release.
type ReleaseOptions struct {
	CurrAbr Currency `url:"curr_abbr,omitempty"`
}

// ReleaseResponse represents the response from the Discogs API for a release.
type ReleaseResponse struct {
	Title   string `json:"title"`
	ID      int64  `json:"id"`
	Artists []struct {
		ANV         string `json:"anv"`
		ID          *int64 `json:"id"`
		Join        string `json:"join"`
		Name        string `json:"name"`
		ResourceURL string `json:"resource_url"`
		Role        string `json:"role"`
		Tracks      string `json:"tracks"`
	} `json:"artists"`
	DataQuality string `json:"data_quality"`
	Thumb       string `json:"thumb"`
	Community   *struct {
		Contributors []struct {
			ResourceURL string `json:"resource_url"`
			Username    string `json:"username"`
		} `json:"contributors"`
		DataQuality string `json:"data_quality"`
		Have        *int64 `json:"have"`
		Rating      *struct {
			Average *float64 `json:"average"`
			Count   *int64   `json:"count"`
		} `json:"rating"`
		Status    *string `json:"status"`
		Submitter *struct {
			ResourceURL string `json:"resource_url"`
			Username    string `json:"username"`
		} `json:"submitter"`
		Want *int64 `json:"want"`
	} `json:"community"`
	Companies []struct {
		CatNo          string `json:"catno"`
		EntityType     string `json:"entity_type"`
		EntityTypeName string `json:"entity_type_name"`
		ID             *int64 `json:"id"`
		Name           string `json:"name"`
		ResourceURL    string `json:"resource_url"`
	} `json:"companies"`
	Country         string     `json:"country"`
	DateAdded       *time.Time `json:"date_added"`
	DateChanged     *time.Time `json:"date_changed"`
	EstimatedWeight *int64     `json:"estimated_weight"`
	ExtraArtists    []struct {
		ANV         string `json:"anv"`
		ID          *int64 `json:"id"`
		Join        string `json:"join"`
		Name        string `json:"name"`
		ResourceURL string `json:"resource_url"`
		Role        string `json:"role"`
		Tracks      string `json:"tracks"`
	} `json:"extraartists"`
	FormatQuantity *int64 `json:"format_quantity"`
	Formats        []struct {
		Descriptions []string `json:"descriptions"`
		Name         string   `json:"name"`
		Qty          string   `json:"qty"`
	} `json:"formats"`
	Genres      []string `json:"genres"`
	Identifiers []struct {
		Type  string `json:"type"`
		Value string `json:"value"`
	} `json:"identifiers"`
	Images []struct {
		Height      *int64 `json:"height"`
		ResourceURL string `json:"resource_url"`
		Type        string `json:"type"`
		URI         string `json:"uri"`
		URI150      string `json:"uri150"`
		Width       *int64 `json:"width"`
	} `json:"images"`
	Labels []struct {
		CatNo       string `json:"catno"`
		EntityType  string `json:"entity_type"`
		ID          *int64 `json:"id"`
		Name        string `json:"name"`
		ResourceURL string `json:"resource_url"`
	} `json:"labels"`
	LowestPrice       *float64      `json:"lowest_price"`
	MasterID          *int64        `json:"master_id"`
	MasterURL         string        `json:"master_url"`
	Notes             string        `json:"notes"`
	NumForSale        *int64        `json:"num_for_sale"`
	Released          string        `json:"released"`
	ReleasedFormatted string        `json:"released_formatted"`
	ResourceURL       string        `json:"resource_url"`
	Series            []interface{} `json:"series"`
	Status            string        `json:"status"`
	Styles            []string      `json:"styles"`
	Tracklist         []struct {
		Duration string `json:"duration"`
		Position string `json:"position"`
		Title    string `json:"title"`
		Type_    string `json:"type_"`
	} `json:"tracklist"`
	URI    string `json:"uri"`
	Videos []struct {
		Description string `json:"description"`
		Duration    *int64 `json:"duration"`
		Embed       *bool  `json:"embed"`
		Title       string `json:"title"`
		URI         string `json:"uri"`
	} `json:"videos"`
	Year *int64 `json:"year"`
}

// SearchOptions represents the options for performing a search query in the Discogs database.
type SearchOptions struct {
	Pagination   PaginationParams
	Query        string `url:"q,omitempty"`
	Type         Type   `url:"type,omitempty"`
	Title        string `url:"title,omitempty"`
	ReleaseTitle string `url:"release_title,omitempty"`
	Credit       string `url:"credit,omitempty"`
	Artist       string `url:"artist,omitempty"`
	ANV          string `url:"anv,omitempty"`
	Label        string `url:"label,omitempty"`
	Genre        string `url:"genre,omitempty"`
	Style        string `url:"style,omitempty"`
	Country      string `url:"country,omitempty"`
	Year         string `url:"year,omitempty"`
	Format       string `url:"format,omitempty"`
	CatNo        string `url:"catno,omitempty"`
	Barcode      string `url:"barcode,omitempty"`
	Track        string `url:"track,omitempty"`
	Submitter    string `url:"submitter,omitempty"`
	Contributor  string `url:"contributor,omitempty"`
}

// SearchResponse represents the response from the Discogs API for a search query.
type SearchResponse struct {
	Pagination *Pagination    `json:"pagination"`
	Results    []SearchResult `json:"results"`
}

// SearchResult represents a single result from a search query.
type SearchResult struct {
	Style     []string `json:"style"`
	Thumb     string   `json:"thumb"`
	Title     string   `json:"title"`
	Country   string   `json:"country"`
	Format    []string `json:"format"`
	URI       string   `json:"uri"`
	Community struct {
		Want *int64 `json:"want"`
		Have *int64 `json:"have"`
	} `json:"community"`
	Label       []string `json:"label"`
	CatNo       string   `json:"catno"`
	Year        string   `json:"year"`
	Genre       []string `json:"genre"`
	ResourceURL string   `json:"resource_url"`
	Type        Type     `json:"type"`
	ID          *int64   `json:"id"`
}

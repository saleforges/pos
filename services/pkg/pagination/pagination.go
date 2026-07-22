// Package pagination provides shared pagination types for list endpoints.
package pagination

import (
	"net/http"
	"strconv"
)

// Params holds offset/limit for pagination.
// Offset defaults to 0, Limit defaults to 20 (max 100).
type Params struct {
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

// Metadata holds pagination response metadata.
type Metadata struct {
	Total       int64 `json:"total"`
	Offset      int   `json:"offset"`
	Limit       int   `json:"limit"`
	ReturnCount int   `json:"return_count"`
}

// Parse extracts pagination params from query string.
// Accepts: ?offset=0&limit=20
// If all=true, returns offset=0, limit=-1 (caller should skip SQL LIMIT).
func Parse(r *http.Request) Params {
	q := r.URL.Query()

	if q.Get("all") == "true" {
		return Params{Offset: 0, Limit: -1}
	}

	offset, _ := strconv.Atoi(q.Get("offset"))
	limit, _ := strconv.Atoi(q.Get("limit"))

	if offset < 0 {
		offset = 0
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return Params{Offset: offset, Limit: limit}
}

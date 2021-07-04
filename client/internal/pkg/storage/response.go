package storage

import (
	"encoding/json"
	"strings"
)

// NewResponse updates the reference of `incoming` struct with the
// response `body` string.
func NewResponse(body string, incoming interface{}) error {
	return json.Unmarshal(
		[]byte(strings.TrimSpace(body)),
		incoming,
	)
}

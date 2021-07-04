package storage

import (
	"bytes"
	"encoding/json"
)

// NewRequest creates a network request from a copy of `outgoing` struct.
func NewRequest(outgoing interface{}) (*bytes.Buffer, error) {
	request := bytes.NewBuffer(nil)
	if err := json.NewEncoder(request).Encode(outgoing); err != nil {
		return nil, err
	}

	return request, nil
}

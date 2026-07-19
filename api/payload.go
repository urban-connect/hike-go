package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// ParsePayload decodes a JSON request body into in and then applies the
// `transform` tags of its string fields, so a payload can be normalized the
// same way request params are (e.g. `json:"device_id" transform:"lowercase"`).
func ParsePayload(r *http.Request, in any) error {
	if err := json.NewDecoder(r.Body).Decode(in); err != nil {
		return fmt.Errorf("failed to parse request body: %w", err)
	}

	return ApplyTransforms(in)
}

package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParsePayload covers a JSON body against every transform (lowercase,
// uppercase), plus a field left untouched because it carries no transform tag.
func TestParsePayload(t *testing.T) {
	t.Run("lowercase", func(t *testing.T) {
		var p struct {
			DeviceID string `json:"device_id" transform:"lowercase"`
		}

		r := httptest.NewRequest(http.MethodPost, "/devices",
			strings.NewReader(`{"device_id": "CC7B5C85C8D8"}`))

		require.NoError(t, ParsePayload(r, &p))
		assert.Equal(t, "cc7b5c85c8d8", p.DeviceID)
	})

	t.Run("uppercase", func(t *testing.T) {
		var p struct {
			Code string `json:"code" transform:"uppercase"`
		}

		r := httptest.NewRequest(http.MethodPost, "/devices",
			strings.NewReader(`{"code": "abc123"}`))

		require.NoError(t, ParsePayload(r, &p))
		assert.Equal(t, "ABC123", p.Code)
	})

	t.Run("leaves a field without a transform tag untouched", func(t *testing.T) {
		var p struct {
			DeviceID string `json:"device_id" transform:"lowercase"`
			Label    string `json:"label"`
		}

		r := httptest.NewRequest(http.MethodPost, "/devices",
			strings.NewReader(`{"device_id": "CC7B5C85C8D8", "label": "Gate A"}`))

		require.NoError(t, ParsePayload(r, &p))
		assert.Equal(t, "cc7b5c85c8d8", p.DeviceID)
		assert.Equal(t, "Gate A", p.Label)
	})
}

func TestParsePayloadRejectsInvalidJSON(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/devices", strings.NewReader(`{not json`))

	var p struct {
		DeviceID string `json:"device_id"`
	}

	assert.Error(t, ParsePayload(r, &p))
}

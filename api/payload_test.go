package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsePayload(t *testing.T) {
	type payload struct {
		DeviceID string `json:"device_id" transform:"lowercase"`
		Label    string `json:"label"`
	}

	r := httptest.NewRequest(http.MethodPost, "/devices",
		strings.NewReader(`{"device_id": "CC7B5C85C8D8", "label": "Gate A"}`))

	var p payload
	require.NoError(t, ParsePayload(r, &p))

	assert.Equal(t, "cc7b5c85c8d8", p.DeviceID)
	assert.Equal(t, "Gate A", p.Label) // untransformed
}

func TestParsePayloadRejectsInvalidJSON(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/devices", strings.NewReader(`{not json`))

	var p struct {
		DeviceID string `json:"device_id"`
	}

	assert.Error(t, ParsePayload(r, &p))
}

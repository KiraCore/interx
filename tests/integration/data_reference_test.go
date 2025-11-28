package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestQueryDataReference tests GET /api/kira/gov/data/{key}
func TestQueryDataReference(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	t.Run("query data reference", func(t *testing.T) {
		resp, err := client.Get("/api/kira/gov/data/sample_png_file", nil)
		require.NoError(t, err)
		// Data may or may not exist
		assert.True(t, resp.StatusCode == 200 || resp.StatusCode == 404,
			"Expected 200 or 404, got %d", resp.StatusCode)
	})
}

// TestQueryDataReferenceKeys tests GET /api/kira/gov/data_keys
func TestQueryDataReferenceKeys(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	t.Run("query data reference keys", func(t *testing.T) {
		resp, err := client.Get("/api/kira/gov/data_keys", nil)
		require.NoError(t, err)
		assert.True(t, resp.IsSuccess(), "Expected success, got status %d: %s", resp.StatusCode, string(resp.Body))
	})
}

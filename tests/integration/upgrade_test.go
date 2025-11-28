package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCurrentUpgradePlan tests GET /api/kira/upgrade/current_plan
func TestCurrentUpgradePlan(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	t.Run("query current upgrade plan", func(t *testing.T) {
		resp, err := client.Get("/api/kira/upgrade/current_plan", nil)
		require.NoError(t, err)
		assert.True(t, resp.IsSuccess(), "Expected success, got status %d: %s", resp.StatusCode, string(resp.Body))
	})
}

// TestNextUpgradePlan tests GET /api/kira/upgrade/next_plan
func TestNextUpgradePlan(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	t.Run("query next upgrade plan", func(t *testing.T) {
		resp, err := client.Get("/api/kira/upgrade/next_plan", nil)
		require.NoError(t, err)
		assert.True(t, resp.IsSuccess(), "Expected success, got status %d: %s", resp.StatusCode, string(resp.Body))
	})
}

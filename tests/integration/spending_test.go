package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestQuerySpendingPools tests GET /api/kira/spending-pools
func TestQuerySpendingPools(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	t.Run("query spending pools", func(t *testing.T) {
		resp, err := client.Get("/api/kira/spending-pools", nil)
		require.NoError(t, err)
		assert.True(t, resp.IsSuccess(), "Expected success, got status %d: %s", resp.StatusCode, string(resp.Body))
	})

	t.Run("query spending pool by name", func(t *testing.T) {
		resp, err := client.Get("/api/kira/spending-pools", map[string]string{
			"name": "ValidatorBasicRewardsPool",
		})
		require.NoError(t, err)
		assert.True(t, resp.IsSuccess(), "Expected success, got status %d", resp.StatusCode)
	})

	t.Run("query spending pools by account", func(t *testing.T) {
		resp, err := client.Get("/api/kira/spending-pools", map[string]string{
			"account": "kira1qveg288t6n7yar4mhhmw7kqvxzq8q08xtz56mm",
		})
		require.NoError(t, err)
		assert.True(t, resp.IsSuccess(), "Expected success, got status %d", resp.StatusCode)
	})
}

// TestQuerySpendingPoolProposals tests GET /api/kira/spending-pool-proposals
func TestQuerySpendingPoolProposals(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	t.Run("query spending pool proposals by name", func(t *testing.T) {
		resp, err := client.Get("/api/kira/spending-pool-proposals", map[string]string{
			"name": "ValidatorBasicRewardsPool",
		})
		require.NoError(t, err)
		assert.True(t, resp.IsSuccess(), "Expected success, got status %d: %s", resp.StatusCode, string(resp.Body))
	})
}

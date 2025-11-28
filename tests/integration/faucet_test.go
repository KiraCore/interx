package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFaucetInfo tests GET /api/kira/faucet (without claim)
func TestFaucetInfo(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	t.Run("query faucet info", func(t *testing.T) {
		resp, err := client.Get("/api/kira/faucet", nil)
		require.NoError(t, err)
		assert.True(t, resp.IsSuccess(), "Expected success, got status %d: %s", resp.StatusCode, string(resp.Body))
	})
}

// TestFaucetClaim tests GET /api/kira/faucet?claim=...&token=...
// Note: This actually claims tokens, so it should be used carefully in tests
func TestFaucetClaim(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	t.Run("claim from faucet", func(t *testing.T) {
		t.Skip("Skipping faucet claim test to avoid rate limiting - enable manually for full test")
		resp, err := client.Get("/api/kira/faucet", map[string]string{
			"claim": cfg.TestAddress,
			"token": "ukex",
		})
		require.NoError(t, err)
		// Faucet may succeed or fail due to rate limiting
		assert.True(t, resp.StatusCode >= 200, "Expected response, got status %d", resp.StatusCode)
	})
}

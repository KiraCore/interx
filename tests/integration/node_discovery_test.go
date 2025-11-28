package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPubP2PList tests GET /api/pub_p2p_list
func TestPubP2PList(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	t.Run("query public p2p list", func(t *testing.T) {
		resp, err := client.Get("/api/pub_p2p_list", nil)
		require.NoError(t, err)
		assert.True(t, resp.IsSuccess(), "Expected success, got status %d: %s", resp.StatusCode, string(resp.Body))
	})
}

// TestPrivP2PList tests GET /api/priv_p2p_list
func TestPrivP2PList(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	t.Run("query private p2p list", func(t *testing.T) {
		resp, err := client.Get("/api/priv_p2p_list", nil)
		require.NoError(t, err)
		assert.True(t, resp.IsSuccess(), "Expected success, got status %d: %s", resp.StatusCode, string(resp.Body))
	})
}

// TestInterxList tests GET /api/interx_list
func TestInterxList(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	t.Run("query interx list", func(t *testing.T) {
		resp, err := client.Get("/api/interx_list", nil)
		require.NoError(t, err)
		assert.True(t, resp.IsSuccess(), "Expected success, got status %d: %s", resp.StatusCode, string(resp.Body))
	})
}

// TestSnapList tests GET /api/snap_list
func TestSnapList(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	t.Run("query snap list", func(t *testing.T) {
		resp, err := client.Get("/api/snap_list", nil)
		require.NoError(t, err)
		assert.True(t, resp.IsSuccess(), "Expected success, got status %d: %s", resp.StatusCode, string(resp.Body))
	})
}

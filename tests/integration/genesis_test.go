package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestQueryGenesis tests GET /api/genesis
func TestQueryGenesis(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	t.Run("query genesis", func(t *testing.T) {
		resp, err := client.Get("/api/genesis", nil)
		require.NoError(t, err)
		assert.True(t, resp.IsSuccess(), "Expected success, got status %d: %s", resp.StatusCode, string(resp.Body))
	})
}

// TestQueryGenesisChecksum tests GET /api/gensum
func TestQueryGenesisChecksum(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	t.Run("query genesis checksum", func(t *testing.T) {
		resp, err := client.Get("/api/gensum", nil)
		require.NoError(t, err)
		assert.True(t, resp.IsSuccess(), "Expected success, got status %d: %s", resp.StatusCode, string(resp.Body))
	})
}

// TestQuerySnapshot tests GET /api/snapshot
func TestQuerySnapshot(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	t.Run("query snapshot", func(t *testing.T) {
		resp, err := client.Get("/api/snapshot", nil)
		require.NoError(t, err)
		// Snapshot may not be available on all nodes
		assert.True(t, resp.StatusCode == 200 || resp.StatusCode == 404,
			"Expected 200 or 404, got %d", resp.StatusCode)
	})
}

// TestQuerySnapshotInfo tests GET /api/snapshot_info
func TestQuerySnapshotInfo(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	t.Run("query snapshot info", func(t *testing.T) {
		resp, err := client.Get("/api/snapshot_info", nil)
		require.NoError(t, err)
		assert.True(t, resp.StatusCode == 200 || resp.StatusCode == 404,
			"Expected 200 or 404, got %d", resp.StatusCode)
	})
}

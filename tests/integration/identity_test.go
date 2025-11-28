package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIdentityRecordByID tests GET /api/kira/gov/identity_record/{id}
func TestIdentityRecordByID(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	t.Run("query identity record by id", func(t *testing.T) {
		resp, err := client.Get("/api/kira/gov/identity_record/9", nil)
		require.NoError(t, err)
		assert.True(t, resp.StatusCode == 200 || resp.StatusCode == 404,
			"Expected 200 or 404, got %d", resp.StatusCode)
	})
}

// TestIdentityRecordsByAddress tests GET /api/kira/gov/identity_records/{address}
func TestIdentityRecordsByAddress(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	t.Run("query identity records by address", func(t *testing.T) {
		resp, err := client.Get("/api/kira/gov/identity_records/"+cfg.TestAddress, nil)
		require.NoError(t, err)
		assert.True(t, resp.IsSuccess(), "Expected success, got status %d: %s", resp.StatusCode, string(resp.Body))
	})

	t.Run("query identity records with pagination", func(t *testing.T) {
		resp, err := client.Get("/api/kira/gov/identity_records/"+cfg.TestAddress, map[string]string{
			"limit":  "5",
			"offset": "0",
		})
		require.NoError(t, err)
		assert.True(t, resp.IsSuccess(), "Expected success, got status %d", resp.StatusCode)
	})
}

// TestAllIdentityRecords tests GET /api/kira/gov/all_identity_records
func TestAllIdentityRecords(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	t.Run("query all identity records", func(t *testing.T) {
		resp, err := client.Get("/api/kira/gov/all_identity_records", nil)
		require.NoError(t, err)
		assert.True(t, resp.IsSuccess(), "Expected success, got status %d: %s", resp.StatusCode, string(resp.Body))
	})

	t.Run("query all identity records with count total", func(t *testing.T) {
		resp, err := client.Get("/api/kira/gov/all_identity_records", map[string]string{
			"count_total": "true",
		})
		require.NoError(t, err)
		assert.True(t, resp.IsSuccess(), "Expected success, got status %d", resp.StatusCode)
	})
}

// TestIdentityVerifyRecord tests GET /api/kira/gov/identity_verify_record/{id}
func TestIdentityVerifyRecord(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	t.Run("query identity verify record", func(t *testing.T) {
		resp, err := client.Get("/api/kira/gov/identity_verify_record/1", nil)
		require.NoError(t, err)
		assert.True(t, resp.StatusCode == 200 || resp.StatusCode == 404,
			"Expected 200 or 404, got %d", resp.StatusCode)
	})
}

// TestIdentityVerifyRequestsByRequester tests GET /api/kira/gov/identity_verify_requests_by_requester/{address}
func TestIdentityVerifyRequestsByRequester(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	t.Run("query identity verify requests by requester", func(t *testing.T) {
		resp, err := client.Get("/api/kira/gov/identity_verify_requests_by_requester/"+cfg.TestAddress, nil)
		require.NoError(t, err)
		assert.True(t, resp.IsSuccess(), "Expected success, got status %d: %s", resp.StatusCode, string(resp.Body))
	})
}

// TestIdentityVerifyRequestsByApprover tests GET /api/kira/gov/identity_verify_requests_by_approver/{address}
func TestIdentityVerifyRequestsByApprover(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	t.Run("query identity verify requests by approver", func(t *testing.T) {
		approverAddr := "kira177lwmjyjds3cy7trers83r4pjn3dhv8zrqk9dl"
		resp, err := client.Get("/api/kira/gov/identity_verify_requests_by_approver/"+approverAddr, nil)
		require.NoError(t, err)
		assert.True(t, resp.IsSuccess(), "Expected success, got status %d: %s", resp.StatusCode, string(resp.Body))
	})
}

// TestAllIdentityVerifyRequests tests GET /api/kira/gov/all_identity_verify_requests
func TestAllIdentityVerifyRequests(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	t.Run("query all identity verify requests", func(t *testing.T) {
		resp, err := client.Get("/api/kira/gov/all_identity_verify_requests", nil)
		require.NoError(t, err)
		assert.True(t, resp.IsSuccess(), "Expected success, got status %d: %s", resp.StatusCode, string(resp.Body))
	})
}

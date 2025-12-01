package integration

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// Staking Pool Endpoint Tests - /api/kira/staking-pool
// Expected format from miro: lib/infra/dto/api_kira/query_staking_pool/
// ============================================================================

// TestStakingPoolResponseFormat validates the response matches expected miro format
func TestStakingPoolResponseFormat(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	resp, err := client.Get("/api/kira/staking-pool", map[string]string{
		"validatorAddress": cfg.ValidatorAddr,
	})
	require.NoError(t, err)
	require.True(t, resp.IsSuccess(), "Expected success, got %d: %s", resp.StatusCode, string(resp.Body))

	v, err := NewFormatValidator(t, resp.Body)
	require.NoError(t, err, "Failed to parse response")

	t.Run("id field is int (Issue #16)", func(t *testing.T) {
		if v.HasField("id") {
			v.ValidateFieldType("id", TypeInt, "Issue #16")
		}
	})

	t.Run("total_delegators field is int (Issue #16)", func(t *testing.T) {
		if v.HasField("total_delegators") {
			v.ValidateFieldType("total_delegators", TypeInt, "Issue #16")
		}
		// Check for camelCase version
		v.ValidateFieldNotExists("totalDelegators", "total_delegators", "Issue #13")
	})

	t.Run("commission is string", func(t *testing.T) {
		if v.HasField("commission") {
			v.ValidateFieldType("commission", TypeString, "Format")
		}
	})

	t.Run("slashed is string", func(t *testing.T) {
		if v.HasField("slashed") {
			v.ValidateFieldType("slashed", TypeString, "Format")
		}
	})

	t.Run("tokens is array", func(t *testing.T) {
		if v.HasField("tokens") {
			v.ValidateFieldType("tokens", TypeArray, "Format")
		}
	})

	t.Run("voting_power is array of coins", func(t *testing.T) {
		if v.HasField("voting_power") {
			v.ValidateFieldType("voting_power", TypeArray, "Format")
		}
		// Check for camelCase version
		v.ValidateFieldNotExists("votingPower", "voting_power", "Issue #13")
	})
}

// TestQueryStakingPool tests GET /api/kira/staking-pool
func TestQueryStakingPool(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	t.Run("query staking pool by validator address", func(t *testing.T) {
		resp, err := client.Get("/api/kira/staking-pool", map[string]string{
			"validatorAddress": cfg.ValidatorAddr,
		})
		require.NoError(t, err)
		assert.True(t, resp.IsSuccess(), "Expected success, got status %d: %s", resp.StatusCode, string(resp.Body))

		var result map[string]interface{}
		json.Unmarshal(resp.Body, &result)
		t.Logf("Staking pool response keys: %v", getMapKeys(result))
	})
}

// ============================================================================
// Delegations Endpoint Tests - /api/kira/delegations
// Expected format from miro: lib/infra/dto/api_kira/query_delegations/
// ============================================================================

// TestDelegationsResponseFormat validates the response matches expected miro format
func TestDelegationsResponseFormat(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	resp, err := client.Get("/api/kira/delegations", map[string]string{
		"delegatorAddress": cfg.DelegatorAddr,
	})
	require.NoError(t, err)
	require.True(t, resp.IsSuccess(), "Expected success, got %d: %s", resp.StatusCode, string(resp.Body))

	v, err := NewFormatValidator(t, resp.Body)
	require.NoError(t, err, "Failed to parse response")

	t.Run("delegations array exists", func(t *testing.T) {
		v.ValidateFieldExists("delegations", "Format")
		v.ValidateFieldType("delegations", TypeArray, "Format")
	})

	t.Run("pagination object exists", func(t *testing.T) {
		if !v.HasField("pagination") {
			t.Log("Note: 'pagination' field not present")
		}
	})

	// Check delegation structure if we have data
	if v.GetArrayLength("delegations") > 0 {
		t.Run("delegation has validator_info (Issue #13)", func(t *testing.T) {
			raw := v.GetRaw()
			delegations := raw["delegations"].([]interface{})
			delegation := delegations[0].(map[string]interface{})

			if _, exists := delegation["validator_info"]; !exists {
				if _, camelExists := delegation["validatorInfo"]; camelExists {
					t.Error("Issue #13: Using 'validatorInfo' instead of 'validator_info'")
				}
			}
		})

		t.Run("delegation has pool_info (Issue #13)", func(t *testing.T) {
			raw := v.GetRaw()
			delegations := raw["delegations"].([]interface{})
			delegation := delegations[0].(map[string]interface{})

			if _, exists := delegation["pool_info"]; !exists {
				if _, camelExists := delegation["poolInfo"]; camelExists {
					t.Error("Issue #13: Using 'poolInfo' instead of 'pool_info'")
				}
			}
		})

		t.Run("pool_info.id is int (Issue #16)", func(t *testing.T) {
			raw := v.GetRaw()
			delegations := raw["delegations"].([]interface{})
			delegation := delegations[0].(map[string]interface{})

			if poolInfo, exists := delegation["pool_info"]; exists {
				pool := poolInfo.(map[string]interface{})
				if id, idExists := pool["id"]; idExists {
					switch id.(type) {
					case float64:
						// OK
					case string:
						t.Error("Issue #16: 'pool_info.id' should be int, got string")
					}
				}
			}
		})
	}
}

// TestQueryDelegations tests GET /api/kira/delegations
func TestQueryDelegations(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	t.Run("query delegations by delegator address", func(t *testing.T) {
		resp, err := client.Get("/api/kira/delegations", map[string]string{
			"delegatorAddress": cfg.DelegatorAddr,
		})
		require.NoError(t, err)
		assert.True(t, resp.IsSuccess(), "Expected success, got status %d: %s", resp.StatusCode, string(resp.Body))

		var result map[string]interface{}
		json.Unmarshal(resp.Body, &result)

		if delegations, ok := result["delegations"].([]interface{}); ok {
			t.Logf("Found %d delegations", len(delegations))
		}
	})

	t.Run("query delegations with pagination", func(t *testing.T) {
		resp, err := client.Get("/api/kira/delegations", map[string]string{
			"delegatorAddress": cfg.DelegatorAddr,
			"limit":            "5",
			"offset":           "0",
		})
		require.NoError(t, err)
		assert.True(t, resp.IsSuccess(), "Expected success, got status %d", resp.StatusCode)
	})

	t.Run("query delegations with count total", func(t *testing.T) {
		resp, err := client.Get("/api/kira/delegations", map[string]string{
			"delegatorAddress": cfg.DelegatorAddr,
			"countTotal":       "true",
		})
		require.NoError(t, err)
		assert.True(t, resp.IsSuccess(), "Expected success, got status %d", resp.StatusCode)
	})
}

// ============================================================================
// Undelegations Endpoint Tests - /api/kira/undelegations
// ============================================================================

// TestQueryUndelegations tests GET /api/kira/undelegations
func TestQueryUndelegations(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	t.Run("query undelegations", func(t *testing.T) {
		resp, err := client.Get("/api/kira/undelegations", map[string]string{
			"undelegatorAddress": cfg.TestAddress,
		})
		require.NoError(t, err)
		assert.True(t, resp.IsSuccess(), "Expected success, got status %d: %s", resp.StatusCode, string(resp.Body))
	})

	t.Run("query undelegations with count total", func(t *testing.T) {
		resp, err := client.Get("/api/kira/undelegations", map[string]string{
			"undelegatorAddress": cfg.TestAddress,
			"countTotal":         "true",
		})
		require.NoError(t, err)
		assert.True(t, resp.IsSuccess(), "Expected success, got status %d", resp.StatusCode)
	})

	t.Run("query undelegations with pagination", func(t *testing.T) {
		resp, err := client.Get("/api/kira/undelegations", map[string]string{
			"undelegatorAddress": cfg.TestAddress,
			"limit":              "5",
			"offset":             "0",
		})
		require.NoError(t, err)
		assert.True(t, resp.IsSuccess(), "Expected success, got status %d", resp.StatusCode)
	})
}

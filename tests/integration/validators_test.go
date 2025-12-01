package integration

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// Validator Endpoint Tests - /api/valopers
// Expected format from miro: lib/infra/dto/api/query_validators/response/
// ============================================================================

// TestValidatorsResponseFormat validates the response matches expected miro format
// Issues: #13 (snake_case), #16 (types)
func TestValidatorsResponseFormat(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	resp, err := client.Get("/api/valopers", map[string]string{
		"all": "true",
	})
	require.NoError(t, err)
	require.True(t, resp.IsSuccess(), "Expected success, got %d: %s", resp.StatusCode, string(resp.Body))

	v, err := NewFormatValidator(t, resp.Body)
	require.NoError(t, err, "Failed to parse response")

	t.Run("validators array exists", func(t *testing.T) {
		v.ValidateFieldExists("validators", "Format")
		v.ValidateFieldType("validators", TypeArray, "Format")
	})

	t.Run("waiting array exists", func(t *testing.T) {
		if !v.HasField("waiting") {
			t.Log("Note: 'waiting' field not present in response")
		} else {
			v.ValidateFieldType("waiting", TypeArray, "Format")
		}
	})

	t.Run("status object exists with int counts", func(t *testing.T) {
		if !v.HasField("status") {
			t.Log("Note: 'status' field not present (may require status_only=false)")
			return
		}

		nested, ok := v.GetNestedValidator("status")
		if !ok {
			return
		}

		// miro expects these as integers
		intFields := []string{
			"active_validators",
			"paused_validators",
			"inactive_validators",
			"jailed_validators",
			"total_validators",
			"waiting_validators",
		}

		for _, field := range intFields {
			if nested.HasField(field) {
				nested.ValidateFieldType(field, TypeInt, "Issue #16")
			}
		}

		// Check snake_case naming
		nested.ValidateSnakeCase("Issue #13")
	})

	// Check validator structure if we have validators
	if v.GetArrayLength("validators") > 0 {
		t.Run("validator fields use snake_case (Issue #13)", func(t *testing.T) {
			raw := v.GetRaw()
			validators := raw["validators"].([]interface{})
			validator := validators[0].(map[string]interface{})

			expectedSnakeCaseFields := []string{
				"mischance_confidence",
				"start_height",
				"inactive_until",
				"last_present_block",
				"missed_blocks_counter",
				"produced_blocks_counter",
				"staking_pool_id",
				"staking_pool_status",
				"validator_node_id",
				"sentry_node_id",
			}

			// Check for camelCase versions (would be a bug)
			camelCaseVersions := map[string]string{
				"mischanceConfidence":   "mischance_confidence",
				"startHeight":           "start_height",
				"inactiveUntil":         "inactive_until",
				"lastPresentBlock":      "last_present_block",
				"missedBlocksCounter":   "missed_blocks_counter",
				"producedBlocksCounter": "produced_blocks_counter",
				"stakingPoolId":         "staking_pool_id",
				"stakingPoolStatus":     "staking_pool_status",
				"validatorNodeId":       "validator_node_id",
				"sentryNodeId":          "sentry_node_id",
			}

			for camelCase, snakeCase := range camelCaseVersions {
				if _, exists := validator[camelCase]; exists {
					if _, snakeExists := validator[snakeCase]; !snakeExists {
						t.Errorf("Issue #13: Using '%s' instead of '%s'", camelCase, snakeCase)
					}
				}
			}

			// Verify snake_case fields exist
			for _, field := range expectedSnakeCaseFields {
				if _, exists := validator[field]; !exists {
					// Not all fields are always present
					t.Logf("Note: Field '%s' not present in validator", field)
				}
			}
		})

		t.Run("validator fields are strings", func(t *testing.T) {
			raw := v.GetRaw()
			validators := raw["validators"].([]interface{})
			validator := validators[0].(map[string]interface{})

			// These should all be strings according to miro
			stringFields := []string{"top", "address", "valkey", "pubkey", "proposer", "moniker", "status", "rank", "streak", "mischance"}

			for _, field := range stringFields {
				if val, exists := validator[field]; exists {
					if _, ok := val.(string); !ok {
						t.Errorf("Issue #16: Field '%s' should be string, got %T", field, val)
					}
				}
			}
		})
	}
}

// TestQueryValidators tests GET /api/valopers
func TestQueryValidators(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	t.Run("query validators with status only", func(t *testing.T) {
		resp, err := client.Get("/api/valopers", map[string]string{
			"status_only": "true",
		})
		require.NoError(t, err)
		assert.True(t, resp.IsSuccess(), "Expected success, got status %d: %s", resp.StatusCode, string(resp.Body))

		var result map[string]interface{}
		json.Unmarshal(resp.Body, &result)

		// With status_only=true, should have status object
		if _, exists := result["status"]; !exists {
			t.Log("Note: status_only=true did not return status object")
		}
	})

	t.Run("query all validators", func(t *testing.T) {
		resp, err := client.Get("/api/valopers", map[string]string{
			"all": "true",
		})
		require.NoError(t, err)
		assert.True(t, resp.IsSuccess(), "Expected success, got status %d", resp.StatusCode)

		var result map[string]interface{}
		json.Unmarshal(resp.Body, &result)

		validators, ok := result["validators"].([]interface{})
		if ok {
			t.Logf("Found %d validators", len(validators))
		}
	})

	t.Run("query validator by address", func(t *testing.T) {
		resp, err := client.Get("/api/valopers", map[string]string{
			"address": cfg.ValidatorAddr,
		})
		require.NoError(t, err)
		assert.True(t, resp.IsSuccess(), "Expected success, got status %d", resp.StatusCode)

		var result map[string]interface{}
		json.Unmarshal(resp.Body, &result)

		validators, ok := result["validators"].([]interface{})
		if ok && len(validators) > 0 {
			// Should filter to the specific address
			validator := validators[0].(map[string]interface{})
			if addr, exists := validator["address"]; exists {
				if addr != cfg.ValidatorAddr {
					t.Errorf("Address filter not working: expected %s, got %s", cfg.ValidatorAddr, addr)
				}
			}
		}
	})

	t.Run("query validators with count total", func(t *testing.T) {
		resp, err := client.Get("/api/valopers", map[string]string{
			"count_total": "true",
		})
		require.NoError(t, err)
		assert.True(t, resp.IsSuccess(), "Expected success, got status %d", resp.StatusCode)
	})

	t.Run("query validators by status filter", func(t *testing.T) {
		resp, err := client.Get("/api/valopers", map[string]string{
			"status": "ACTIVE",
		})
		require.NoError(t, err)
		assert.True(t, resp.IsSuccess(), "Expected success, got status %d", resp.StatusCode)

		var result map[string]interface{}
		json.Unmarshal(resp.Body, &result)

		validators, ok := result["validators"].([]interface{})
		if ok && len(validators) > 0 {
			// All validators should have ACTIVE status
			for i, v := range validators {
				validator := v.(map[string]interface{})
				if status, exists := validator["status"]; exists {
					if status != "ACTIVE" {
						t.Errorf("Status filter not working: validator %d has status '%v', expected 'ACTIVE'", i, status)
						break
					}
				}
			}
		}
	})

	t.Run("pagination works", func(t *testing.T) {
		resp1, err := client.Get("/api/valopers", map[string]string{"limit": "5"})
		require.NoError(t, err)
		assert.True(t, resp1.IsSuccess())

		resp2, err := client.Get("/api/valopers", map[string]string{"limit": "5", "offset": "5"})
		require.NoError(t, err)
		assert.True(t, resp2.IsSuccess())

		var result1, result2 map[string]interface{}
		json.Unmarshal(resp1.Body, &result1)
		json.Unmarshal(resp2.Body, &result2)

		v1, _ := result1["validators"].([]interface{})
		v2, _ := result2["validators"].([]interface{})

		if len(v1) > 0 && len(v2) > 0 {
			val1 := v1[0].(map[string]interface{})
			val2 := v2[0].(map[string]interface{})
			if val1["address"] == val2["address"] {
				t.Error("Offset not working - same validators returned")
			}
		}
	})
}

// TestQueryValidatorInfos tests GET /api/valoperinfos
func TestQueryValidatorInfos(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	t.Run("query validator infos", func(t *testing.T) {
		resp, err := client.Get("/api/valoperinfos", nil)
		require.NoError(t, err)
		assert.True(t, resp.IsSuccess(), "Expected success, got status %d: %s", resp.StatusCode, string(resp.Body))
	})
}

// TestQueryConsensus tests GET /api/consensus
func TestQueryConsensus(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	t.Run("query consensus", func(t *testing.T) {
		resp, err := client.Get("/api/consensus", nil)
		require.NoError(t, err)
		assert.True(t, resp.IsSuccess(), "Expected success, got status %d: %s", resp.StatusCode, string(resp.Body))

		var result map[string]interface{}
		json.Unmarshal(resp.Body, &result)
		t.Logf("Consensus response keys: %v", getMapKeys(result))
	})
}

// TestDumpConsensusState tests GET /api/dump_consensus_state
func TestDumpConsensusState(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	t.Run("dump consensus state", func(t *testing.T) {
		resp, err := client.Get("/api/dump_consensus_state", nil)
		require.NoError(t, err)
		assert.True(t, resp.IsSuccess(), "Expected success, got status %d: %s", resp.StatusCode, string(resp.Body))
	})
}

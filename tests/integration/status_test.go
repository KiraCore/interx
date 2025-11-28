package integration

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// Status Endpoint Tests - /api/status
// Expected format from miro: lib/infra/dto/api/query_interx_status/
// ============================================================================

// TestInterxStatusResponseFormat validates the response matches expected miro format
// Issues: #13 (snake_case), #16 (types)
func TestInterxStatusResponseFormat(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	resp, err := client.Get("/api/status", nil)
	require.NoError(t, err)
	require.True(t, resp.IsSuccess(), "Expected success, got %d: %s", resp.StatusCode, string(resp.Body))

	v, err := NewFormatValidator(t, resp.Body)
	require.NoError(t, err, "Failed to parse response")

	t.Run("id field exists", func(t *testing.T) {
		v.ValidateFieldExists("id", "Format")
	})

	t.Run("interx_info object exists", func(t *testing.T) {
		v.ValidateFieldExists("interx_info", "Format")
		v.ValidateFieldType("interx_info", TypeObject, "Format")
	})

	t.Run("node_info object exists", func(t *testing.T) {
		v.ValidateFieldExists("node_info", "Format")
		v.ValidateFieldType("node_info", TypeObject, "Format")
	})

	t.Run("sync_info object exists", func(t *testing.T) {
		v.ValidateFieldExists("sync_info", "Format")
		v.ValidateFieldType("sync_info", TypeObject, "Format")
	})

	t.Run("validator_info object exists", func(t *testing.T) {
		v.ValidateFieldExists("validator_info", "Format")
		v.ValidateFieldType("validator_info", TypeObject, "Format")
	})

	t.Run("interx_info fields use snake_case (Issue #13)", func(t *testing.T) {
		if nested, ok := v.GetNestedValidator("interx_info"); ok {
			nested.ValidateSnakeCase("Issue #13")

			// Validate specific expected fields
			expectedFields := []string{
				"catching_up",
				"chain_id",
				"genesis_checksum",
				"kira_addr",
				"kira_pub_key",
				"latest_block_height",
				"moniker",
				"version",
			}

			for _, field := range expectedFields {
				if !nested.HasField(field) {
					// Check for camelCase version
					camelCase := snakeToCamel(field)
					if nested.HasField(camelCase) {
						t.Errorf("Issue #13: Using '%s' instead of '%s'", camelCase, field)
					}
				}
			}
		}
	})

	t.Run("interx_info.catching_up is bool (Issue #16)", func(t *testing.T) {
		if nested, ok := v.GetNestedValidator("interx_info"); ok {
			if nested.HasField("catching_up") {
				nested.ValidateFieldType("catching_up", TypeBool, "Issue #16")
			}
		}
	})

	t.Run("sync_info fields use snake_case (Issue #13)", func(t *testing.T) {
		if nested, ok := v.GetNestedValidator("sync_info"); ok {
			nested.ValidateSnakeCase("Issue #13")

			// Expected snake_case fields
			expectedFields := []string{
				"earliest_app_hash",
				"earliest_block_hash",
				"earliest_block_height",
				"earliest_block_time",
				"latest_app_hash",
				"latest_block_hash",
				"latest_block_height",
				"latest_block_time",
			}

			for _, field := range expectedFields {
				if !nested.HasField(field) {
					camelCase := snakeToCamel(field)
					if nested.HasField(camelCase) {
						t.Errorf("Issue #13: Using '%s' instead of '%s'", camelCase, field)
					}
				}
			}
		}
	})

	t.Run("validator_info fields use snake_case (Issue #13)", func(t *testing.T) {
		if nested, ok := v.GetNestedValidator("validator_info"); ok {
			nested.ValidateSnakeCase("Issue #13")

			// Check for pub_key vs pubKey
			if !nested.HasField("pub_key") && nested.HasField("pubKey") {
				t.Error("Issue #13: Using 'pubKey' instead of 'pub_key'")
			}

			// Check for voting_power vs votingPower
			if !nested.HasField("voting_power") && nested.HasField("votingPower") {
				t.Error("Issue #13: Using 'votingPower' instead of 'voting_power'")
			}
		}
	})
}

// TestInterxStatus tests GET /api/status
func TestInterxStatus(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	t.Run("query interx status", func(t *testing.T) {
		resp, err := client.Get("/api/status", nil)
		require.NoError(t, err)
		assert.True(t, resp.IsSuccess(), "Expected success, got status %d: %s", resp.StatusCode, string(resp.Body))

		var result map[string]interface{}
		err = resp.JSON(&result)
		require.NoError(t, err)

		// Verify expected fields exist
		assert.Contains(t, result, "interx_info", "Response should contain interx_info field")
	})
}

// ============================================================================
// Dashboard Endpoint Tests - /api/dashboard
// Expected format from miro: lib/infra/dto/api/dashboard/
// ============================================================================

// TestDashboardResponseFormat validates the response matches expected miro format
func TestDashboardResponseFormat(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	resp, err := client.Get("/api/dashboard", nil)
	require.NoError(t, err)
	require.True(t, resp.IsSuccess(), "Expected success, got %d: %s", resp.StatusCode, string(resp.Body))

	v, err := NewFormatValidator(t, resp.Body)
	require.NoError(t, err, "Failed to parse response")

	t.Run("consensus_health field exists", func(t *testing.T) {
		v.ValidateFieldExists("consensus_health", "Format")
	})

	t.Run("current_block_validator object exists", func(t *testing.T) {
		v.ValidateFieldExists("current_block_validator", "Format")
		v.ValidateFieldType("current_block_validator", TypeObject, "Format")
	})

	t.Run("validators object exists with int counts", func(t *testing.T) {
		v.ValidateFieldExists("validators", "Format")

		if nested, ok := v.GetNestedValidator("validators"); ok {
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

			nested.ValidateSnakeCase("Issue #13")
		}
	})

	t.Run("blocks object exists with correct types", func(t *testing.T) {
		v.ValidateFieldExists("blocks", "Format")

		if nested, ok := v.GetNestedValidator("blocks"); ok {
			// Integer fields
			intFields := []string{"current_height", "since_genesis", "pending_transactions", "current_transactions"}
			for _, field := range intFields {
				if nested.HasField(field) {
					nested.ValidateFieldType(field, TypeInt, "Issue #16")
				}
			}

			// Float fields
			floatFields := []string{"latest_time", "average_time"}
			for _, field := range floatFields {
				if nested.HasField(field) {
					// Can be int or float
					raw := nested.GetRaw()
					if val, exists := raw[field]; exists {
						switch val.(type) {
						case float64:
							// OK
						case string:
							t.Errorf("Issue #16: Field '%s' should be number, got string", field)
						}
					}
				}
			}

			nested.ValidateSnakeCase("Issue #13")
		}
	})

	t.Run("proposals object exists with int counts", func(t *testing.T) {
		v.ValidateFieldExists("proposals", "Format")

		if nested, ok := v.GetNestedValidator("proposals"); ok {
			intFields := []string{"total", "active", "enacting", "finished", "successful"}
			for _, field := range intFields {
				if nested.HasField(field) {
					nested.ValidateFieldType(field, TypeInt, "Issue #16")
				}
			}
		}
	})
}

// TestKiraDashboard tests GET /api/dashboard
func TestKiraDashboard(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	t.Run("query dashboard", func(t *testing.T) {
		resp, err := client.Get("/api/dashboard", nil)
		require.NoError(t, err)
		assert.True(t, resp.IsSuccess(), "Expected success, got status %d: %s", resp.StatusCode, string(resp.Body))

		var result map[string]interface{}
		json.Unmarshal(resp.Body, &result)
		t.Logf("Dashboard response keys: %v", getMapKeys(result))
	})
}

// ============================================================================
// Kira Status Endpoint Tests - /api/kira/status
// ============================================================================

// TestKiraStatus tests GET /api/kira/status
func TestKiraStatus(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	t.Run("query kira status", func(t *testing.T) {
		resp, err := client.Get("/api/kira/status", nil)
		require.NoError(t, err)
		assert.True(t, resp.IsSuccess(), "Expected success, got status %d: %s", resp.StatusCode, string(resp.Body))

		var result map[string]interface{}
		err = resp.JSON(&result)
		require.NoError(t, err)

		t.Logf("Kira status response keys: %v", getMapKeys(result))
	})
}

// ============================================================================
// Other Status-Related Endpoints
// ============================================================================

// TestTotalSupply tests GET /api/kira/supply
func TestTotalSupply(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	t.Run("query total supply", func(t *testing.T) {
		resp, err := client.Get("/api/kira/supply", nil)
		require.NoError(t, err)
		assert.True(t, resp.IsSuccess(), "Expected success, got status %d: %s", resp.StatusCode, string(resp.Body))
	})
}

// TestRPCMethods tests GET /api/rpc_methods
func TestRPCMethods(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	t.Run("query rpc methods", func(t *testing.T) {
		resp, err := client.Get("/api/rpc_methods", nil)
		require.NoError(t, err)
		assert.True(t, resp.IsSuccess(), "Expected success, got status %d: %s", resp.StatusCode, string(resp.Body))
	})
}

// TestMetadata tests GET /api/metadata
func TestMetadata(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	t.Run("query interx metadata", func(t *testing.T) {
		resp, err := client.Get("/api/metadata", nil)
		require.NoError(t, err)
		assert.True(t, resp.IsSuccess(), "Expected success, got status %d: %s", resp.StatusCode, string(resp.Body))
	})
}

// TestKiraMetadata tests GET /api/kira/metadata
func TestKiraMetadata(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	t.Run("query kira metadata", func(t *testing.T) {
		resp, err := client.Get("/api/kira/metadata", nil)
		require.NoError(t, err)
		assert.True(t, resp.IsSuccess(), "Expected success, got status %d: %s", resp.StatusCode, string(resp.Body))
	})
}

// TestAddrBook tests GET /api/addrbook
func TestAddrBook(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	t.Run("query address book", func(t *testing.T) {
		resp, err := client.Get("/api/addrbook", nil)
		require.NoError(t, err)
		assert.True(t, resp.IsSuccess(), "Expected success, got status %d: %s", resp.StatusCode, string(resp.Body))
	})
}

// TestNetInfo tests GET /api/net_info
func TestNetInfo(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	t.Run("query net info", func(t *testing.T) {
		resp, err := client.Get("/api/net_info", nil)
		require.NoError(t, err)
		assert.True(t, resp.IsSuccess(), "Expected success, got status %d: %s", resp.StatusCode, string(resp.Body))
	})
}

// ============================================================================
// Helper functions
// ============================================================================

// snakeToCamel converts snake_case to camelCase
func snakeToCamel(s string) string {
	result := ""
	capitalizeNext := false
	for i, r := range s {
		if r == '_' {
			capitalizeNext = true
			continue
		}
		if capitalizeNext && i > 0 {
			result += string(r - 32) // uppercase
			capitalizeNext = false
		} else {
			result += string(r)
		}
	}
	return result
}

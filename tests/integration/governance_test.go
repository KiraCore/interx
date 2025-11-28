package integration

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// Governance Endpoint Tests
// Expected format from miro: lib/infra/dto/api_kira/query_network_properties/
// ============================================================================

// TestNetworkPropertiesResponseFormat validates the response matches expected miro format
func TestNetworkPropertiesResponseFormat(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	resp, err := client.Get("/api/kira/gov/network_properties", nil)
	require.NoError(t, err)
	require.True(t, resp.IsSuccess(), "Expected success, got %d: %s", resp.StatusCode, string(resp.Body))

	v, err := NewFormatValidator(t, resp.Body)
	require.NoError(t, err, "Failed to parse response")

	t.Run("properties object exists", func(t *testing.T) {
		v.ValidateFieldExists("properties", "Format")
		v.ValidateFieldType("properties", TypeObject, "Format")
	})

	t.Run("properties fields use snake_case (Issue #13)", func(t *testing.T) {
		if nested, ok := v.GetNestedValidator("properties"); ok {
			nested.ValidateSnakeCase("Issue #13")

			// Check key properties exist
			expectedFields := []string{
				"min_tx_fee",
				"vote_quorum",
				"min_validators",
				"unstaking_period",
				"max_delegators",
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

	t.Run("boolean properties are bools (Issue #16)", func(t *testing.T) {
		if nested, ok := v.GetNestedValidator("properties"); ok {
			boolFields := []string{
				"enable_foreign_fee_payments",
				"enable_token_blacklist",
				"enable_token_whitelist",
			}

			for _, field := range boolFields {
				if nested.HasField(field) {
					nested.ValidateFieldType(field, TypeBool, "Issue #16")
				}
			}
		}
	})
}

// TestQueryNetworkProperties tests GET /api/kira/gov/network_properties
func TestQueryNetworkProperties(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	t.Run("query network properties", func(t *testing.T) {
		resp, err := client.Get("/api/kira/gov/network_properties", nil)
		require.NoError(t, err)
		assert.True(t, resp.IsSuccess(), "Expected success, got status %d: %s", resp.StatusCode, string(resp.Body))

		var result map[string]interface{}
		json.Unmarshal(resp.Body, &result)

		if props, ok := result["properties"].(map[string]interface{}); ok {
			t.Logf("Network properties has %d fields", len(props))
		}
	})
}

// ============================================================================
// Execution Fee Endpoint Tests
// ============================================================================

// TestExecutionFeeResponseFormat validates the response format
func TestExecutionFeeResponseFormat(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	resp, err := client.Get("/api/kira/gov/execution_fee", map[string]string{
		"message": "MsgSend",
	})
	require.NoError(t, err)
	require.True(t, resp.IsSuccess(), "Expected success, got %d: %s", resp.StatusCode, string(resp.Body))

	v, err := NewFormatValidator(t, resp.Body)
	require.NoError(t, err, "Failed to parse response")

	t.Run("fee object exists", func(t *testing.T) {
		v.ValidateFieldExists("fee", "Format")
	})

	t.Run("fee fields use snake_case (Issue #13)", func(t *testing.T) {
		if nested, ok := v.GetNestedValidator("fee"); ok {
			nested.ValidateSnakeCase("Issue #13")

			expectedFields := []string{
				"default_parameters",
				"execution_fee",
				"failure_fee",
				"timeout",
				"transaction_type",
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
}

// TestQueryExecutionFee tests GET /api/kira/gov/execution_fee
func TestQueryExecutionFee(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	t.Run("query execution fee for MsgSend", func(t *testing.T) {
		resp, err := client.Get("/api/kira/gov/execution_fee", map[string]string{
			"message": "MsgSend",
		})
		require.NoError(t, err)
		assert.True(t, resp.IsSuccess(), "Expected success, got status %d: %s", resp.StatusCode, string(resp.Body))
	})

	t.Run("query execution fee without message param", func(t *testing.T) {
		resp, err := client.Get("/api/kira/gov/execution_fee", nil)
		require.NoError(t, err)
		// May return error or default fee
		t.Logf("Response status: %d", resp.StatusCode)
	})
}

// TestQueryExecutionFees tests GET /api/kira/gov/execution_fees
func TestQueryExecutionFees(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	t.Run("query all execution fees", func(t *testing.T) {
		resp, err := client.Get("/api/kira/gov/execution_fees", nil)
		require.NoError(t, err)
		assert.True(t, resp.IsSuccess(), "Expected success, got status %d: %s", resp.StatusCode, string(resp.Body))
	})
}

// ============================================================================
// Proposals Endpoint Tests
// ============================================================================

// TestQueryProposals tests GET /api/kira/gov/proposals
func TestQueryProposals(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	t.Run("query all proposals", func(t *testing.T) {
		resp, err := client.Get("/api/kira/gov/proposals", map[string]string{
			"all": "true",
		})
		require.NoError(t, err)
		assert.True(t, resp.IsSuccess(), "Expected success, got status %d: %s", resp.StatusCode, string(resp.Body))
	})

	t.Run("query proposals with reverse", func(t *testing.T) {
		resp, err := client.Get("/api/kira/gov/proposals", map[string]string{
			"reverse": "true",
		})
		require.NoError(t, err)
		assert.True(t, resp.IsSuccess(), "Expected success, got status %d", resp.StatusCode)
	})
}

// TestQueryProposalById tests GET /api/kira/gov/proposals/{id}
func TestQueryProposalById(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	t.Run("query proposal by valid id", func(t *testing.T) {
		resp, err := client.Get("/api/kira/gov/proposals/1", nil)
		require.NoError(t, err)
		// May be 200 or 404 depending on if proposal exists
		assert.True(t, resp.StatusCode == 200 || resp.StatusCode == 404,
			"Expected 200 or 404, got %d", resp.StatusCode)
	})

	t.Run("query proposal by invalid id", func(t *testing.T) {
		resp, err := client.Get("/api/kira/gov/proposals/invalid", nil)
		require.NoError(t, err)
		// Should return error
		t.Logf("Invalid proposal id response: %d", resp.StatusCode)
	})
}

// TestQueryVoters tests GET /api/kira/gov/voters
func TestQueryVoters(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	t.Run("query voters", func(t *testing.T) {
		resp, err := client.Get("/api/kira/gov/voters", nil)
		require.NoError(t, err)
		assert.True(t, resp.IsSuccess(), "Expected success, got status %d: %s", resp.StatusCode, string(resp.Body))
	})
}

// TestQueryVotes tests GET /api/kira/gov/votes/{proposal_id}
func TestQueryVotes(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	t.Run("query votes for proposal", func(t *testing.T) {
		resp, err := client.Get("/api/kira/gov/votes/1", nil)
		require.NoError(t, err)
		// May be 200 or 404 depending on if proposal exists
		t.Logf("Votes response status: %d", resp.StatusCode)
	})
}

// ============================================================================
// Data Keys Endpoint Tests
// ============================================================================

// TestQueryDataKeys tests GET /api/kira/gov/data_keys
func TestQueryDataKeys(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	t.Run("query data keys", func(t *testing.T) {
		resp, err := client.Get("/api/kira/gov/data_keys", nil)
		require.NoError(t, err)
		assert.True(t, resp.IsSuccess(), "Expected success, got status %d: %s", resp.StatusCode, string(resp.Body))
	})
}

// TestQueryData tests GET /api/kira/gov/data/{key}
func TestQueryData(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	t.Run("query data by key", func(t *testing.T) {
		resp, err := client.Get("/api/kira/gov/data/test_key", nil)
		require.NoError(t, err)
		// May return data or not found
		t.Logf("Data query response status: %d", resp.StatusCode)
	})
}

// ============================================================================
// Permissions and Roles Endpoint Tests
// ============================================================================

// TestQueryPermissionsByAddress tests GET /api/kira/gov/permissions_by_address/{address}
func TestQueryPermissionsByAddress(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	t.Run("query permissions by address", func(t *testing.T) {
		resp, err := client.Get("/api/kira/gov/permissions_by_address/"+cfg.TestAddress, nil)
		require.NoError(t, err)
		assert.True(t, resp.IsSuccess(), "Expected success, got status %d: %s", resp.StatusCode, string(resp.Body))
	})
}

// TestQueryAllRoles tests GET /api/kira/gov/all_roles
func TestQueryAllRoles(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	t.Run("query all roles", func(t *testing.T) {
		resp, err := client.Get("/api/kira/gov/all_roles", nil)
		require.NoError(t, err)
		assert.True(t, resp.IsSuccess(), "Expected success, got status %d: %s", resp.StatusCode, string(resp.Body))
	})
}

// TestQueryRolesByAddress tests GET /api/kira/gov/roles_by_address/{address}
func TestQueryRolesByAddress(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	t.Run("query roles by address", func(t *testing.T) {
		resp, err := client.Get("/api/kira/gov/roles_by_address/"+cfg.TestAddress, nil)
		require.NoError(t, err)
		assert.True(t, resp.IsSuccess(), "Expected success, got status %d: %s", resp.StatusCode, string(resp.Body))
	})
}

// ============================================================================
// Identity Endpoint Tests
// ============================================================================

// TestQueryIdentityRecords tests GET /api/kira/gov/identity_records/{address}
func TestQueryIdentityRecords(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	t.Run("query identity records by address", func(t *testing.T) {
		resp, err := client.Get("/api/kira/gov/identity_records/"+cfg.TestAddress, nil)
		require.NoError(t, err)
		assert.True(t, resp.IsSuccess(), "Expected success, got status %d: %s", resp.StatusCode, string(resp.Body))
	})
}

// TestQueryIdentityRecord tests GET /api/kira/gov/identity_record/{id}
func TestQueryIdentityRecord(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	t.Run("query identity record by id", func(t *testing.T) {
		resp, err := client.Get("/api/kira/gov/identity_record/1", nil)
		require.NoError(t, err)
		// May return data or not found
		t.Logf("Identity record response status: %d", resp.StatusCode)
	})
}

// TestQueryIdentityVerifyRequestsByApprover tests the approver endpoint
func TestQueryIdentityVerifyRequestsByApprover(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	t.Run("query verify requests by approver", func(t *testing.T) {
		resp, err := client.Get("/api/kira/gov/identity_verify_requests_by_approver/"+cfg.TestAddress, nil)
		require.NoError(t, err)
		t.Logf("Approver verify requests response status: %d", resp.StatusCode)
	})
}

// TestQueryIdentityVerifyRequestsByRequester tests the requester endpoint
func TestQueryIdentityVerifyRequestsByRequester(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	t.Run("query verify requests by requester", func(t *testing.T) {
		resp, err := client.Get("/api/kira/gov/identity_verify_requests_by_requester/"+cfg.TestAddress, nil)
		require.NoError(t, err)
		t.Logf("Requester verify requests response status: %d", resp.StatusCode)
	})
}

// TestQueryAllIdentityRecords tests GET /api/kira/gov/all_identity_records
func TestQueryAllIdentityRecords(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	t.Run("query all identity records", func(t *testing.T) {
		resp, err := client.Get("/api/kira/gov/all_identity_records", nil)
		require.NoError(t, err)
		assert.True(t, resp.IsSuccess(), "Expected success, got status %d: %s", resp.StatusCode, string(resp.Body))
	})
}

// TestQueryAllIdentityVerifyRequests tests GET /api/kira/gov/all_identity_verify_requests
func TestQueryAllIdentityVerifyRequests(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	t.Run("query all identity verify requests", func(t *testing.T) {
		resp, err := client.Get("/api/kira/gov/all_identity_verify_requests", nil)
		require.NoError(t, err)
		assert.True(t, resp.IsSuccess(), "Expected success, got status %d: %s", resp.StatusCode, string(resp.Body))
	})
}

// ============================================================================
// Upgrade Endpoint Tests
// ============================================================================

// TestQueryCurrentPlan tests GET /api/kira/upgrade/current_plan
func TestQueryCurrentPlan(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	t.Run("query current upgrade plan", func(t *testing.T) {
		resp, err := client.Get("/api/kira/upgrade/current_plan", nil)
		require.NoError(t, err)
		assert.True(t, resp.IsSuccess(), "Expected success, got status %d: %s", resp.StatusCode, string(resp.Body))
	})
}

// TestQueryNextPlan tests GET /api/kira/upgrade/next_plan
func TestQueryNextPlan(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	t.Run("query next upgrade plan", func(t *testing.T) {
		resp, err := client.Get("/api/kira/upgrade/next_plan", nil)
		require.NoError(t, err)
		assert.True(t, resp.IsSuccess(), "Expected success, got status %d: %s", resp.StatusCode, string(resp.Body))
	})
}

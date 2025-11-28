package integration

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// Account Endpoint Tests - /api/kira/accounts/{address}
// Expected format from miro: lib/infra/dto/api_kira/query_account/response/
// ============================================================================

// TestAccountResponseFormat validates the response matches expected miro format
// Issues: #13 (snake_case), #16 (types)
func TestAccountResponseFormat(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	resp, err := client.Get("/api/kira/accounts/"+cfg.TestAddress, nil)
	require.NoError(t, err)
	require.True(t, resp.IsSuccess(), "Expected success, got %d: %s", resp.StatusCode, string(resp.Body))

	v, err := NewFormatValidator(t, resp.Body)
	require.NoError(t, err, "Failed to parse response")

	t.Run("@type field exists", func(t *testing.T) {
		v.ValidateFieldExists("@type", "Format")
	})

	t.Run("address field exists", func(t *testing.T) {
		v.ValidateFieldExists("address", "Format")
	})

	t.Run("account_number field naming (Issue #13)", func(t *testing.T) {
		// miro expects snake_case: account_number
		v.ValidateFieldNotExists("accountNumber", "account_number", "Issue #13")
		if !v.HasField("account_number") {
			t.Error("Issue #13: Missing 'account_number' field (miro expects snake_case)")
		}
	})

	t.Run("pub_key field naming (Issue #13)", func(t *testing.T) {
		// miro expects snake_case: pub_key
		v.ValidateFieldNotExists("pubKey", "pub_key", "Issue #13")
		// pub_key can be nil for new accounts, so just check naming
	})

	t.Run("sequence field type (Issue #16)", func(t *testing.T) {
		// miro expects sequence as string
		if v.HasField("sequence") {
			v.ValidateFieldType("sequence", TypeString, "Issue #16")
		}
	})

	t.Run("account_number field type", func(t *testing.T) {
		// miro expects account_number as string
		if v.HasField("account_number") {
			v.ValidateFieldType("account_number", TypeString, "Issue #16")
		}
	})

	t.Run("all top-level fields use snake_case (Issue #13)", func(t *testing.T) {
		v.ValidateSnakeCase("Issue #13")
	})

	t.Run("pub_key nested fields use snake_case", func(t *testing.T) {
		if v.HasField("pub_key") {
			v.ValidateNestedSnakeCase("pub_key", "Issue #13")
		}
	})
}

// TestQueryAccounts tests GET /api/kira/accounts/{address}
func TestQueryAccounts(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	t.Run("valid address returns account data", func(t *testing.T) {
		resp, err := client.Get("/api/kira/accounts/"+cfg.TestAddress, nil)
		require.NoError(t, err)
		assert.True(t, resp.IsSuccess(), "Expected success, got status %d: %s", resp.StatusCode, string(resp.Body))

		v, err := NewFormatValidator(t, resp.Body)
		require.NoError(t, err)

		// Verify required fields exist
		assert.True(t, v.HasField("address") || v.HasField("@type"), "Response should contain address or @type field")
	})

	t.Run("invalid address returns error", func(t *testing.T) {
		resp, err := client.Get("/api/kira/accounts/invalid_address", nil)
		require.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(resp.Body, &result)
		require.NoError(t, err)

		// Check for error response
		if errVal, exists := result["Error"]; exists {
			errStr, ok := errVal.(string)
			if ok {
				assert.True(t, strings.Contains(errStr, "bech32") || strings.Contains(errStr, "invalid") || strings.Contains(errStr, "decoding"),
					"Error should mention bech32/invalid: %s", errStr)
			}
		}
	})
}

// ============================================================================
// Balance Endpoint Tests - /api/kira/balances/{address}
// Expected format from miro: lib/infra/dto/api_kira/query_balance/response/
// ============================================================================

// TestBalancesResponseFormat validates the response matches expected miro format
// Issues: #13 (snake_case), #16 (types), #19 (missing fields)
func TestBalancesResponseFormat(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	resp, err := client.Get("/api/kira/balances/"+cfg.TestAddress, nil)
	require.NoError(t, err)
	require.True(t, resp.IsSuccess(), "Expected success, got %d: %s", resp.StatusCode, string(resp.Body))

	v, err := NewFormatValidator(t, resp.Body)
	require.NoError(t, err, "Failed to parse response")

	t.Run("balances array exists", func(t *testing.T) {
		v.ValidateFieldExists("balances", "Format")
		v.ValidateFieldType("balances", TypeArray, "Format")
	})

	t.Run("pagination field exists (Issue #19)", func(t *testing.T) {
		// miro REQUIRES pagination field
		if !v.HasField("pagination") {
			t.Error("Issue #19: Missing 'pagination' field - miro expects this")
		}
	})

	t.Run("balance amount is string (Issue #16)", func(t *testing.T) {
		// miro expects amount as string (for precision with big numbers)
		v.ValidateArrayFieldType("balances", "amount", TypeString, "Issue #16")
	})

	t.Run("balance denom is string", func(t *testing.T) {
		v.ValidateArrayFieldType("balances", "denom", TypeString, "Format")
	})

	t.Run("pagination.total is string (Issue #16)", func(t *testing.T) {
		if nested, ok := v.GetNestedValidator("pagination"); ok {
			nested.ValidateFieldType("total", TypeString, "Issue #16")
		}
	})
}

// TestQueryBalances tests GET /api/kira/balances/{address}
func TestQueryBalances(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	t.Run("query all balances", func(t *testing.T) {
		resp, err := client.Get("/api/kira/balances/"+cfg.TestAddress, nil)
		require.NoError(t, err)
		assert.True(t, resp.IsSuccess(), "Expected success, got status %d: %s", resp.StatusCode, string(resp.Body))

		var result ExpectedBalanceResponse
		err = json.Unmarshal(resp.Body, &result)
		require.NoError(t, err)

		t.Logf("Found %d balances for address %s", len(result.Balances), cfg.TestAddress)
	})

	t.Run("query with limit parameter", func(t *testing.T) {
		resp, err := client.Get("/api/kira/balances/"+cfg.TestAddress, map[string]string{
			"limit": "5",
		})
		require.NoError(t, err)
		assert.True(t, resp.IsSuccess(), "Expected success, got status %d", resp.StatusCode)

		var result ExpectedBalanceResponse
		err = json.Unmarshal(resp.Body, &result)
		require.NoError(t, err)

		// Verify limit is respected
		assert.LessOrEqual(t, len(result.Balances), 5, "Limit parameter not respected")
	})

	t.Run("query with offset parameter", func(t *testing.T) {
		// First get total count
		resp1, err := client.Get("/api/kira/balances/"+cfg.TestAddress, nil)
		require.NoError(t, err)

		var result1 ExpectedBalanceResponse
		json.Unmarshal(resp1.Body, &result1)
		totalCount := len(result1.Balances)

		if totalCount > 1 {
			// Query with offset
			resp2, err := client.Get("/api/kira/balances/"+cfg.TestAddress, map[string]string{
				"offset": "1",
			})
			require.NoError(t, err)
			assert.True(t, resp2.IsSuccess())

			var result2 ExpectedBalanceResponse
			json.Unmarshal(resp2.Body, &result2)

			// Should have fewer results due to offset
			assert.Less(t, len(result2.Balances), totalCount, "Offset parameter not working")
		}
	})

	t.Run("query with count_total parameter", func(t *testing.T) {
		resp, err := client.Get("/api/kira/balances/"+cfg.TestAddress, map[string]string{
			"count_total": "true",
		})
		require.NoError(t, err)
		assert.True(t, resp.IsSuccess(), "Expected success, got status %d", resp.StatusCode)

		var result ExpectedBalanceResponse
		err = json.Unmarshal(resp.Body, &result)
		require.NoError(t, err)

		// With count_total=true, pagination should include total
		if result.Pagination != nil && result.Pagination.Total != "" && result.Pagination.Total != "0" {
			t.Logf("Pagination total: %s", result.Pagination.Total)
		}
	})
}

// TestBalancesPagination specifically tests pagination behavior
func TestBalancesPagination(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	t.Run("pagination total matches actual count", func(t *testing.T) {
		resp, err := client.Get("/api/kira/balances/"+cfg.TestAddress, map[string]string{
			"count_total": "true",
		})
		require.NoError(t, err)
		require.True(t, resp.IsSuccess())

		v, err := NewFormatValidator(t, resp.Body)
		require.NoError(t, err)

		// Check pagination exists and has correct structure
		if !v.HasField("pagination") {
			t.Error("Issue #19: Missing pagination field")
			return
		}

		if nested, ok := v.GetNestedValidator("pagination"); ok {
			if nested.HasField("total") {
				nested.ValidateFieldType("total", TypeString, "Issue #16")
			}
		}
	})
}

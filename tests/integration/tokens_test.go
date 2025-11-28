package integration

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// Token Alias Endpoint Tests - /api/kira/tokens/aliases
// Expected format from miro: lib/infra/dto/api_kira/query_kira_tokens_aliases/
// ============================================================================

// TestTokenAliasesResponseFormat validates the response matches expected miro format
func TestTokenAliasesResponseFormat(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	resp, err := client.Get("/api/kira/tokens/aliases", nil)
	require.NoError(t, err)
	require.True(t, resp.IsSuccess(), "Expected success, got %d: %s", resp.StatusCode, string(resp.Body))

	v, err := NewFormatValidator(t, resp.Body)
	require.NoError(t, err, "Failed to parse response")

	t.Run("token_aliases_data array exists", func(t *testing.T) {
		v.ValidateFieldExists("token_aliases_data", "Format")
		v.ValidateFieldType("token_aliases_data", TypeArray, "Format")
	})

	t.Run("default_denom field exists", func(t *testing.T) {
		v.ValidateFieldExists("default_denom", "Format")
		v.ValidateFieldType("default_denom", TypeString, "Format")
	})

	t.Run("bech32_prefix field exists", func(t *testing.T) {
		v.ValidateFieldExists("bech32_prefix", "Format")
		v.ValidateFieldType("bech32_prefix", TypeString, "Format")
	})

	// Check token alias structure if we have data
	if v.GetArrayLength("token_aliases_data") > 0 {
		t.Run("token alias decimals is int (Issue #16)", func(t *testing.T) {
			raw := v.GetRaw()
			aliases := raw["token_aliases_data"].([]interface{})
			alias := aliases[0].(map[string]interface{})

			if decimals, exists := alias["decimals"]; exists {
				switch decimals.(type) {
				case float64:
					// OK - JSON numbers are float64
				case string:
					t.Error("Issue #16: 'decimals' should be int, got string")
				}
			}
		})

		t.Run("token alias amount is string (Issue #16)", func(t *testing.T) {
			raw := v.GetRaw()
			aliases := raw["token_aliases_data"].([]interface{})
			alias := aliases[0].(map[string]interface{})

			if amount, exists := alias["amount"]; exists {
				switch amount.(type) {
				case string:
					// OK
				case float64:
					t.Error("Issue #16: 'amount' should be string (for precision), got number")
				}
			}
		})

		t.Run("token alias denoms is array", func(t *testing.T) {
			raw := v.GetRaw()
			aliases := raw["token_aliases_data"].([]interface{})
			alias := aliases[0].(map[string]interface{})

			if denoms, exists := alias["denoms"]; exists {
				if _, ok := denoms.([]interface{}); !ok {
					t.Error("'denoms' should be an array")
				}
			}
		})
	}
}

// TestQueryTokenAliases tests GET /api/kira/tokens/aliases
func TestQueryTokenAliases(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	t.Run("query token aliases", func(t *testing.T) {
		resp, err := client.Get("/api/kira/tokens/aliases", nil)
		require.NoError(t, err)
		assert.True(t, resp.IsSuccess(), "Expected success, got status %d: %s", resp.StatusCode, string(resp.Body))

		var result map[string]interface{}
		json.Unmarshal(resp.Body, &result)

		if aliases, ok := result["token_aliases_data"].([]interface{}); ok {
			t.Logf("Found %d token aliases", len(aliases))
		}
	})

	t.Run("query token aliases with pagination", func(t *testing.T) {
		resp, err := client.Get("/api/kira/tokens/aliases", map[string]string{
			"limit":  "1",
			"offset": "0",
		})
		require.NoError(t, err)
		assert.True(t, resp.IsSuccess(), "Expected success, got status %d", resp.StatusCode)
	})

	t.Run("query token aliases with count total", func(t *testing.T) {
		resp, err := client.Get("/api/kira/tokens/aliases", map[string]string{
			"count_total": "true",
		})
		require.NoError(t, err)
		assert.True(t, resp.IsSuccess(), "Expected success, got status %d", resp.StatusCode)
	})

	t.Run("query specific token alias", func(t *testing.T) {
		resp, err := client.Get("/api/kira/tokens/aliases", map[string]string{
			"tokens": "ukex",
		})
		require.NoError(t, err)
		assert.True(t, resp.IsSuccess(), "Expected success, got status %d", resp.StatusCode)
	})
}

// ============================================================================
// Token Rates Endpoint Tests - /api/kira/tokens/rates
// Expected format from miro: lib/infra/dto/api_kira/query_kira_tokens_rates/
// ============================================================================

// TestTokenRatesResponseFormat validates the response matches expected miro format
func TestTokenRatesResponseFormat(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	resp, err := client.Get("/api/kira/tokens/rates", nil)
	require.NoError(t, err)
	require.True(t, resp.IsSuccess(), "Expected success, got %d: %s", resp.StatusCode, string(resp.Body))

	v, err := NewFormatValidator(t, resp.Body)
	require.NoError(t, err, "Failed to parse response")

	t.Run("data array exists", func(t *testing.T) {
		v.ValidateFieldExists("data", "Format")
		v.ValidateFieldType("data", TypeArray, "Format")
	})

	// Check token rate structure if we have data
	if v.GetArrayLength("data") > 0 {
		t.Run("token rate fields use snake_case (Issue #13)", func(t *testing.T) {
			raw := v.GetRaw()
			rates := raw["data"].([]interface{})
			rate := rates[0].(map[string]interface{})

			expectedSnakeCaseFields := []string{
				"fee_payments",
				"fee_rate",
				"stake_cap",
				"stake_min",
				"stake_token",
			}

			for _, field := range expectedSnakeCaseFields {
				if _, exists := rate[field]; !exists {
					camelCase := snakeToCamel(field)
					if _, camelExists := rate[camelCase]; camelExists {
						t.Errorf("Issue #13: Using '%s' instead of '%s'", camelCase, field)
					}
				}
			}
		})

		t.Run("token rate boolean fields are bools (Issue #16)", func(t *testing.T) {
			raw := v.GetRaw()
			rates := raw["data"].([]interface{})
			rate := rates[0].(map[string]interface{})

			boolFields := []string{"fee_payments", "stake_token"}
			for _, field := range boolFields {
				if val, exists := rate[field]; exists {
					switch val.(type) {
					case bool:
						// OK
					case string:
						t.Errorf("Issue #16: '%s' should be bool, got string", field)
					case float64:
						t.Errorf("Issue #16: '%s' should be bool, got number", field)
					}
				}
			}
		})

		t.Run("token rate string fields are strings", func(t *testing.T) {
			raw := v.GetRaw()
			rates := raw["data"].([]interface{})
			rate := rates[0].(map[string]interface{})

			stringFields := []string{"denom", "fee_rate", "stake_cap", "stake_min"}
			for _, field := range stringFields {
				if val, exists := rate[field]; exists {
					switch val.(type) {
					case string:
						// OK
					case float64:
						t.Errorf("Issue #16: '%s' should be string, got number", field)
					}
				}
			}
		})
	}
}

// TestQueryTokenRates tests GET /api/kira/tokens/rates
func TestQueryTokenRates(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	t.Run("query token rates", func(t *testing.T) {
		resp, err := client.Get("/api/kira/tokens/rates", nil)
		require.NoError(t, err)
		assert.True(t, resp.IsSuccess(), "Expected success, got status %d: %s", resp.StatusCode, string(resp.Body))

		var result map[string]interface{}
		json.Unmarshal(resp.Body, &result)

		if rates, ok := result["data"].([]interface{}); ok {
			t.Logf("Found %d token rates", len(rates))
		}
	})
}

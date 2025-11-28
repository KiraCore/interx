package integration

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// Transaction Endpoint Tests - /api/transactions
// Expected format from miro: lib/infra/dto/api/query_transactions/response/
// ============================================================================

// ActualTransactionsResponse captures what the API actually returns
type ActualTransactionsResponse struct {
	Transactions []ActualTransaction `json:"transactions"`
	Total        interface{}         `json:"total,omitempty"`
	TotalCount   interface{}         `json:"total_count,omitempty"`
}

// ActualTransaction captures the actual response structure
type ActualTransaction struct {
	Hash      string          `json:"hash"`
	Height    json.RawMessage `json:"height"`
	Timestamp json.RawMessage `json:"timestamp"`
	Time      json.RawMessage `json:"time"`
	Status    string          `json:"status"`
	Direction string          `json:"direction"`
	Messages  []interface{}   `json:"messages"`
	Txs       []interface{}   `json:"txs"`
	TxResult  interface{}     `json:"tx_result"`
}

// TestTransactionsResponseFormat validates the response matches expected miro format
// Issues: #13 (snake_case), #16 (types), #19 (missing fields), #40 (type filter), #41 (address/direction filter)
func TestTransactionsResponseFormat(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	resp, err := client.Get("/api/transactions", map[string]string{
		"limit": "10",
	})
	require.NoError(t, err)
	require.True(t, resp.IsSuccess(), "Expected success, got %d: %s", resp.StatusCode, string(resp.Body))

	v, err := NewFormatValidator(t, resp.Body)
	require.NoError(t, err, "Failed to parse response")

	t.Run("transactions array exists", func(t *testing.T) {
		v.ValidateFieldExists("transactions", "Format")
		v.ValidateFieldType("transactions", TypeArray, "Format")
	})

	t.Run("total_count field exists and is int (Issue #19, #16)", func(t *testing.T) {
		// miro expects total_count as snake_case INT
		v.ValidateFieldNotExists("totalCount", "total_count", "Issue #13")
		if !v.ValidateFieldExists("total_count", "Issue #19") {
			t.Error("Issue #19: Missing 'total_count' field")
		} else {
			v.ValidateFieldType("total_count", TypeInt, "Issue #16")
		}
	})

	// Check transaction structure if we have transactions
	if v.GetArrayLength("transactions") > 0 {
		t.Run("transaction.time is Unix int (Issue #16, #19)", func(t *testing.T) {
			// miro expects 'time' as Unix timestamp INT, not string, not ISO
			raw := v.GetRaw()
			txs := raw["transactions"].([]interface{})
			tx := txs[0].(map[string]interface{})

			// Check field exists
			if _, exists := tx["time"]; !exists {
				t.Error("Issue #19: Missing 'time' field in transaction")
				return
			}

			// Check type
			switch tx["time"].(type) {
			case float64:
				// OK - JSON numbers are float64
			case string:
				t.Error("Issue #16: 'time' should be Unix int, got string (possibly ISO format)")
			default:
				t.Errorf("Issue #16: 'time' has unexpected type %T", tx["time"])
			}
		})

		t.Run("transaction.hash exists", func(t *testing.T) {
			v.ValidateArrayFieldType("transactions", "hash", TypeString, "Format")
		})

		t.Run("transaction.status field exists (Issue #19)", func(t *testing.T) {
			// miro expects 'status' field: "confirmed", "pending", "failed"
			raw := v.GetRaw()
			txs := raw["transactions"].([]interface{})
			tx := txs[0].(map[string]interface{})

			if _, exists := tx["status"]; !exists {
				t.Error("Issue #19: Missing 'status' field in transaction")
			}
		})

		t.Run("transaction.direction field exists (Issue #19)", func(t *testing.T) {
			// miro expects 'direction' field: "inbound" or "outbound"
			raw := v.GetRaw()
			txs := raw["transactions"].([]interface{})
			tx := txs[0].(map[string]interface{})

			if _, exists := tx["direction"]; !exists {
				t.Error("Issue #19: Missing 'direction' field in transaction")
			}
		})

		t.Run("transaction.txs field not messages (Issue #19)", func(t *testing.T) {
			// miro expects 'txs' array, NOT 'messages'
			raw := v.GetRaw()
			txs := raw["transactions"].([]interface{})
			tx := txs[0].(map[string]interface{})

			if _, exists := tx["messages"]; exists {
				if _, txsExists := tx["txs"]; !txsExists {
					t.Error("Issue #19: Using 'messages' instead of 'txs'")
				}
			}

			if _, exists := tx["txs"]; !exists {
				t.Error("Issue #19: Missing 'txs' field in transaction")
			}
		})

		t.Run("transaction.fee is array of coins", func(t *testing.T) {
			raw := v.GetRaw()
			txs := raw["transactions"].([]interface{})
			tx := txs[0].(map[string]interface{})

			if fee, exists := tx["fee"]; exists {
				_, ok := fee.([]interface{})
				if !ok {
					t.Error("'fee' should be an array of coins")
				}
			}
		})

		t.Run("transaction fields use snake_case (Issue #13)", func(t *testing.T) {
			raw := v.GetRaw()
			txs := raw["transactions"].([]interface{})
			tx := txs[0].(map[string]interface{})

			for key := range tx {
				if isCamelCase(key) {
					t.Errorf("Issue #13: Transaction field '%s' uses camelCase, expected snake_case", key)
				}
			}
		})
	}
}

// TestTransactionsAddressFilter tests the address query parameter (Issue #41)
func TestTransactionsAddressFilter(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	t.Run("address filter returns only relevant transactions (Issue #41)", func(t *testing.T) {
		// Get all transactions first
		allResp, err := client.Get("/api/transactions", map[string]string{"limit": "50"})
		require.NoError(t, err)
		var allResult map[string]interface{}
		json.Unmarshal(allResp.Body, &allResult)
		allTxs, _ := allResult["transactions"].([]interface{})

		// Query with address filter
		resp, err := client.Get("/api/transactions", map[string]string{
			"address": cfg.TestAddress,
			"limit":   "50",
		})
		require.NoError(t, err)
		require.True(t, resp.IsSuccess())

		var result map[string]interface{}
		err = json.Unmarshal(resp.Body, &result)
		require.NoError(t, err)

		txs, ok := result["transactions"].([]interface{})
		if !ok || len(txs) == 0 {
			t.Skip("No transactions found for test address")
			return
		}

		// BUG CHECK: If filtered count equals unfiltered count, filter may not be working
		if len(txs) == len(allTxs) && len(allTxs) > 10 {
			t.Errorf("Issue #41: Address filter may not be working - filtered count (%d) equals unfiltered count (%d)", len(txs), len(allTxs))
		}

		// Verify all returned transactions involve the filtered address
		for i, tx := range txs {
			txMap := tx.(map[string]interface{})
			txJSON, _ := json.Marshal(txMap)
			txStr := string(txJSON)

			// The address should appear somewhere in the transaction
			if !containsSubstr(txStr, cfg.TestAddress) {
				t.Errorf("Issue #41: Transaction %d does not involve filtered address %s", i, cfg.TestAddress)
				break // Only report first occurrence
			}
		}
	})

	t.Run("non-existent address returns no transactions (Issue #41)", func(t *testing.T) {
		nonExistentAddr := "kira1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq"
		resp, err := client.Get("/api/transactions", map[string]string{
			"address": nonExistentAddr,
			"limit":   "50",
		})
		require.NoError(t, err)
		require.True(t, resp.IsSuccess())

		var result map[string]interface{}
		json.Unmarshal(resp.Body, &result)
		txs, _ := result["transactions"].([]interface{})

		if len(txs) > 0 {
			t.Errorf("Issue #41: Non-existent address returned %d transactions (expected 0)", len(txs))
		}
	})
}

// TestTransactionsTypeFilter tests the type query parameter (Issue #40)
func TestTransactionsTypeFilter(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	t.Run("type filter returns only matching transaction types (Issue #40)", func(t *testing.T) {
		// Query with type filter for MsgSend
		resp, err := client.Get("/api/transactions", map[string]string{
			"type":  "MsgSend",
			"limit": "50",
		})
		require.NoError(t, err)
		require.True(t, resp.IsSuccess())

		var result map[string]interface{}
		err = json.Unmarshal(resp.Body, &result)
		require.NoError(t, err)

		txs, ok := result["transactions"].([]interface{})
		if !ok || len(txs) == 0 {
			t.Skip("No MsgSend transactions found")
			return
		}

		// Verify all returned transactions are MsgSend type
		for i, tx := range txs {
			txMap := tx.(map[string]interface{})
			txJSON, _ := json.Marshal(txMap)
			txStr := string(txJSON)

			// Check for MsgSend in the transaction
			if !containsSubstr(txStr, "MsgSend") && !containsSubstr(txStr, "msg_send") && !containsSubstr(txStr, "Send") {
				t.Errorf("Issue #40: Transaction %d is not MsgSend type", i)
				t.Logf("Transaction: %s", txStr[:minInt(500, len(txStr))])
				break
			}
		}
	})

	t.Run("non-existent type returns no transactions (Issue #40)", func(t *testing.T) {
		resp, err := client.Get("/api/transactions", map[string]string{
			"type":  "/nonexistent.type.MsgNonExistent",
			"limit": "50",
		})
		require.NoError(t, err)
		require.True(t, resp.IsSuccess())

		var result map[string]interface{}
		json.Unmarshal(resp.Body, &result)
		txs, _ := result["transactions"].([]interface{})

		if len(txs) > 0 {
			t.Errorf("Issue #40: Non-existent type returned %d transactions (expected 0)", len(txs))
		}
	})
}

// TestTransactionsDirectionFilter tests the direction query parameter (Issue #41)
func TestTransactionsDirectionFilter(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	t.Run("direction=inbound filter (Issue #41)", func(t *testing.T) {
		resp, err := client.Get("/api/transactions", map[string]string{
			"address":   cfg.TestAddress,
			"direction": "inbound",
			"limit":     "20",
		})
		require.NoError(t, err)
		require.True(t, resp.IsSuccess())

		var result map[string]interface{}
		json.Unmarshal(resp.Body, &result)

		txs, ok := result["transactions"].([]interface{})
		if !ok || len(txs) == 0 {
			t.Skip("No inbound transactions found")
			return
		}

		// Check if direction field exists and has correct value
		for i, tx := range txs {
			txMap := tx.(map[string]interface{})
			if dir, exists := txMap["direction"]; exists {
				if dir != "inbound" {
					t.Errorf("Issue #41: Transaction %d has direction '%v', expected 'inbound'", i, dir)
					break
				}
			}
		}
	})

	t.Run("direction=outbound filter (Issue #41)", func(t *testing.T) {
		resp, err := client.Get("/api/transactions", map[string]string{
			"address":   cfg.TestAddress,
			"direction": "outbound",
			"limit":     "20",
		})
		require.NoError(t, err)
		require.True(t, resp.IsSuccess())

		var result map[string]interface{}
		json.Unmarshal(resp.Body, &result)

		txs, ok := result["transactions"].([]interface{})
		if !ok || len(txs) == 0 {
			t.Skip("No outbound transactions found")
			return
		}

		// Check if direction field exists and has correct value
		for i, tx := range txs {
			txMap := tx.(map[string]interface{})
			if dir, exists := txMap["direction"]; exists {
				if dir != "outbound" {
					t.Errorf("Issue #41: Transaction %d has direction '%v', expected 'outbound'", i, dir)
					break
				}
			}
		}
	})

	t.Run("inbound + outbound should not equal all (Issue #41)", func(t *testing.T) {
		// Get all transactions for address
		allResp, err := client.Get("/api/transactions", map[string]string{
			"address": cfg.TestAddress,
			"limit":   "100",
		})
		require.NoError(t, err)
		var allResult map[string]interface{}
		json.Unmarshal(allResp.Body, &allResult)
		allTxs, _ := allResult["transactions"].([]interface{})

		// Get inbound
		inResp, _ := client.Get("/api/transactions", map[string]string{
			"address": cfg.TestAddress, "direction": "inbound", "limit": "100",
		})
		var inResult map[string]interface{}
		json.Unmarshal(inResp.Body, &inResult)
		inTxs, _ := inResult["transactions"].([]interface{})

		// Get outbound
		outResp, _ := client.Get("/api/transactions", map[string]string{
			"address": cfg.TestAddress, "direction": "outbound", "limit": "100",
		})
		var outResult map[string]interface{}
		json.Unmarshal(outResp.Body, &outResult)
		outTxs, _ := outResult["transactions"].([]interface{})

		// If direction filter works, inbound + outbound should roughly equal all
		// If both return same count as all, filter is broken
		if len(allTxs) > 5 && len(inTxs) == len(allTxs) && len(outTxs) == len(allTxs) {
			t.Errorf("Issue #41: Direction filter broken - all=%d, inbound=%d, outbound=%d (both should not equal all)",
				len(allTxs), len(inTxs), len(outTxs))
		}
	})
}

// TestTransactionsSorting tests the sort query parameter
func TestTransactionsSorting(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	t.Run("sort=dateASC returns oldest first", func(t *testing.T) {
		resp, err := client.Get("/api/transactions", map[string]string{
			"sort":  "dateASC",
			"limit": "10",
		})
		require.NoError(t, err)
		require.True(t, resp.IsSuccess())

		var result map[string]interface{}
		json.Unmarshal(resp.Body, &result)

		txs, ok := result["transactions"].([]interface{})
		if !ok || len(txs) < 2 {
			t.Skip("Not enough transactions to test sorting")
			return
		}

		// Get timestamps or heights to verify order
		var times []float64
		for _, tx := range txs {
			txMap := tx.(map[string]interface{})
			if time, exists := txMap["time"]; exists {
				if timeFloat, ok := time.(float64); ok {
					times = append(times, timeFloat)
				}
			} else if height, exists := txMap["height"]; exists {
				if heightFloat, ok := height.(float64); ok {
					times = append(times, heightFloat)
				}
			}
		}

		// Verify ascending order
		for i := 1; i < len(times); i++ {
			if times[i] < times[i-1] {
				t.Errorf("Sort dateASC not working: position %d (%v) < position %d (%v)", i, times[i], i-1, times[i-1])
				break
			}
		}
	})

	t.Run("sort=dateDESC returns newest first", func(t *testing.T) {
		resp, err := client.Get("/api/transactions", map[string]string{
			"sort":  "dateDESC",
			"limit": "10",
		})
		require.NoError(t, err)
		require.True(t, resp.IsSuccess())

		var result map[string]interface{}
		json.Unmarshal(resp.Body, &result)

		txs, ok := result["transactions"].([]interface{})
		if !ok || len(txs) < 2 {
			t.Skip("Not enough transactions to test sorting")
			return
		}

		// Get timestamps or heights to verify order
		var times []float64
		for _, tx := range txs {
			txMap := tx.(map[string]interface{})
			if time, exists := txMap["time"]; exists {
				if timeFloat, ok := time.(float64); ok {
					times = append(times, timeFloat)
				}
			} else if height, exists := txMap["height"]; exists {
				if heightFloat, ok := height.(float64); ok {
					times = append(times, heightFloat)
				}
			}
		}

		// Verify descending order
		for i := 1; i < len(times); i++ {
			if times[i] > times[i-1] {
				t.Errorf("NEW BUG - Sort dateDESC not working: position %d (%v) > position %d (%v)", i, times[i], i-1, times[i-1])
				break
			}
		}
	})
}

// TestTransactionsPagination tests limit and offset parameters
func TestTransactionsPagination(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	t.Run("limit parameter works", func(t *testing.T) {
		resp, err := client.Get("/api/transactions", map[string]string{
			"limit": "5",
		})
		require.NoError(t, err)
		require.True(t, resp.IsSuccess())

		var result map[string]interface{}
		json.Unmarshal(resp.Body, &result)

		txs, ok := result["transactions"].([]interface{})
		if !ok {
			t.Error("Missing transactions array")
			return
		}

		assert.LessOrEqual(t, len(txs), 5, "Limit parameter not respected")
	})

	t.Run("offset parameter works", func(t *testing.T) {
		// Get first page
		resp1, err := client.Get("/api/transactions", map[string]string{
			"limit": "5",
		})
		require.NoError(t, err)

		var result1 map[string]interface{}
		json.Unmarshal(resp1.Body, &result1)
		txs1, _ := result1["transactions"].([]interface{})

		// Get second page
		resp2, err := client.Get("/api/transactions", map[string]string{
			"limit":  "5",
			"offset": "5",
		})
		require.NoError(t, err)

		var result2 map[string]interface{}
		json.Unmarshal(resp2.Body, &result2)
		txs2, _ := result2["transactions"].([]interface{})

		if len(txs1) > 0 && len(txs2) > 0 {
			// First transaction of page 2 should be different from page 1
			tx1 := txs1[0].(map[string]interface{})
			tx2 := txs2[0].(map[string]interface{})

			if tx1["hash"] == tx2["hash"] {
				t.Error("Offset parameter not working - same transactions returned")
			}
		}
	})
}

// TestTransactionsStatusFilter tests the status query parameter
func TestTransactionsStatusFilter(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	t.Run("status=confirmed filter", func(t *testing.T) {
		resp, err := client.Get("/api/transactions", map[string]string{
			"status": "confirmed",
			"limit":  "10",
		})
		require.NoError(t, err)
		require.True(t, resp.IsSuccess())

		var result map[string]interface{}
		json.Unmarshal(resp.Body, &result)

		txs, ok := result["transactions"].([]interface{})
		if !ok || len(txs) == 0 {
			t.Skip("No confirmed transactions found")
			return
		}

		// Check if all transactions have status=confirmed
		for i, tx := range txs {
			txMap := tx.(map[string]interface{})
			if status, exists := txMap["status"]; exists {
				if status != "confirmed" {
					t.Errorf("Status filter not working: transaction %d has status '%v'", i, status)
					break
				}
			}
		}
	})
}

// ============================================================================
// Block Endpoint Tests - /api/blocks
// ============================================================================

// TestBlocksEndpoint tests GET /api/blocks
func TestBlocksEndpoint(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	t.Run("returns block list", func(t *testing.T) {
		resp, err := client.Get("/api/blocks", map[string]string{
			"limit": "10",
		})
		require.NoError(t, err)
		require.True(t, resp.IsSuccess(), "Expected success, got %d: %s", resp.StatusCode, string(resp.Body))

		var result map[string]interface{}
		err = json.Unmarshal(resp.Body, &result)
		require.NoError(t, err)

		// Should have blocks array or block data
		t.Logf("Blocks response keys: %v", getMapKeys(result))
	})

	t.Run("pagination works", func(t *testing.T) {
		resp1, err := client.Get("/api/blocks", map[string]string{"limit": "5"})
		require.NoError(t, err)

		resp2, err := client.Get("/api/blocks", map[string]string{"limit": "5", "offset": "5"})
		require.NoError(t, err)

		assert.True(t, resp1.IsSuccess())
		assert.True(t, resp2.IsSuccess())
	})
}

// TestBlockByHeight tests GET /api/blocks/{height}
func TestBlockByHeight(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	t.Run("valid height returns block data", func(t *testing.T) {
		resp, err := client.Get("/api/blocks/1", nil)
		require.NoError(t, err)
		require.True(t, resp.IsSuccess(), "Expected success for block 1, got %d: %s", resp.StatusCode, string(resp.Body))

		var result map[string]interface{}
		err = json.Unmarshal(resp.Body, &result)
		require.NoError(t, err)

		t.Logf("Block 1 response keys: %v", getMapKeys(result))
	})

	t.Run("invalid height returns error", func(t *testing.T) {
		resp, err := client.Get("/api/blocks/999999999999", nil)
		require.NoError(t, err)
		// Should either return error status or error in body
		if resp.IsSuccess() {
			var result map[string]interface{}
			json.Unmarshal(resp.Body, &result)
			if _, hasError := result["Error"]; !hasError {
				t.Log("Note: Invalid block height did not return error")
			}
		}
	})
}

// TestTransactionHash tests GET /api/kira/txs/{hash}
func TestTransactionHash(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	t.Run("query transaction by hash", func(t *testing.T) {
		txHash := "0xBF97A902D95517A4538531B21E5C2FE6FE3C5F04E392D7A1F43A802E9121232C"
		resp, err := client.Get("/api/kira/txs/"+txHash, nil)
		require.NoError(t, err)
		assert.True(t, resp.StatusCode == 200 || resp.StatusCode == 404,
			"Expected 200 or 404, got %d", resp.StatusCode)
	})
}

// TestQueryUnconfirmedTransactions tests GET /api/unconfirmed_txs
func TestQueryUnconfirmedTransactions(t *testing.T) {
	cfg := GetConfig()
	client := NewClient(cfg)

	t.Run("query unconfirmed transactions", func(t *testing.T) {
		resp, err := client.Get("/api/unconfirmed_txs", nil)
		require.NoError(t, err)
		assert.True(t, resp.IsSuccess(), "Expected success, got status %d: %s", resp.StatusCode, string(resp.Body))
	})
}

// ============================================================================
// Helper functions
// ============================================================================

func containsSubstr(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

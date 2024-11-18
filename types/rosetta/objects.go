package rosetta

type OperationStatus struct {
	Status     string `json:"status"`
	Successful bool   `json:"successful"`
}

type BalanceExemption struct {
	SubAccountAddress string        `json:"sub_account_address,omitempty"`
	Currency          Currency      `json:"currency,omitempty"`
	ExemptionType     ExemptionType `json:"exemption_type,omitempty"`
}

type Allow struct {
	OperationStatuses       []OperationStatus  `json:"operation_statuses"`
	OperationTypes          []string           `json:"operation_types"`
	Errors                  []Error            `json:"errors"`
	HistoricalBalanceLookup bool               `json:"historical_balance_lookup"`
	TimestampStartIndex     int64              `json:"timestamp_start_index,omitempty"`
	CallMethods             []string           `json:"call_methods"`
	BalanceExemptions       []BalanceExemption `json:"balance_exemptions"`
	MempoolCoins            bool               `json:"mempool_coins"`
}

type Currency struct {
	Symbol   string           `json:"symbol"`
	Decimals int64            `json:"decimals"`
	Metadata CurrencyMetadata `json:"metadata,omitempty"`
}

type Amount struct {
	Value    string         `json:"value"`
	Currency Currency       `json:"currency"`
	Metadata AmountMetadata `json:"metadata,omitempty"`
}

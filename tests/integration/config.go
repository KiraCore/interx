package integration

import (
	"os"
	"time"
)

// Config holds the test configuration
type Config struct {
	BaseURL        string
	TestAddress    string
	Timeout        time.Duration
	ValidatorAddr  string
	DelegatorAddr  string
}

// Predefined environments
var (
	ChaosnetConfig = Config{
		BaseURL:       "http://3.123.154.245:11000",
		TestAddress:   "kira143q8vxpvuykt9pq50e6hng9s38vmy844n8k9wx",
		Timeout:       30 * time.Second,
		ValidatorAddr: "kira1vvcj9avffvyav83gmptdlzrprgvsrjxzh7f9sz",
		DelegatorAddr: "kira177lwmjyjds3cy7trers83r4pjn3dhv8zrqk9dl",
	}

	LocalConfig = Config{
		BaseURL:       "http://localhost:11000",
		TestAddress:   "kira143q8vxpvuykt9pq50e6hng9s38vmy844n8k9wx",
		Timeout:       10 * time.Second,
		ValidatorAddr: "kira1vvcj9avffvyav83gmptdlzrprgvsrjxzh7f9sz",
		DelegatorAddr: "kira177lwmjyjds3cy7trers83r4pjn3dhv8zrqk9dl",
	}
)

// GetConfig returns the configuration based on environment variables
func GetConfig() Config {
	cfg := ChaosnetConfig // default

	if url := os.Getenv("INTERX_URL"); url != "" {
		cfg.BaseURL = url
	}

	if addr := os.Getenv("TEST_ADDRESS"); addr != "" {
		cfg.TestAddress = addr
	}

	if valAddr := os.Getenv("VALIDATOR_ADDRESS"); valAddr != "" {
		cfg.ValidatorAddr = valAddr
	}

	if delAddr := os.Getenv("DELEGATOR_ADDRESS"); delAddr != "" {
		cfg.DelegatorAddr = delAddr
	}

	if env := os.Getenv("TEST_ENV"); env == "local" {
		cfg = LocalConfig
	}

	return cfg
}

package log

const (

	// LogFormatPlain is a format for colored text
	LogFormatPlain = "plain"
	// LogFormatJSON is a format for json output
	LogFormatJSON = "json"

	// DefaultLogLevel defines a default log level as INFO.
	DefaultLogLevel = "info"
)

// NOTE: Most of the structs & relevant comments + the
// default configuration options were used to manually
// generate the config.toml. Please reflect any changes
// made here in the defaultConfigTemplate constant in
// config/toml.go
// NOTE: libs/cli must know to look in the config dir!
// var (
// 	// taken from https://semver.org/
// 	semverRegexp = regexp.MustCompile(`^(?P<major>0|[1-9]\d*)\.(?P<minor>0|[1-9]\d*)\.(?P<patch>0|[1-9]\d*)(?:-(?P<prerelease>(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+(?P<buildmetadata>[0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`)
// )

//-----------------------------------------------------------------------------
// BaseConfig

// BaseConfig defines the base configuration for a CometBFT node
type BaseConfig struct { //nolint: maligned

	// Database directory
	DBPath string `mapstructure:"db_dir"`

	// Output level for logging
	LogLevel string `mapstructure:"log_level"`

	// Output format: 'plain' (colored text) or 'json'
	LogFormat string `mapstructure:"log_format"`

	// Path to the JSON file containing the private key to use as a validator in the consensus protocol
	PrivValidatorKey string `mapstructure:"priv_validator_key_file"`

	// Path to the JSON file containing the last sign state of a validator
	PrivValidatorState string `mapstructure:"priv_validator_state_file"`

	// TCP or UNIX socket address for CometBFT to listen on for
	// connections from an external PrivValidator process
	PrivValidatorListenAddr string `mapstructure:"priv_validator_laddr"`
}

// DefaultBaseConfig returns a default base configuration for a CometBFT node
func DefaultBaseConfig() BaseConfig {
	return BaseConfig{

		LogLevel:  DefaultLogLevel,
		LogFormat: LogFormatPlain,
	}
}

// TestBaseConfig returns a base configuration for testing a CometBFT node
func TestBaseConfig() BaseConfig {
	cfg := DefaultBaseConfig()
	return cfg
}

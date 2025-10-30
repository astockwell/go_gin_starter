package main

import (
	"encoding/hex"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/gorilla/securecookie"
	"github.com/spf13/viper"
)

type AppConfig struct {
	// Web server configuration
	HostPort                     string `mapstructure:"host_port"`
	LogFile                      string `mapstructure:"log_file"`
	LogLevel                     int    `mapstructure:"log_level"`
	RegenerateSecureKeys         bool   `mapstructure:"regenerate_secure_keys"`
	SSLDisabled                  bool   `mapstructure:"ssl_disabled"`
	SSLCertFile                  string `mapstructure:"ssl_cert_file"`
	SSLKeyFile                   string `mapstructure:"ssl_key_file"`
	CacheBustVal                 string `mapstructure:"cache_bust"`
	CacheTemplates               bool   `mapstructure:"cache_templates"`
	SecureCookieSigningKey       []byte
	SecureCookieSigningKeyHex    string `mapstructure:"secure_cookie_signing_key"`
	SecureCookieEncryptionKey    []byte
	SecureCookieEncryptionKeyHex string `mapstructure:"secure_cookie_encryption_key"`
	SecureCookieMaxAge           int    `mapstructure:"secure_cookie_max_age"`

	WorkingDir  string
	DebugConfig bool `mapstructure:"debug_config"`
}

func NewAppConfigFromFile(filename string) (*AppConfig, error) {
	// Get current executable's directory
	ex, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("os.Executable(): %w", err)
	}
	wd := filepath.Dir(ex)

	// Add CLI subcommand to generate a skeleton config file
	if len(os.Args) > 1 && os.Args[1] == "generateconfig" {
		// For generateconfig, use current working directory instead of executable directory
		// This is more intuitive when using 'go run . generateconfig'
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Printf("Error getting current directory: %v\n", err)
			os.Exit(1)
		}
		err = generateConfig(filename, cwd)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Skeleton %s.toml generated successfully.\n", filename)
		os.Exit(0)
	}

	// Config file setup
	viper.SetConfigName(filename)
	viper.AddConfigPath(wd)
	if cwd, err := os.Getwd(); err == nil {
		viper.AddConfigPath(cwd)
	}
	viper.AddConfigPath(".")
	err = viper.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("viper.ReadInConfig(): %w", err)
	}

	// Unmarshall to struct
	ac := AppConfig{}
	err = viper.Unmarshal(&ac)
	if err != nil {
		return nil, fmt.Errorf("viper.Unmarshal(): %w", err)
	}

	// Dump new keys keys flag
	if ac.RegenerateSecureKeys {
		// Print to terminal and exit
		generateNewSecureKeys()
		os.Exit(0)
	}

	// Parse all values for any environmental variable inclusions
	ac.ParseEnvVariables()

	// Validate and parse secure cookie keys
	ac.ParseSecureKeys()

	// Save the determination for working directory (where the config file lives)
	if cfgPath := viper.ConfigFileUsed(); cfgPath != "" {
		ac.WorkingDir = filepath.Dir(cfgPath)
	} else {
		ac.WorkingDir = wd
	}

	// Post-processing
	if !strings.HasPrefix(ac.HostPort, ":") {
		ac.HostPort = fmt.Sprintf(":%s", ac.HostPort)
	}

	// Dump config if flag set
	if ac.DebugConfig {
		spew.Dump(ac)
		os.Exit(0)
	}

	return &ac, nil
}

func (a *AppConfig) ParseEnvVariables() {
	HostPort := os.ExpandEnv(a.HostPort)
	if HostPort != "" {
		a.HostPort = HostPort
	}
	LogFile := os.ExpandEnv(a.LogFile)
	if LogFile != "" {
		a.LogFile = LogFile
	}
	SSLCertFile := os.ExpandEnv(a.SSLCertFile)
	if SSLCertFile != "" {
		a.SSLCertFile = SSLCertFile
	}
	SSLKeyFile := os.ExpandEnv(a.SSLKeyFile)
	if SSLKeyFile != "" {
		a.SSLKeyFile = SSLKeyFile
	}
	SecureCookieSigningKeyHex := os.ExpandEnv(a.SecureCookieSigningKeyHex)
	if SecureCookieSigningKeyHex != "" {
		a.SecureCookieSigningKeyHex = SecureCookieSigningKeyHex
	}
	SecureCookieEncryptionKeyHex := os.ExpandEnv(a.SecureCookieEncryptionKeyHex)
	if SecureCookieEncryptionKeyHex != "" {
		a.SecureCookieEncryptionKeyHex = SecureCookieEncryptionKeyHex
	}
}

func (a *AppConfig) ParseSecureKeys() {
	if a.SecureCookieSigningKeyHex == "" {
		slog.Error("No 'secure_cookie_signing_key' found in config file")
		os.Exit(1)
	}
	secureCookieSigningKey, err := hex.DecodeString(a.SecureCookieSigningKeyHex)
	if err != nil {
		slog.Error("Error decoding 'secure_cookie_signing_key' from hex", "error", err)
		os.Exit(1)
	}
	if len(secureCookieSigningKey) != 64 {
		slog.Error("Error 'secure_cookie_signing_key' must be 64 bytes long")
		os.Exit(1)
	}
	a.SecureCookieSigningKey = secureCookieSigningKey

	if a.SecureCookieEncryptionKeyHex == "" {
		slog.Error("No 'secure_cookie_encryption_key' found in config file")
		os.Exit(1)
	}
	secureCookieEncryptionKey, err := hex.DecodeString(a.SecureCookieEncryptionKeyHex)
	if err != nil {
		slog.Error("Error decoding 'secure_cookie_encryption_key' from hex", "error", err)
		os.Exit(1)
	}
	if len(secureCookieEncryptionKey) != 32 {
		slog.Error("Error 'secure_cookie_encryption_key' must be 64 bytes long")
		os.Exit(1)
	}
	a.SecureCookieEncryptionKey = secureCookieEncryptionKey
}

func generateNewSecureKeys() {
	// Generate and print out new keys
	fmt.Println("\n* * * Start generating new secure keys * * *")
	fmt.Printf("secure_cookie_signing_key = '%s'\n", hex.EncodeToString(securecookie.GenerateRandomKey(64)))
	fmt.Printf("secure_cookie_encryption_key = '%s'\n", hex.EncodeToString(securecookie.GenerateRandomKey(32)))
	fmt.Println("* * * Finished generating new secure keys * * *")
	fmt.Println("Set config option 'regenerate_secure_keys' to false (no quotes) to permit the application to start normally")
}

func generateConfig(filename string, wd string) error {
	configPath := filepath.Join(wd, fmt.Sprintf("%s.toml", filename))

	// Abort if file exists
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("%s already exists", configPath)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("os.Stat(): %w", err)
	}

	// Generate fresh secure cookie keys
	signingKey := hex.EncodeToString(securecookie.GenerateRandomKey(64))
	encryptionKey := hex.EncodeToString(securecookie.GenerateRandomKey(32))

	// Prep file contents with helpful comments and good defaults
	configContent := fmt.Sprintf(`# debug_config = true

# Web Server Configuration
# These values can be overridden via environment variables using ${VAR_NAME} syntax
host_port = '8080'        # Can also use: '${HOST_PORT}' to load from env var
cache_bust = 'abc'
# log_file = './log.log'  # Can also use: '${LOG_FILE}'. If not set, logs to STDOUT
log_level = 3             # Detail of logging (1-5): 1 is fatal-level only, 5 is trace-level detail. Level 3 is recommended for production.
cache_templates = false   # TURN ON (set to true) IN PRODUCTION for better performance!!!
ssl_disabled = true       # TURN OFF (set to false) IN PRODUCTION!!!
# ssl_cert_file = './tls/cert.pem'  # Can also use: '${SSL_CERT_FILE}'
# ssl_key_file = './tls/key.pem'    # Can also use: '${SSL_KEY_FILE}'

# Session/Cookie Configuration
secure_cookie_max_age = 0 # In seconds, e.g. 86400*7=604800; 0 means cookie persists for this session only
# Secure keys should be set via environment variables in production using ${ENV_VAR} syntax
# The keys below are freshly generated - use them for local development
secure_cookie_signing_key = '%s'
secure_cookie_encryption_key = '%s'
regenerate_secure_keys = false # Set to true and execute the binary to generate new keys and then exit
`, signingKey, encryptionKey)

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		return fmt.Errorf("os.WriteFile(): %w", err)
	}

	return nil
}

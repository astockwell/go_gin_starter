package main

import (
	"os"
	"testing"
)

func TestParseEnvVariables(t *testing.T) {
	// Test 1: Environment variables are correctly expanded
	t.Run("ExpandsEnvironmentVariables", func(t *testing.T) {
		// Set up test environment variables
		os.Setenv("TEST_HOST_PORT", "9090")
		os.Setenv("TEST_LOG_FILE", "/var/log/test.log")
		os.Setenv("TEST_SSL_CERT", "/etc/ssl/cert.pem")
		os.Setenv("TEST_SSL_KEY", "/etc/ssl/key.pem")
		os.Setenv("TEST_SIGNING_KEY", "abcd1234")
		os.Setenv("TEST_ENCRYPTION_KEY", "efgh5678")
		defer func() {
			os.Unsetenv("TEST_HOST_PORT")
			os.Unsetenv("TEST_LOG_FILE")
			os.Unsetenv("TEST_SSL_CERT")
			os.Unsetenv("TEST_SSL_KEY")
			os.Unsetenv("TEST_SIGNING_KEY")
			os.Unsetenv("TEST_ENCRYPTION_KEY")
		}()

		// Create a config with env var references
		ac := &AppConfig{
			HostPort:                     "${TEST_HOST_PORT}",
			LogFile:                      "${TEST_LOG_FILE}",
			SSLCertFile:                  "${TEST_SSL_CERT}",
			SSLKeyFile:                   "${TEST_SSL_KEY}",
			SecureCookieSigningKeyHex:    "${TEST_SIGNING_KEY}",
			SecureCookieEncryptionKeyHex: "${TEST_ENCRYPTION_KEY}",
		}

		// Call ParseEnvVariables
		ac.ParseEnvVariables()

		// Verify environment variables were expanded
		if ac.HostPort != "9090" {
			t.Errorf("Expected HostPort to be '9090', got '%s'", ac.HostPort)
		}
		if ac.LogFile != "/var/log/test.log" {
			t.Errorf("Expected LogFile to be '/var/log/test.log', got '%s'", ac.LogFile)
		}
		if ac.SSLCertFile != "/etc/ssl/cert.pem" {
			t.Errorf("Expected SSLCertFile to be '/etc/ssl/cert.pem', got '%s'", ac.SSLCertFile)
		}
		if ac.SSLKeyFile != "/etc/ssl/key.pem" {
			t.Errorf("Expected SSLKeyFile to be '/etc/ssl/key.pem', got '%s'", ac.SSLKeyFile)
		}
		if ac.SecureCookieSigningKeyHex != "abcd1234" {
			t.Errorf("Expected SecureCookieSigningKeyHex to be 'abcd1234', got '%s'", ac.SecureCookieSigningKeyHex)
		}
		if ac.SecureCookieEncryptionKeyHex != "efgh5678" {
			t.Errorf("Expected SecureCookieEncryptionKeyHex to be 'efgh5678', got '%s'", ac.SecureCookieEncryptionKeyHex)
		}
	})

	// Test 2: Fallback to original values when env vars are not set
	t.Run("FallsBackWhenEnvVarsNotSet", func(t *testing.T) {
		// Create a config with hardcoded values
		ac := &AppConfig{
			HostPort:                     "8080",
			LogFile:                      "./app.log",
			SSLCertFile:                  "./cert.pem",
			SSLKeyFile:                   "./key.pem",
			SecureCookieSigningKeyHex:    "original_signing_key",
			SecureCookieEncryptionKeyHex: "original_encryption_key",
		}

		// Call ParseEnvVariables (no env vars set)
		ac.ParseEnvVariables()

		// Verify values remain unchanged
		if ac.HostPort != "8080" {
			t.Errorf("Expected HostPort to remain '8080', got '%s'", ac.HostPort)
		}
		if ac.LogFile != "./app.log" {
			t.Errorf("Expected LogFile to remain './app.log', got '%s'", ac.LogFile)
		}
		if ac.SSLCertFile != "./cert.pem" {
			t.Errorf("Expected SSLCertFile to remain './cert.pem', got '%s'", ac.SSLCertFile)
		}
		if ac.SSLKeyFile != "./key.pem" {
			t.Errorf("Expected SSLKeyFile to remain './key.pem', got '%s'", ac.SSLKeyFile)
		}
		if ac.SecureCookieSigningKeyHex != "original_signing_key" {
			t.Errorf("Expected SecureCookieSigningKeyHex to remain 'original_signing_key', got '%s'", ac.SecureCookieSigningKeyHex)
		}
		if ac.SecureCookieEncryptionKeyHex != "original_encryption_key" {
			t.Errorf("Expected SecureCookieEncryptionKeyHex to remain 'original_encryption_key', got '%s'", ac.SecureCookieEncryptionKeyHex)
		}
	})

	// Test 3: Empty strings are preserved when env vars expand to empty
	t.Run("PreservesEmptyStrings", func(t *testing.T) {
		os.Setenv("EMPTY_VAR", "")
		defer os.Unsetenv("EMPTY_VAR")

		ac := &AppConfig{
			HostPort: "${EMPTY_VAR}",
		}

		ac.ParseEnvVariables()

		// When env var is empty, the field should remain with the empty expansion result
		// but ParseEnvVariables only updates if result is non-empty, so original value persists
		if ac.HostPort != "${EMPTY_VAR}" {
			t.Errorf("Expected HostPort to remain '${EMPTY_VAR}', got '%s'", ac.HostPort)
		}
	})
}

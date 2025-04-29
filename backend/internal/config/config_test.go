package config

import (
	"flag"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConfig_EnvPriority(t *testing.T) {
	os.Setenv("SERVER_ADDRESS", "env-address")
	os.Setenv("MONGO_URI", "env-mongo")
	os.Setenv("ACCESS_TOKEN_EXPIRES", "30m")
	os.Setenv("REFRESH_TOKEN_EXPIRES", "2h")
	os.Setenv("ACCESS_TOKEN_SECRET", "env-access-secret")
	os.Setenv("REFRESH_TOKEN_SECRET", "env-refresh-secret")
	t.Cleanup(func() {
		os.Clearenv()
	})

	cfg := GetConfig()

	assert.Equal(t, "env-address", cfg.ServerAddress)
	assert.Equal(t, "env-mongo", cfg.MongoURI)
	assert.Equal(t, 30*time.Minute, cfg.AccessTokenExpires)
	assert.Equal(t, 2*time.Hour, cfg.RefreshTokenExpires)
	assert.Equal(t, "env-access-secret", cfg.AccessTokenSecret)
	assert.Equal(t, "env-refresh-secret", cfg.RefreshTokenSecret)
}

func TestConfig_FlagPriority(t *testing.T) {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	os.Args = []string{
		"cmd",
		"-a", "flag-address",
		"-m", "flag-mongo",
		"-access-exp", "45m",
		"-refresh-exp", "3h",
		"-access-secret", "flag-access-secret",
		"-refresh-secret", "flag-refresh-secret",
	}
	cfg := GetConfig()

	assert.Equal(t, "flag-address", cfg.ServerAddress)
	assert.Equal(t, "flag-mongo", cfg.MongoURI)
	assert.Equal(t, 45*time.Minute, cfg.AccessTokenExpires)
	assert.Equal(t, 3*time.Hour, cfg.RefreshTokenExpires)
	assert.Equal(t, "flag-access-secret", cfg.AccessTokenSecret)
	assert.Equal(t, "flag-refresh-secret", cfg.RefreshTokenSecret)
}

func TestConfig_JsonPriority(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "config*.json")
	assert.NoError(t, err)
	t.Cleanup(func() { os.Remove(tmpFile.Name()) })

	jsonContent := `{
		"server_address": "json-address",
		"mongo_uri": "json-mongo",
		"access_token_expires": "1h",
		"refresh_token_expires": "6h",
		"access_token_secret": "json-access-secret",
		"refresh_token_secret": "json-refresh-secret"
	}`
	_, err = tmpFile.WriteString(jsonContent)
	assert.NoError(t, err)
	tmpFile.Close()

	os.Args = []string{"cmd", "-c", tmpFile.Name()}
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	cfg := GetConfig()

	assert.Equal(t, "json-address", cfg.ServerAddress)
	assert.Equal(t, "json-mongo", cfg.MongoURI)
	assert.Equal(t, time.Hour, cfg.AccessTokenExpires)
	assert.Equal(t, 6*time.Hour, cfg.RefreshTokenExpires)
	assert.Equal(t, "json-access-secret", cfg.AccessTokenSecret)
	assert.Equal(t, "json-refresh-secret", cfg.RefreshTokenSecret)
}

func TestConfig_Defaults(t *testing.T) {
	os.Clearenv()
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	os.Args = []string{"cmd"}

	cfg := GetConfig()

	assert.Equal(t, cfg.DefaultServerAddress, cfg.ServerAddress)
	assert.Equal(t, cfg.DefaultMongoURI, cfg.MongoURI)
	assert.Equal(t, cfg.DefaultAccessTokenExpires, cfg.AccessTokenExpires)
	assert.Equal(t, cfg.DefaultRefreshTokenExpires, cfg.RefreshTokenExpires)
	assert.Equal(t, cfg.DefaultAccessTokenSecret, cfg.AccessTokenSecret)
	assert.Equal(t, cfg.DefaultRefreshTokenSecret, cfg.RefreshTokenSecret)
}

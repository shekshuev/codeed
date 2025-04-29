package config

import (
	"encoding/json"
	"flag"
	"os"
	"time"

	"github.com/caarlos0/env"
	_ "github.com/joho/godotenv/autoload"
	"github.com/shekshuev/codeed/backend/internal/logger"
	"go.uber.org/zap"
)

// Config stores all runtime configuration values loaded from flags, env or JSON.
// Priority: flags > env > json > defaults.
type Config struct {
	ServerAddress              string        // Address the server listens on
	MongoURI                   string        // MongoDB connection URI
	AccessTokenExpires         time.Duration // Access token expiration duration
	RefreshTokenExpires        time.Duration // Refresh token expiration duration
	AccessTokenSecret          string        // Secret used for signing access tokens
	RefreshTokenSecret         string        // Secret used for signing refresh tokens
	DefaultServerAddress       string        // Default address if not provided
	DefaultMongoURI            string        // Default Mongo URI if not provided
	DefaultAccessTokenExpires  time.Duration // Default duration for access token expiry
	DefaultRefreshTokenExpires time.Duration // Default duration for refresh token expiry
	DefaultAccessTokenSecret   string        // Default access token secret
	DefaultRefreshTokenSecret  string        // Default refresh token secret
}

type envConfig struct {
	ServerAddress       string        `env:"SERVER_ADDRESS"`
	MongoURI            string        `env:"MONGO_URI"`
	AccessTokenExpires  time.Duration `env:"ACCESS_TOKEN_EXPIRES"`
	RefreshTokenExpires time.Duration `env:"REFRESH_TOKEN_EXPIRES"`
	AccessTokenSecret   string        `env:"ACCESS_TOKEN_SECRET"`
	RefreshTokenSecret  string        `env:"REFRESH_TOKEN_SECRET"`
}

type jsonConfig struct {
	ServerAddress       string `json:"server_address"`
	MongoURI            string `json:"mongo_uri"`
	AccessTokenExpires  string `json:"access_token_expires"`
	RefreshTokenExpires string `json:"refresh_token_expires"`
	AccessTokenSecret   string `json:"access_token_secret"`
	RefreshTokenSecret  string `json:"refresh_token_secret"`
}

func GetConfig() Config {
	var cfg Config
	cfg.DefaultServerAddress = "localhost:3000"
	cfg.DefaultMongoURI = "mongo://localhost:27017"
	cfg.DefaultAccessTokenExpires = time.Hour
	cfg.DefaultRefreshTokenExpires = 24 * time.Hour
	cfg.DefaultAccessTokenSecret = "super_secret_access_token_key"
	cfg.DefaultRefreshTokenSecret = "super_secret_refresh_token_key"
	parseFlags(&cfg)
	parsEnv(&cfg)
	return cfg
}

func parseFlags(cfg *Config) {
	var configPath string
	if f := flag.Lookup("c"); f == nil {
		flag.StringVar(&configPath, "c", "", "path to JSON config file")
	} else if f := flag.Lookup("config"); f == nil {
		flag.StringVar(&configPath, "config", "", "path to JSON config file")
	}
	if f := flag.Lookup("a"); f == nil {
		flag.StringVar(&cfg.ServerAddress, "a", cfg.DefaultServerAddress, "address and port to run server")
	} else {
		cfg.ServerAddress = cfg.DefaultServerAddress
	}
	if f := flag.Lookup("m"); f == nil {
		flag.StringVar(&cfg.MongoURI, "m", cfg.DefaultMongoURI, "mongo connection string")
	} else {
		cfg.MongoURI = cfg.DefaultMongoURI
	}
	if f := flag.Lookup("access-exp"); f == nil {
		flag.DurationVar(&cfg.AccessTokenExpires, "access-exp", cfg.DefaultAccessTokenExpires, "access token expiration duration")
	} else {
		cfg.AccessTokenExpires = cfg.DefaultAccessTokenExpires
	}
	if f := flag.Lookup("refresh-exp"); f == nil {
		flag.DurationVar(&cfg.RefreshTokenExpires, "refresh-exp", cfg.DefaultRefreshTokenExpires, "refresh token expiration duration")
	} else {
		cfg.RefreshTokenExpires = cfg.DefaultRefreshTokenExpires
	}
	if f := flag.Lookup("access-secret"); f == nil {
		flag.StringVar(&cfg.AccessTokenSecret, "access-secret", cfg.DefaultAccessTokenSecret, "access token secret")
	} else {
		cfg.AccessTokenSecret = cfg.DefaultAccessTokenSecret
	}
	if f := flag.Lookup("refresh-secret"); f == nil {
		flag.StringVar(&cfg.RefreshTokenSecret, "refresh-secret", cfg.DefaultRefreshTokenSecret, "refresh token secret")
	} else {
		cfg.RefreshTokenSecret = cfg.DefaultRefreshTokenSecret
	}
	flag.Parse()
	parseJSON(configPath, cfg)
	parsEnv(cfg)
}

func parsEnv(cfg *Config) {
	l := logger.NewLogger()
	var envCfg envConfig
	err := env.Parse(&envCfg)
	if err != nil {
		l.Log.Error("Error starting server", zap.Error(err))
	}
	if len(envCfg.ServerAddress) > 0 {
		cfg.ServerAddress = envCfg.ServerAddress
	}
	if len(envCfg.MongoURI) > 0 {
		cfg.MongoURI = envCfg.MongoURI
	}
	if envCfg.AccessTokenExpires > 0 {
		cfg.AccessTokenExpires = envCfg.AccessTokenExpires
	}
	if envCfg.RefreshTokenExpires > 0 {
		cfg.RefreshTokenExpires = envCfg.RefreshTokenExpires
	}
	if envCfg.AccessTokenSecret != "" {
		cfg.AccessTokenSecret = envCfg.AccessTokenSecret
	}
	if envCfg.RefreshTokenSecret != "" {
		cfg.RefreshTokenSecret = envCfg.RefreshTokenSecret
	}
}

func parseJSON(path string, cfg *Config) {
	if path == "" {
		path = os.Getenv("CONFIG")
	}
	if path == "" {
		return
	}

	file, err := os.Open(path)
	if err != nil {
		logger.NewLogger().Log.Warn("Could not open config file", zap.Error(err))
		return
	}
	defer file.Close()

	var jCfg jsonConfig
	if err := json.NewDecoder(file).Decode(&jCfg); err != nil {
		logger.NewLogger().Log.Warn("Could not decode config JSON", zap.Error(err))
		return
	}

	if cfg.ServerAddress == cfg.DefaultServerAddress && jCfg.ServerAddress != "" {
		cfg.ServerAddress = jCfg.ServerAddress
	}
	if cfg.MongoURI == cfg.DefaultMongoURI && jCfg.MongoURI != "" {
		cfg.MongoURI = jCfg.MongoURI
	}
	if cfg.AccessTokenExpires == cfg.DefaultAccessTokenExpires && jCfg.AccessTokenExpires != "" {
		value, err := time.ParseDuration(jCfg.AccessTokenExpires)
		if err == nil {
			cfg.AccessTokenExpires = value
		}
	}
	if cfg.RefreshTokenExpires == cfg.DefaultRefreshTokenExpires && jCfg.RefreshTokenExpires != "" {
		value, err := time.ParseDuration(jCfg.RefreshTokenExpires)
		if err == nil {
			cfg.RefreshTokenExpires = value
		}
	}
	if cfg.AccessTokenSecret == cfg.DefaultAccessTokenSecret && jCfg.AccessTokenSecret != "" {
		cfg.AccessTokenSecret = jCfg.AccessTokenSecret
	}
	if cfg.RefreshTokenSecret == cfg.DefaultRefreshTokenSecret && jCfg.RefreshTokenSecret != "" {
		cfg.RefreshTokenSecret = jCfg.RefreshTokenSecret
	}
}

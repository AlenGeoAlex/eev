package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type DBType string

const (
	DBTypeSqlite       DBType = "sqlite"
	jwtSecretFilePath         = ".jwt_secret"
	jwtRefreshFilePath        = ".jwt_refresh_secret"
)

type AppConfig struct {
	Env               string  `env:"APP_ENV"  envDefault:"prod"`
	Port              int     `env:"APP_PORT" envDefault:"8080"`
	MaxUploadSizeInMB int     `env:"APP_MAX_UPLOAD_SIZE_IN_MB" envDefault:"1024"`
	Host              *string `env:"APP_HOST"`

	DB    DBConfig
	Jwt   JwtConfig
	OAuth GoogleOAuthConfig
	S3    S3Config
}

func (receiver AppConfig) IsProduction() {

}

type DBConfig struct {
	DbType DBType `env:"DB_TYPE" envDefault:"sqlite"`

	MaxOpenConns int           `env:"DB_MAX_OPEN_CONNS" envDefault:"10"`
	MaxIdleConns int           `env:"DB_MAX_IDLE_CONNS" envDefault:"5"`
	ConnMaxLife  time.Duration `env:"DB_CONN_MAX_LIFE"  envDefault:"5m"`

	SqlitePath string `env:"DB_SQLITE_PATH" envDefault:"./sqlite.db"`
}

type S3Config struct {
	AccessKey   string `env:"S3_ACCESS_KEY"`
	SecretKey   string `env:"S3_SECRET_KEY"`
	Region      string `env:"S3_REGION" envDefault:"auto"`
	Bucket      string `env:"S3_BUCKET"`
	EndpointURL string `env:"S3_ENDPOINT_URL"`
}

type JwtConfig struct {
	Secret        string `env:"JWT_SECRET"`
	RefreshSecret string `env:"JWT_REFRESH_SECRET"`
	Issuer        string `env:"JWT_ISSUER" envDefault:"me.eev.alenalex"`

	AccessTTL  time.Duration `env:"JWT_ACCESS_TTL"  envDefault:"1h"`
	RefreshTTL time.Duration `env:"JWT_REFRESH_TTL" envDefault:"480h"` // 20 days
}

type GoogleOAuthConfig struct {
	ClientID     string `env:"GOOGLE_CLIENT_ID,required"`
	ClientSecret string `env:"GOOGLE_CLIENT_SECRET,required"`
	RedirectURL  string `env:"GOOGLE_REDIRECT_URL,required"`
}

func (receiver AppConfig) IsDevelopment() bool {
	return receiver.Env == "dev"
}

func (cfg *DBConfig) ConnectionString() string {
	switch cfg.DbType {
	case DBTypeSqlite:
		return fmt.Sprintf("file:%s?_foreign_keys=on", cfg.SqlitePath)
	default:
		panic("unsupported database type: " + string(cfg.DbType))
	}
}

func (cfg *DBConfig) MigrationConnectionString() string {
	switch cfg.DbType {
	case DBTypeSqlite:
		return fmt.Sprintf("sqlite://%s?_foreign_keys=on", cfg.SqlitePath)
	default:
		panic("unsupported database type: " + string(cfg.DbType))
	}
}

func NewAppConfig() *AppConfig {
	_ = godotenv.Load()
	cfg := &AppConfig{}

	if err := env.Parse(cfg); err != nil {
		log.Fatal(err)
	}

	validate(&cfg.DB)

	cfg.Jwt.Secret = resolveJwtSecret(cfg.Jwt.Secret)
	cfg.Jwt.RefreshSecret = resolveRefreshSecret(cfg.Jwt.RefreshSecret)

	return cfg
}

func NewDBConfig() *DBConfig {
	_ = godotenv.Load()
	cfg := &DBConfig{}

	if err := env.Parse(cfg); err != nil {
		log.Fatal(err)
	}

	validate(cfg)
	return cfg
}

func resolveJwtSecret(envSecret string) string {
	if stored, err := loadSecretFromCustomFile(jwtSecretFilePath); err == nil {
		if envSecret != "" {
			log.Println("WARNING: JWT Secret is loaded from disk, but also provided in ENV. Using the persisted value. If you ever want to change it delete the .jwt_secret file")
		}
		return stored
	}

	secret := envSecret
	if secret == "" {
		generated, err := generateSecret(32)
		if err != nil {
			log.Fatal("failed to generate JWT secret: ", err)
		}
		secret = generated
		log.Println("JWT secret not found — generated a new one")
	} else {
		log.Println("JWT secret loaded from environment — persisting to disk")
	}

	if err := saveSecretToFile(secret); err != nil {
		log.Fatal("failed to persist JWT secret to disk: ", err)
	}

	return secret
}

func loadSecretFromCustomFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	secret := strings.TrimSpace(string(data))
	if secret == "" {
		return "", fmt.Errorf("secret file is empty")
	}
	return secret, nil
}

func saveSecretToCustomFile(path, secret string) error {
	return os.WriteFile(path, []byte(secret), 0600)
}

func saveSecretToFile(secret string) error {
	return os.WriteFile(jwtSecretFilePath, []byte(secret), 0600)
}

func generateSecret(bytes int) (string, error) {
	b := make([]byte, bytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func resolveRefreshSecret(envSecret string) string {
	if stored, err := loadSecretFromCustomFile(jwtRefreshFilePath); err == nil {
		if envSecret != "" {
			log.Println("WARNING: Refresh secret loaded from disk but also provided in ENV. Using persisted value.")
		}
		return stored
	}

	secret := envSecret
	if secret == "" {
		generated, err := generateSecret(32)
		if err != nil {
			log.Fatal("failed to generate refresh secret: ", err)
		}
		secret = generated
		log.Println("Refresh secret not found — generated a new one")
	} else {
		log.Println("Refresh secret loaded from environment — persisting to disk")
	}

	if err := saveSecretToCustomFile(jwtRefreshFilePath, secret); err != nil {
		log.Fatal("failed to persist refresh secret to disk: ", err)
	}

	return secret
}

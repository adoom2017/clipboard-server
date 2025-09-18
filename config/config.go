package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config application configuration structure
type Config struct {
	ServerHost string
	ServerPort string

	JWTSecret     string
	JWTExpireHour int

	DBPath  string
	DBDebug bool

	CORSAllowOrigins []string
	CORSAllowMethods []string
	CORSAllowHeaders []string

	LogLevel string
	LogFile  string

	MaxContentSize  int64
	CleanupDays     int
	EnableCleanup   bool
	CleanupInterval string

	RateLimitRPS   int
	RateLimitBurst int

	UploadMaxSize int64
	UploadPath    string
}

var AppConfig *Config

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		// .env file not existing is fine
	}

	config := &Config{
		ServerHost: getEnv("SERVER_HOST", "localhost"),
		ServerPort: getEnv("SERVER_PORT", "8080"),

		JWTSecret:     getEnv("JWT_SECRET", "clipboard-sync-secret-key-change-in-production"),
		JWTExpireHour: getEnvAsInt("JWT_EXPIRE_HOUR", 24*7),

		DBPath:  getEnv("DB_PATH", "data/clipboard.db"),
		DBDebug: getEnvAsBool("DB_DEBUG", false),

		CORSAllowOrigins: getEnvAsSlice("CORS_ALLOW_ORIGINS", []string{"*"}, ","),
		CORSAllowMethods: getEnvAsSlice("CORS_ALLOW_METHODS", []string{
			"GET", "POST", "PUT", "DELETE", "OPTIONS",
		}, ","),
		CORSAllowHeaders: getEnvAsSlice("CORS_ALLOW_HEADERS", []string{
			"Origin", "Content-Type", "Accept", "Authorization", "Cache-Control",
		}, ","),

		LogLevel: getEnv("LOG_LEVEL", "info"),
		LogFile:  getEnv("LOG_FILE", "logs/app.log"),

		MaxContentSize:  getEnvAsInt64("MAX_CONTENT_SIZE", 1024*1024),
		CleanupDays:     getEnvAsInt("CLEANUP_DAYS", 30),
		EnableCleanup:   getEnvAsBool("ENABLE_CLEANUP", true),
		CleanupInterval: getEnv("CLEANUP_INTERVAL", "0 2 * * *"),

		RateLimitRPS:   getEnvAsInt("RATE_LIMIT_RPS", 100),
		RateLimitBurst: getEnvAsInt("RATE_LIMIT_BURST", 200),

		UploadMaxSize: getEnvAsInt64("UPLOAD_MAX_SIZE", 10*1024*1024),
		UploadPath:    getEnv("UPLOAD_PATH", "data/uploads"),
	}

	AppConfig = config
	return config
}

func GetConfig() *Config {
	if AppConfig == nil {
		return LoadConfig()
	}
	return AppConfig
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(name string, defaultVal int) int {
	valueStr := getEnv(name, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultVal
}

func getEnvAsInt64(name string, defaultVal int64) int64 {
	valueStr := getEnv(name, "")
	if value, err := strconv.ParseInt(valueStr, 10, 64); err == nil {
		return value
	}
	return defaultVal
}

func getEnvAsBool(name string, defaultVal bool) bool {
	valueStr := getEnv(name, "")
	if value, err := strconv.ParseBool(valueStr); err == nil {
		return value
	}
	return defaultVal
}

func getEnvAsSlice(name string, defaultVal []string, sep string) []string {
	valueStr := getEnv(name, "")
	if valueStr == "" {
		return defaultVal
	}

	values := strings.Split(valueStr, sep)
	result := make([]string, 0, len(values))
	for _, v := range values {
		if trimmed := strings.TrimSpace(v); trimmed != "" {
			result = append(result, trimmed)
		}
	}

	if len(result) == 0 {
		return defaultVal
	}
	return result
}

func (c *Config) IsDevelopment() bool {
	env := getEnv("GO_ENV", "development")
	return env == "development" || env == "dev"
}

func (c *Config) IsProduction() bool {
	env := getEnv("GO_ENV", "development")
	return env == "production" || env == "prod"
}

func (c *Config) GetAddress() string {
	return c.ServerHost + ":" + c.ServerPort
}

func (c *Config) Validate() error {
	if c.JWTSecret == "clipboard-sync-secret-key-change-in-production" && c.IsProduction() {
		return fmt.Errorf("JWT_SECRET must be changed in production")
	}

	if c.MaxContentSize <= 0 {
		return fmt.Errorf("MAX_CONTENT_SIZE must be greater than 0")
	}

	if c.CleanupDays <= 0 {
		return fmt.Errorf("CLEANUP_DAYS must be greater than 0")
	}

	return nil
}

func (c *Config) Print() {
	fmt.Println("Clipboard Sync Server Configuration:")
	fmt.Println("  Server:", c.GetAddress())
	fmt.Println("  Environment:", getEnv("GO_ENV", "development"))
	fmt.Println("  Database Path:", c.DBPath)
	fmt.Println("  Log Level:", c.LogLevel)
	fmt.Printf("  Max Content Size: %d bytes\n", c.MaxContentSize)
	fmt.Println("  Cleanup Days:", c.CleanupDays)
	fmt.Printf("  Rate Limit: %d RPS, %d Burst\n", c.RateLimitRPS, c.RateLimitBurst)
}

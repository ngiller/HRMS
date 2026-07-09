package config

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	JWTSecret        string
	JWTAccessExpiry  time.Duration
	JWTRefreshExpiry time.Duration

	ServerPort string
	ServerHost string

	ResetTokenExpiry time.Duration

	FrontendURL string

	EncryptionKey string

	UploadDir string

	SMTPHost     string
	SMTPPort     string
	SMTPUser     string
	SMTPPassword string
	SMTPFrom     string
	SMTPFromName string

	VAPIDPublicKey  string
	VAPIDPrivateKey string

	RedisHost     string
	RedisPort     string
	RedisPassword string
	RedisDB       int
}

func Load() *Config {
	// Try to load .env from current directory and parent directory
	// (backend can be run from project root or backend/ folder)
	if err := godotenv.Load(); err != nil {
		log.Printf("INFO: No .env in current directory: %v", err)
	}
	if err := godotenv.Load("../.env"); err != nil {
		log.Printf("INFO: No .env in parent directory: %v", err)
	}

	cfg := &Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBName:     getEnv("DB_NAME", "hrms"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),

		JWTSecret:        getEnv("JWT_SECRET", ""),
		JWTAccessExpiry:  getDuration("JWT_ACCESS_EXPIRY", 15*time.Minute),
		JWTRefreshExpiry: getDuration("JWT_REFRESH_EXPIRY", 7*24*time.Hour),

		ServerPort: getEnv("SERVER_PORT", "8900"),
		ServerHost: getEnv("SERVER_HOST", "0.0.0.0"),

		ResetTokenExpiry: getDuration("RESET_TOKEN_EXPIRY", 1*time.Hour),

		FrontendURL: getEnv("FRONTEND_URL", "http://localhost:5173"),

		EncryptionKey: getEncryptionKey(),
		UploadDir:     getEnv("UPLOAD_DIR", "./uploads"),

		SMTPHost:     getEnv("SMTP_HOST", ""),
		SMTPPort:     getEnv("SMTP_PORT", "587"),
		SMTPUser:     getEnv("SMTP_USER", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),
		SMTPFrom:     getEnv("SMTP_FROM", "noreply@hrms.com"),
		SMTPFromName: getEnv("SMTP_FROM_NAME", "HRMS System"),

		VAPIDPublicKey:  getEnv("VAPID_PUBLIC_KEY", ""),
		VAPIDPrivateKey: getEnv("VAPID_PRIVATE_KEY", ""),

		RedisHost:     getEnv("REDIS_HOST", "localhost"),
		RedisPort:     getEnv("REDIS_PORT", "6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getInt("REDIS_DB", 0),
	}

	// Warn if sensitive defaults are used (development only)
	if cfg.DBPassword == "" {
		log.Println("⚠️  WARNING: DB_PASSWORD tidak di-set. Gunakan .env untuk production.")
	}
	if cfg.JWTSecret == "" {
		log.Println("⚠️  WARNING: JWT_SECRET tidak di-set. Menggunakan secret random.")
		log.Println("⚠️  Set JWT_SECRET di .env untuk persistensi token setelah restart.")
		cfg.JWTSecret = generateRandomSecret()
	}

	// Ensure upload directory exists
	if _, err := os.Stat(cfg.UploadDir); os.IsNotExist(err) {
		os.MkdirAll(cfg.UploadDir, 0755)
	}

	return cfg
}

func (c *Config) DatabaseURL() string {
	return "postgres://" + c.DBUser + ":" + url.QueryEscape(c.DBPassword) + "@" + c.DBHost + ":" + c.DBPort + "/" + c.DBName + "?sslmode=" + c.DBSSLMode
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		d, err := time.ParseDuration(value)
		if err == nil {
			return d
		}
	}
	return defaultValue
}

func getInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		i, err := strconv.Atoi(value)
		if err == nil {
			return i
		}
	}
	return defaultValue
}

func generateRandomSecret() string {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		log.Fatalf("FATAL: Failed to generate random JWT secret: %v", err)
	}
	return hex.EncodeToString(key)
}

func getEncryptionKey() string {
	if key := os.Getenv("ENCRYPTION_KEY"); key != "" {
		return key
	}
	// Generate a random 32-byte (256-bit) AES key automatically
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		log.Printf("WARNING: Failed to generate random encryption key, using fallback: %v", err)
		return "hrms-default-encryption-key-32bytes!!"
	}
	encoded := hex.EncodeToString(key)
	log.Println("🔐 INFO: ENCRYPTION_KEY tidak di-set. Auto-generated random encryption key.")
	log.Printf("🔐 INFO: Generated key: %s\n", encoded)
	log.Println("🔐 INFO: Copy key di atas ke .env sebagai ENCRYPTION_KEY=<key> agar data tetap terbaca setelah restart.")
	return encoded
}

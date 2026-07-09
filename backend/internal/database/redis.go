package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

// RedisConfig holds configuration for Redis connection
type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

// ConnectRedis initializes the Redis client connection
func ConnectRedis(cfg RedisConfig) error {
	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)

	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
		MinIdleConns: 3,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis connection failed: %w", err)
	}

	RedisClient = client
	log.Printf("✅ Redis connected at %s", addr)
	return nil
}

// CloseRedis closes the Redis connection
func CloseRedis() {
	if RedisClient != nil {
		if err := RedisClient.Close(); err != nil {
			log.Printf("⚠️ Error closing Redis: %v", err)
		} else {
			log.Println("Redis connection closed")
		}
	}
}

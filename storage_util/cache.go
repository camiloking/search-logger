package storage_util

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

// InitRedis connects to a real Redis server or starts a miniredis instance (for tests).
func InitRedis() *redis.Client {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		InitMockRedis()
	}

	password := os.Getenv("REDIS_PASSWORD") // optional
	db := 0
	if dbStr := os.Getenv("REDIS_DB"); dbStr != "" {
		if parsed, err := strconv.Atoi(dbStr); err == nil {
			db = parsed
		}
	}

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis at %s: %v", addr, err)
	}

	RedisClient = client
	return client
}

// InitMockRedis starts a miniredis server in-process for testing.
func InitMockRedis() *redis.Client {
	s, err := miniredis.Run()
	if err != nil {
		log.Fatalf("Failed to start miniredis: %v", err)
	}

	client := redis.NewClient(&redis.Options{
		Addr: s.Addr(),
	})

	RedisClient = client
	log.Println("Miniredis started at", s.Addr())
	return client
}

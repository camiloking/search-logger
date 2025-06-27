package cache

import (
	"context"
	"search-logger/storage_util"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func setupTestRedis(t *testing.T) *redis.Client {
	redisClient := storage_util.InitRedis()
	assert.NotNil(t, redisClient)
	return redisClient
}

func TestClientQueryCacheRepository_SetAndGet(t *testing.T) {
	cache := setupTestRedis(t)
	repo := NewLatestClientQueryCacheRepository(cache)

	t.Run("Set and retrieve a key", func(t *testing.T) {
		ctx := context.Background()
		key := "test-key"
		value := &ClientQueryValue{
			QueryText:                 "test-query",
			CreatedAtUnixMilliseconds: time.Now().Unix(),
		}

		// Test Set
		err := repo.Set(ctx, key, value)
		assert.NoError(t, err)

		// Test Get
		result, err := repo.Get(ctx, key)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, value.QueryText, result.QueryText)
		assert.Equal(t, value.CreatedAtUnixMilliseconds, result.CreatedAtUnixMilliseconds)
	})

	t.Run("Retrieve non-existent key", func(t *testing.T) {
		ctx := context.Background()
		key := "non-existent-key"

		// Test Get for a non-existent key
		result, err := repo.Get(ctx, key)
		assert.NoError(t, err)
		assert.Nil(t, result)
	})
}

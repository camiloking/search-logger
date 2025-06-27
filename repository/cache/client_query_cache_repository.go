package cache

import (
	"context"
	"encoding/json"
	"errors"
	"search-logger/config"
	"strings"

	"github.com/redis/go-redis/v9"
)

type ClientQueryValue struct {
	QueryText     string `json:"query_text"`
	CreatedAtUnix int64  `json:"created_at_unix"`
}

func NewClientQueryValue(queryText string, createdAtUnix int64) *ClientQueryValue {
	return &ClientQueryValue{
		QueryText:     strings.ToLower(strings.TrimSpace(queryText)),
		CreatedAtUnix: createdAtUnix,
	}
}

type LatestClientQueryCacheRepository interface {
	Get(ctx context.Context, key string) (*ClientQueryValue, error)
	Set(ctx context.Context, key string, value *ClientQueryValue) error
	Delete(ctx context.Context, key string) error
}

type latestClientQueryCacheRepository struct {
	cache *redis.Client
}

func NewLatestClientQueryCacheRepository(cache *redis.Client) LatestClientQueryCacheRepository {
	return &latestClientQueryCacheRepository{cache: cache}
}

func (c latestClientQueryCacheRepository) Get(ctx context.Context, key string) (*ClientQueryValue, error) {
	value, err := c.cache.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}

	var clientQueryValue ClientQueryValue
	if err = json.Unmarshal([]byte(value), &clientQueryValue); err != nil {
		return nil, err // JSON unmarshal error
	}

	return &clientQueryValue, nil
}

func (c latestClientQueryCacheRepository) Set(ctx context.Context, key string, value *ClientQueryValue) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	err = c.cache.Set(ctx, key, data, config.GetDefaultCacheTTLSeconds()).Err()
	if err != nil {
		return err
	}

	return nil
}

func (c latestClientQueryCacheRepository) Delete(ctx context.Context, key string) error {
	if key == "" {
		return errors.New("key cannot be empty")
	}

	err := c.cache.Del(ctx, key).Err()
	if err != nil {
		return err
	}

	return nil
}

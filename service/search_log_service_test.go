package service

import (
	"context"
	"log/slog"
	"search-logger/config"
	"search-logger/models"
	"search-logger/repository/cache"
	"search-logger/repository/database"
	"search-logger/storage_util"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func setupTestDatabase(t *testing.T) database.SearchLogRepository {
	db := storage_util.InitDB()
	assert.NotNil(t, db)
	err := db.AutoMigrate(&models.SearchLog{})
	assert.NoError(t, err)

	return database.NewSearchLogDatabaseRepository(db)
}

func setupTestRedis(t *testing.T) cache.LatestClientQueryCacheRepository {
	redisClient := storage_util.InitRedis()
	assert.NotNil(t, redisClient)
	return cache.NewLatestClientQueryCacheRepository(redisClient)
}

func TestSearchLogService_LogSearch(t *testing.T) {
	dbRepo := setupTestDatabase(t)
	cacheRepo := setupTestRedis(t)
	service := NewSearchLogService(dbRepo, cacheRepo, slog.Default())

	t.Run("Log search and persist to database", func(t *testing.T) {
		// ARRANGE
		ctx := context.Background()
		clientKey := "test-client-key"
		queryText := "  Test Query  "

		// ACT
		err := service.LogSearch(ctx, clientKey, queryText)
		assert.NoError(t, err)

		// ASSERT
		// Wait for the background goroutine to complete
		time.Sleep(config.GetLogSearchDebounceDelaySeconds() + time.Second)
		searchLog, err := dbRepo.GetByQueryText(ctx, "test query")
		assert.NoError(t, err)
		assert.Equal(t, "test query", searchLog.QueryText)
		assert.Equal(t, 1, searchLog.Count)
	})

	t.Run("Persist only full query string", func(t *testing.T) {
		// ARRANGE
		ctx := context.Background()
		clientKey := "test-client-key"
		fullQuery := "hello world"

		// ACT
		for i := 1; i <= len(fullQuery); i++ {
			substring := fullQuery[:i]
			err := service.LogSearch(ctx, clientKey, substring)
			assert.NoError(t, err)
			time.Sleep(250 * time.Millisecond) // Allow time for cache updates
		}

		// Wait for the background goroutine to complete
		time.Sleep(config.GetLogSearchDebounceDelaySeconds() + time.Second)

		// ASSERT
		for i := 1; i <= len(fullQuery); i++ {
			substring := fullQuery[:i]
			searchLog, err := dbRepo.GetByQueryText(ctx, substring)
			assert.NoError(t, err)
			if substring == fullQuery {
				assert.NotNil(t, searchLog)
				assert.Equal(t, fullQuery, searchLog.QueryText)
				assert.Equal(t, 1, searchLog.Count)
			} else {
				assert.Nil(t, searchLog)
			}
		}
	})

	t.Run("Only persist the full query string, for multiple client keys", func(t *testing.T) {
		// ARRANGE
		ctx := context.Background()
		clientKeys := []string{"client-key-1", "client-key-2", "client-key-3"}
		fullQuery := "abc defg hijk"

		// ACT
		for _, clientKey := range clientKeys {
			for i := 1; i <= len(fullQuery); i++ {
				substring := fullQuery[:i]
				err := service.LogSearch(ctx, clientKey, substring)
				assert.NoError(t, err)
				time.Sleep(250 * time.Millisecond) // Allow time for cache updates
			}
		}

		// Wait for the background goroutine to complete
		time.Sleep(config.GetLogSearchDebounceDelaySeconds() + time.Second)

		// ASSERT
		for i := 1; i <= len(fullQuery); i++ {
			substring := fullQuery[:i]
			searchLog, err := dbRepo.GetByQueryText(ctx, substring)
			assert.NoError(t, err)
			if substring == fullQuery {
				assert.NotNil(t, searchLog)
				assert.Equal(t, fullQuery, searchLog.QueryText)
				assert.Equal(t, len(clientKeys), searchLog.Count)
			} else {
				// For substrings, we expect no logs
				assert.Nil(t, searchLog)
			}
		}
	})

	t.Run("Do not persist if query text prepends the client's latest query", func(t *testing.T) {
		// ARRANGE
		ctx := context.Background()
		clientKey := "test-client-key"
		queryText := "foo-baz"

		// ACT
		err := service.LogSearch(ctx, clientKey, queryText)
		assert.NoError(t, err)
		time.Sleep(1 * time.Second) // Sleep before new request for same client key
		err = service.LogSearch(ctx, clientKey, queryText+"-bar")
		assert.NoError(t, err)

		// ASSERT
		// Wait for the background goroutine to complete
		time.Sleep(config.GetLogSearchDebounceDelaySeconds() + time.Second)
		searchLog, err := dbRepo.GetByQueryText(ctx, queryText)
		assert.Nil(t, searchLog)
	})

	t.Run("Do not persist query if a newer one has occurred after, even if initial query does not prepend the new one", func(t *testing.T) {
		// ARRANGE
		ctx := context.Background()
		clientKey := "test-client-key"
		queryText := "foo-baz"
		queryText2 := "next query"

		// ACT
		err := service.LogSearch(ctx, clientKey, queryText)
		assert.NoError(t, err)
		time.Sleep(1 * time.Second) // Sleep before new request for same client key
		err = service.LogSearch(ctx, clientKey, queryText2)
		assert.NoError(t, err)

		// ASSERT
		// Wait for the background goroutine to complete
		time.Sleep(config.GetLogSearchDebounceDelaySeconds() + time.Second)
		searchLog, err := dbRepo.GetByQueryText(ctx, queryText)
		assert.Nil(t, searchLog)
	})

	t.Run("Log search in database when user has multiple searches, where one search occurs more than LOG_SEARCH_DEBOUNCE_DELAY_SECONDS after the first", func(t *testing.T) {
		// ARRANGE
		ctx := context.Background()
		clientKey := "test-client-key"
		queryText1 := "the query"
		queryText2 := "the query with more text"

		// ACT
		err := service.LogSearch(ctx, clientKey, queryText1)
		assert.NoError(t, err)
		// Wait for debounce delay before logging second query
		time.Sleep(config.GetLogSearchDebounceDelaySeconds() + time.Second)
		err = service.LogSearch(ctx, clientKey, queryText2)
		assert.NoError(t, err)

		// ASSERT
		// Wait for the background goroutine to complete
		time.Sleep(config.GetLogSearchDebounceDelaySeconds() + time.Second)

		// Both queries should be logged, even if one is a prefix of the other
		searchLog, err := dbRepo.GetByQueryText(ctx, queryText1)
		assert.NoError(t, err)
		assert.Equal(t, 1, searchLog.Count)
		searchLog, err = dbRepo.GetByQueryText(ctx, queryText2)
		assert.NoError(t, err)
		assert.Equal(t, 1, searchLog.Count)
	})

	t.Run("Log only second search in database when user has consecutive searches that don't prepend one another, since they are made within LOG_SEARCH_DEBOUNCE_DELAY_SECONDS", func(t *testing.T) {
		// ARRANGE
		ctx := context.Background()
		clientKey := "test-client-key"
		queryText1 := "one query"
		queryText2 := "another query"

		// ACT
		err := service.LogSearch(ctx, clientKey, queryText1)
		assert.NoError(t, err)
		time.Sleep(500 * time.Millisecond)
		err = service.LogSearch(ctx, clientKey, queryText2)
		assert.NoError(t, err)

		// ASSERT
		// Wait for the background goroutine to complete
		time.Sleep(config.GetLogSearchDebounceDelaySeconds() + time.Second)

		// Both queries should be logged, since they're not prefixes of each other
		searchLog, err := dbRepo.GetByQueryText(ctx, queryText1)
		assert.NoError(t, err)
		assert.Nil(t, searchLog)
		searchLog, err = dbRepo.GetByQueryText(ctx, queryText2)
		assert.NoError(t, err)
		assert.Equal(t, 1, searchLog.Count)
	})
}

func TestSearchLogService_GetSearchLogCountByQueryText(t *testing.T) {
	dbRepo := setupTestDatabase(t)
	cacheRepo := setupTestRedis(t)
	service := NewSearchLogService(dbRepo, cacheRepo, slog.Default())

	t.Run("Retrieve count for existing query", func(t *testing.T) {
		// ARRANGE
		ctx := context.Background()
		queryText1 := "  Test Query  "
		queryText2 := "  Test Query 2  "

		// ACT
		numLogs1 := 5
		for i := 0; i < numLogs1; i++ {
			_, _ = dbRepo.IncrementSearchLog(ctx, queryText1)
		}
		numLogs2 := 100
		for i := 0; i < numLogs2; i++ {
			_, _ = dbRepo.IncrementSearchLog(ctx, queryText2)
		}

		count, err := service.GetSearchLogCountByQueryText(ctx, queryText1)
		assert.NoError(t, err)
		assert.Equal(t, numLogs1, count)
		count, err = service.GetSearchLogCountByQueryText(ctx, queryText2)
		assert.NoError(t, err)
		assert.Equal(t, numLogs2, count)
	})

	t.Run("Return 0 for non-existent query", func(t *testing.T) {
		ctx := context.Background()
		queryText := "Non-existent Query"

		count, err := service.GetSearchLogCountByQueryText(ctx, queryText)
		assert.NoError(t, err)
		assert.Equal(t, 0, count)
	})
}

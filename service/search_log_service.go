package service

import (
	"context"
	"fmt"
	"log/slog"
	"search-logger/config"
	"search-logger/repository/cache"
	"search-logger/repository/database"
	"strings"
	"time"
)

type SearchLogService interface {
	LogSearch(ctx context.Context, clientKey, queryText string) error
	GetSearchLogCountByQueryText(ctx context.Context, queryText string) (int, error)
}

type searchLogService struct {
	db     database.SearchLogRepository
	cache  cache.LatestClientQueryCacheRepository
	logger *slog.Logger
}

func NewSearchLogService(db database.SearchLogRepository, cache cache.LatestClientQueryCacheRepository, logger *slog.Logger) SearchLogService {
	return &searchLogService{
		db:     db,
		cache:  cache,
		logger: logger,
	}
}

func (sls searchLogService) LogSearch(ctx context.Context, clientKey, queryText string) error {
	currentNormalizedQueryText := strings.TrimSpace(strings.ToLower(queryText))

	// Immediately set the latest client search in cache
	currentQueryTimeUnix := time.Now().Unix()
	clientQueryValue := cache.NewClientQueryValue(currentNormalizedQueryText, currentQueryTimeUnix)
	if err := sls.cache.Set(ctx, clientKey, clientQueryValue); err != nil {
		return err
	}

	// In logSearchDelaySeconds seconds, make a request to store the search log in the database
	go func() {
		backgroundContext := context.Background()

		// Debounce before attempting to log to DB, in case client is still typing
		// Ideally, the front end would do some debouncing too.
		time.Sleep(config.GetLogSearchDebounceDelaySeconds())

		// Get latest user client query from cache
		latestClientQueryValue, err := sls.cache.Get(backgroundContext, clientKey)
		if err != nil {
			sls.logger.Error("Error getting latest client query from cache", "error", err)
			return
		}

		if !shouldPersistQuery(currentQueryTimeUnix, currentNormalizedQueryText, latestClientQueryValue) {
			return
		}

		if _, err := sls.db.IncrementSearchLog(backgroundContext, currentNormalizedQueryText); err != nil {
			sls.logger.Error("Error logging search", "error", err, "queryText", currentNormalizedQueryText)
		}

		if err := sls.cache.Delete(backgroundContext, clientKey); err != nil {
			sls.logger.Error("Error deleting client query from cache", "error", err, "clientKey", clientKey)
		}
	}()
	return nil
}

func shouldPersistQuery(currentQueryTimeUnix int64, normalizedQueryText string, latestClientQueryValue *cache.ClientQueryValue) bool {
	if latestClientQueryValue != nil {
		// If queryTimeUnix is less than latestClientQueryValue.CreatedAtUnix, do not persist
		if currentQueryTimeUnix < latestClientQueryValue.CreatedAtUnix {
			return false
		}

		// If the current query is a prefix of the latest client query, do not persist
		if len(normalizedQueryText) < len(latestClientQueryValue.QueryText) && strings.HasPrefix(latestClientQueryValue.QueryText, normalizedQueryText) {
			return false
		}
	}
	return true
}

// Used for testing
func (sls searchLogService) GetSearchLogCountByQueryText(ctx context.Context, queryText string) (int, error) {
	if queryText == "" {
		return 0, fmt.Errorf("query text cannot be empty")
	}

	normalizedQueryText := strings.TrimSpace(strings.ToLower(queryText))
	searchLog, err := sls.db.GetByQueryText(ctx, normalizedQueryText)
	if err != nil {
		return 0, fmt.Errorf("error getting search log count: %w", err)
	}
	if searchLog != nil {
		return searchLog.Count, nil
	}

	return 0, nil
}

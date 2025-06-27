package database

import (
	"context"
	"errors"
	"search-logger/models"
	"strings"

	"gorm.io/gorm"
)

type SearchLogRepository interface {
	IncrementSearchLog(ctx context.Context, queryText string) (*models.SearchLog, error)
	GetByQueryText(ctx context.Context, queryText string) (*models.SearchLog, error)
}

type searchLogDatabaseRepository struct {
	db *gorm.DB
}

func NewSearchLogDatabaseRepository(db *gorm.DB) SearchLogRepository {
	return &searchLogDatabaseRepository{db: db}
}

func (i searchLogDatabaseRepository) IncrementSearchLog(ctx context.Context, queryText string) (*models.SearchLog, error) {
	if queryText == "" {
		return nil, errors.New("search log cannot be nil")
	}

	queryText = strings.ToLower(strings.TrimSpace(queryText))
	var result *models.SearchLog
	err := i.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing models.SearchLog
		err := tx.Where("query_text = ?", queryText).First(&existing).Error

		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		if existing.ID != "" {
			existing.Count += 1
			if err = tx.Save(&existing).Error; err != nil {
				return err
			}
			result = &existing
		} else {
			searchLog := models.NewSearchLog(queryText, 1)
			if err = tx.Create(searchLog).Error; err != nil {
				return err
			}
			result = searchLog
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (i searchLogDatabaseRepository) GetByQueryText(ctx context.Context, queryText string) (*models.SearchLog, error) {
	if queryText == "" {
		return nil, errors.New("query text cannot be empty")
	}

	var searchLog models.SearchLog
	err := i.db.WithContext(ctx).Where("query_text = ?", strings.ToLower(strings.TrimSpace(queryText))).
		First(&searchLog).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &searchLog, nil
}

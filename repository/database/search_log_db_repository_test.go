package database

import (
	"context"
	"search-logger/models"
	"search-logger/storage_util"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db := storage_util.InitDB()
	assert.NotNil(t, db)

	err := db.AutoMigrate(&models.SearchLog{})
	assert.NoError(t, err)

	return db
}

func TestSearchLogDatabaseRepository_Upsert(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSearchLogDatabaseRepository(db)

	t.Run("Insert new record and lowercase query field", func(t *testing.T) {
		result, err := repo.IncrementSearchLog(context.Background(), "TEST-querY")
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "test-query", result.QueryText)
		assert.Equal(t, 1, result.Count)
	})

	t.Run("Update existing record", func(t *testing.T) {
		tq := "test-query"
		result, err := repo.IncrementSearchLog(context.Background(), tq)
		assert.NoError(t, err)
		result, err = repo.IncrementSearchLog(context.Background(), tq)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, tq, result.QueryText)
		assert.Equal(t, 3, result.Count) // Count should be updated
	})
}

func TestSearchLogDatabaseRepository_GetByQueryText(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSearchLogDatabaseRepository(db)

	t.Run("Retrieve existing record", func(t *testing.T) {
		searchLog := models.NewSearchLog("test-query", 1)
		db.Create(searchLog)

		result, err := repo.GetByQueryText(context.Background(), "test-query")
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "test-query", result.QueryText)
		assert.Equal(t, 1, result.Count)
	})

	t.Run("Retrieve non-existing record", func(t *testing.T) {
		result, err := repo.GetByQueryText(context.Background(), "non-existent-query")
		assert.NoError(t, err)
		assert.Nil(t, result)
	})
}

package models

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SearchLog struct {
	ID        string    `json:"id" gorm:"type:uuid;primaryKey"`
	QueryText string    `json:"query" gorm:"unique"`
	Count     int       `json:"count"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (*SearchLog) TableName() string {
	return "search_logs"
}

func NewSearchLog(queryText string, count int) *SearchLog {
	return &SearchLog{
		ID:        uuid.New().String(),
		QueryText: queryText,
		Count:     count,
	}
}

func (s *SearchLog) BeforeSave(_ *gorm.DB) (err error) {
	s.QueryText = strings.ToLower(strings.TrimSpace(s.QueryText))
	return nil
}

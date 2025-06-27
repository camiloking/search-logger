package api

import (
	"log/slog"
	"net/http"
	"search-logger/repository/cache"
	"search-logger/repository/database"
	"search-logger/service"

	"github.com/gin-gonic/gin"
)

type SearchLogRequest struct {
	QueryText string `json:"query_text"`
}

func RegisterRoutes(r *gin.Engine, dbRepo database.SearchLogRepository, cacheRepo cache.LatestClientQueryCacheRepository) {
	logger := slog.Default()
	srv := service.NewSearchLogService(dbRepo, cacheRepo, logger)

	// I did not write tests for this endpoint.  I just have it here to show where I would call LogSearch()
	r.POST("/search-logs", func(c *gin.Context) {
		var searchLog SearchLogRequest
		if err := c.ShouldBindJSON(&searchLog); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		//TODO: generate from userID header + IP address. userID would likely come from decoded JWT token.
		clientIdentifier := ""
		err := srv.LogSearch(c.Request.Context(), clientIdentifier, searchLog.QueryText)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusAccepted)
	})

}

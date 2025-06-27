package main

import (
	"log/slog"
	"os"
	"search-logger/api"
	"search-logger/repository/cache"
	"search-logger/repository/database"
	"search-logger/storage_util"

	"github.com/gin-gonic/gin"
)

func main() {
	logHandler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	slog.SetDefault(slog.New(logHandler))
	slog.Info("Starting Search Logger Service")

	// Initialize database and cache repositories
	postgresDB := storage_util.InitDB()
	redisCache := storage_util.InitRedis()
	dbRepo := database.NewSearchLogDatabaseRepository(postgresDB)
	cacheRepo := cache.NewLatestClientQueryCacheRepository(redisCache)

	// Register API routes
	r := gin.Default()
	api.RegisterRoutes(r, dbRepo, cacheRepo)
	r.Run(":8080")
}

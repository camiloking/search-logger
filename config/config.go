package config

import (
	"log"
	"os"
	"strconv"
	"time"
)

var (
	logSearchDebounceDelaySeconds int
	defaultCacheTTLSeconds        int
)

func init() {
	delayStr := os.Getenv("LOG_SEARCH_DEBOUNCE_DELAY_SECONDS")
	if delayStr == "" {
		logSearchDebounceDelaySeconds = 3
	} else {
		val, err := strconv.Atoi(delayStr)
		if err != nil {
			log.Fatalf("Invalid LOG_SEARCH_DEBOUNCE_DELAY_SECONDS: %v", err)
		}
		logSearchDebounceDelaySeconds = val
	}

	ttlStr := os.Getenv("DEFAULT_CACHE_TTL_SECONDS")
	if ttlStr == "" {
		defaultCacheTTLSeconds = 30
	} else {
		val, err := strconv.Atoi(ttlStr)
		if err != nil {
			log.Fatalf("Invalid DEFAULT_CACHE_TTL_SECONDS: %v", err)
		}
		defaultCacheTTLSeconds = val
	}
}

func GetLogSearchDebounceDelaySeconds() time.Duration {
	return time.Duration(logSearchDebounceDelaySeconds) * time.Second
}

func GetDefaultCacheTTLSeconds() time.Duration {
	return time.Duration(defaultCacheTTLSeconds) * time.Second
}

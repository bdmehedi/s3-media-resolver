package services

import (
	"database/sql"
	"fmt"
	"github.com/bdmehedi/s3-media-resolver/internal/config"
	"time"

	"github.com/redis/go-redis/v9"
)

type CacheService struct{}

func NewCacheService() *CacheService {
	return &CacheService{}
}

func (s *CacheService) Get(key string) (string, error) {
	switch config.AppConfig.CacheDriver {
	case "redis":
		return s.getRedisCache(key)
	case "sqlite":
		return s.getSQLiteCache(key)
	default:
		return "", fmt.Errorf("invalid cache driver")
	}
}

func (s *CacheService) Set(key, value string) error {
	switch config.AppConfig.CacheDriver {
	case "redis":
		return s.setRedisCache(key, value)
	case "sqlite":
		return s.setSQLiteCache(key, value)
	default:
		return fmt.Errorf("invalid cache driver")
	}
}

func (s *CacheService) Clear(key string) error {
	switch config.AppConfig.CacheDriver {
	case "redis":
		return s.clearRedisCache(key)
	case "sqlite":
		return s.clearSQLiteCache(key)
	default:
		return fmt.Errorf("invalid cache driver")
	}
}

func (s *CacheService) getRedisCache(key string) (string, error) {
	result, err := config.AppConfig.Cache.(*redis.Client).Get(config.AppConfig.Context, key).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("cache not found")
	}
	return result, err
}

func (s *CacheService) setRedisCache(key, value string) error {
	return config.AppConfig.Cache.(*redis.Client).Set(
		config.AppConfig.Context,
		key,
		value,
		config.AppConfig.Expiry,
	).Err()
}

func (s *CacheService) clearRedisCache(key string) error {
	return config.AppConfig.Cache.(*redis.Client).Del(config.AppConfig.Context, key).Err()
}

func (s *CacheService) getSQLiteCache(key string) (string, error) {
	var value string
	var expiry int64

	err := config.AppConfig.DB.QueryRow(
		"SELECT value, expiry FROM cache WHERE key = ?",
		key,
	).Scan(&value, &expiry)

	if err == sql.ErrNoRows {
		return "", fmt.Errorf("cache not found")
	}
	if err != nil {
		return "", err
	}

	if time.Now().Unix() > expiry {
		s.clearSQLiteCache(key)
		return "", fmt.Errorf("cache expired")
	}

	return value, nil
}

func (s *CacheService) setSQLiteCache(key, value string) error {
	expiryTime := time.Now().Add(config.AppConfig.Expiry).Unix()
	_, err := config.AppConfig.DB.Exec(
		"INSERT OR REPLACE INTO cache (key, value, expiry) VALUES (?, ?, ?)",
		key, value, expiryTime,
	)
	return err
}

func (s *CacheService) clearSQLiteCache(key string) error {
	_, err := config.AppConfig.DB.Exec("DELETE FROM cache WHERE key = ?", key)
	return err
}

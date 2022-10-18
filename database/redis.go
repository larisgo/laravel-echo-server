package database

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/larisgo/laravel-echo-server/options"
	"regexp"
)

type RedisDatabase struct {

	// Redis client.
	redis *redis.Client

	// Configurable server options.
	options *options.Config

	ctx    context.Context
	cancel context.CancelFunc
}

// Create a new cache instance.
func NewRedisDatabase(_options *options.Config) (DatabaseDriver, error) {
	db := &RedisDatabase{}
	db.ctx, db.cancel = context.WithCancel(context.Background())
	db.redis = redis.NewClient(&redis.Options{
		Addr:     _options.DatabaseConfig.Redis.Host + ":" + _options.DatabaseConfig.Redis.Port,
		Username: _options.DatabaseConfig.Redis.Username,
		Password: _options.DatabaseConfig.Redis.Password,
		DB:       _options.DatabaseConfig.Redis.Db,
	})

	if _, err := db.redis.Ping(db.ctx).Result(); err != nil {
		return nil, errors.New(fmt.Sprintf("Redis connection failed: %v", err))
	}
	db.options = _options
	return db, nil
}

func (db *RedisDatabase) Close() error {
	return db.redis.Close()
}

// Retrieve data from redis.
func (db *RedisDatabase) Get(key string) ([]byte, error) {
	data, err := db.redis.Get(db.ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}
	return data, nil
}

// Store data to cache.
func (db *RedisDatabase) Set(key string, value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	if err := db.redis.Set(db.ctx, key, data, 0).Err(); err != nil {
		return err
	}
	if db.options.DatabaseConfig.PublishPresence == true && regexp.MustCompile(`^presence-.*:members$`).MatchString(key) {
		result, err := json.Marshal(map[string]map[string]any{
			"event": map[string]any{
				"channel": key,
				"members": value,
			},
		})
		if err != nil {
			return err
		}
		return db.redis.Publish(db.ctx, `PresenceChannelUpdated`, result).Err()
	}
	return nil
}

package database

import (
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/larisgo/laravel-echo-server/log"
	"github.com/larisgo/laravel-echo-server/options"
	"regexp"
)

type RedisDatabase struct {
	/**
	 * Redis client.
	 */
	redis *redis.Client

	/**
	 * Configurable server options.
	 */
	options options.Config
}

/**
 * Create a new cache instance.
 */
func NewRedisDatabase(Options options.Config) DatabaseDriver {
	this := &RedisDatabase{}
	this.redis = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf(`%s:%s`, Options.DatabaseConfig.Redis.Host, Options.DatabaseConfig.Redis.Port),
		Password: Options.DatabaseConfig.Redis.Password,
		DB:       Options.DatabaseConfig.Redis.Db,
	})
	// defer this.redis.Close()
	if _, err := this.redis.Ping().Result(); err != nil {
		log.Fatal(err)
	}
	this.options = Options
	return DatabaseDriver(this)
}

/**
 * Retrieve data from redis.
 */
func (this *RedisDatabase) Get(key string) (interface{}, error) {
	data, err := this.redis.Get(key).Result()
	if err != nil {
		// if err == redis.Nil {
		// 	return nil, nil
		// }
		return nil, err
	}
	var json_data interface{}
	if err := json.Unmarshal([]byte(data), &json_data); err != nil {
		return nil, err
	}
	return json_data, nil
}

/**
 * Store data to cache.
 */
func (this *RedisDatabase) Set(key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	if err := this.redis.Set(key, data, 0).Err(); err != nil {
		return err
	}
	if this.options.DatabaseConfig.PublishPresence == true && regexp.MustCompile(`^presence-.*:members`).MatchString(key) {
		result, err := json.Marshal(map[string]map[string]interface{}{
			"event": map[string]interface{}{
				"channel": key,
				"members": data,
			},
		})
		if err != nil {
			return err
		}
		this.redis.Publish(`PresenceChannelUpdated`, result)
	}
	return nil
}

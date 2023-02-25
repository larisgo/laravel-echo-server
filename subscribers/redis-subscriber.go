package subscribers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/larisgo/laravel-echo-server/options"
	"github.com/larisgo/laravel-echo-server/types"
	"github.com/redis/go-redis/v9"
	"github.com/zishang520/engine.io/utils"
)

type RedisSubscriber struct {

	// Redis client.
	redis *redis.Client

	// Configurable server options.
	options *options.Config

	// KeyPrefix for used in the redis Connection.
	keyPrefix string

	ctx    context.Context
	cancel context.CancelFunc
}

// Create a new instance of subscriber.
func NewRedisSubscriber(_options *options.Config) (Subscriber, error) {
	sub := &RedisSubscriber{}
	sub.ctx, sub.cancel = context.WithCancel(context.Background())
	sub.keyPrefix = _options.DatabaseConfig.Redis.KeyPrefix
	sub.redis = redis.NewClient(&redis.Options{
		Addr:     _options.DatabaseConfig.Redis.Host + ":" + _options.DatabaseConfig.Redis.Port,
		Username: _options.DatabaseConfig.Redis.Username,
		Password: _options.DatabaseConfig.Redis.Password,
		DB:       _options.DatabaseConfig.Redis.Db,
	})
	if _, err := sub.redis.Ping(sub.ctx).Result(); err != nil {
		return nil, errors.New(fmt.Sprintf("redis connection failed: %v", err))
	}
	sub.options = _options
	return sub, nil
}

// Subscribe to events to broadcast.
func (sub *RedisSubscriber) Subscribe(callback Broadcast) {
	pubsub := sub.redis.PSubscribe(sub.ctx, sub.keyPrefix+"*")
	utils.Log().Success("Listening for redis events...")
	go func() {
		defer pubsub.Close()
	LOOP:
		for {
			select {
			case <-sub.ctx.Done():
				break LOOP
			default:
				// ReceiveTimeout is a low level API. Use ReceiveMessage instead.
				msg, err := pubsub.ReceiveMessage(sub.ctx)
				if err != nil {
					if sub.options.DevMode {
						utils.Log().Error("%v", err)
					}
					break LOOP
				}
				var message *types.Data
				if err := json.Unmarshal([]byte(msg.Payload), &message); err != nil {
					if sub.options.DevMode {
						utils.Log().Error("%v", err)
					}
					break LOOP
				}
				channel := strings.TrimPrefix(msg.Channel, sub.keyPrefix)
				if sub.options.DevMode {
					utils.Log().Info("Channel: " + channel)
					utils.Log().Info("Event: " + message.Event)
				}
				callback(channel, message)
			}
		}
	}()
}

// Unsubscribe from events to broadcast.
func (sub *RedisSubscriber) UnSubscribe() {
	sub.cancel()
	sub.redis.Close()
}

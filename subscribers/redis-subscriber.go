package subscribers

import (
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/larisgo/laravel-echo-server/log"
	"github.com/larisgo/laravel-echo-server/options"
	"github.com/larisgo/laravel-echo-server/types"
	"strings"
)

type RedisSubscriber struct {
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
 * Create a new instance of subscriber.
 *
 * @param {any} options
 */
func NewRedisSubscriber(Options options.Config) Subscriber {
	this := &RedisSubscriber{}
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
	return Subscriber(this)
}

/**
 * Subscribe to events to broadcast.
 *
 * @return {Promise<any>}
 */
func (this *RedisSubscriber) Subscribe(callback Broadcast) {
	pubsub := this.redis.PSubscribe("*")
	log.Success("Listening for redis events...")
	// runtime
	go func() {
		for {
			// ReceiveTimeout is a low level API. Use ReceiveMessage instead.
			msg, err := pubsub.ReceiveMessage()
			if err != nil {
				if this.options.DevMode {
					log.Error(err)
				}
				break
			}
			var message types.Data
			if err := json.Unmarshal([]byte(msg.Payload), &message); err != nil {
				if this.options.DevMode {
					log.Error(err)
				}
				break
			}
			channel := strings.TrimPrefix(msg.Channel, this.options.DatabaseConfig.Prefix)
			if this.options.DevMode {
				log.Info(fmt.Sprintf("Channel: %s", channel))
				log.Info(fmt.Sprintf("Event: %s", message.Event))
			}
			callback(channel, message)
		}
	}()
}

package redis

import (
	"strconv"
	"time"

	"github.com/gophergala2016/togepi/util"
	"gopkg.in/redis.v3"
)

// Redis contains redis connection data.
type Redis struct {
	client *redis.Client
}

// NewClient returns new redis client connection.
func NewClient(host string, db int) (r *Redis, err error) {
	c := redis.NewClient(&redis.Options{
		Addr: host,
		DB:   int64(db),
	})

	_, err = c.Ping().Result()
	if err != nil {
		return
	}

	r = &Redis{
		client: c,
	}

	return
}

// SetGlobalKey generates and records the global secret key.
func (r *Redis) SetGlobalKey() error {
	key, keyErr := util.RandomString(16)
	if keyErr != nil {
		return keyErr
	}

	setErr := r.client.Set("secret", key, 0).Err()
	if setErr != nil {
		return setErr
	}

	return nil
}

// AddUser adds a new user hash.
func (r *Redis) AddUser(id, key string) (err error) {
	err = r.client.HMSet(id, "timestamp", strconv.FormatInt(time.Now().UTC().Unix(), 10), "key", key).Err()
	return
}

// Close closes the client connection.
func (r *Redis) Close() {
	r.client.Close()
}

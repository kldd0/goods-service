package redis

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	timeout  = time.Second * 15
	entryTTL = time.Minute * 20

	ErrKeyNotFound = errors.New("key not found")
)

type configGetter interface {
	RedisUri() string
	RedisPass() string
}

type Client struct {
	rdb *redis.Client
}

func New(ctx context.Context, config configGetter) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         config.RedisUri(),
		Password:     config.RedisPass(),
		ReadTimeout:  timeout,
		WriteTimeout: timeout,
		DB:           0, // use default DB
	})

	return &Client{rdb: rdb}, rdb.Ping(ctx).Err()
}

/*
func (c *Client) Delete(ctx context.Context, key string) error {
	res := c.rdb.Del(ctx, key)

	err := res.Err()
	if err != nil {
		return err
	}

	return nil
}
*/

package redis

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kldd0/goods-service/internal/domain/models"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
)

func (c *Client) GetGood(ctx context.Context, key string) (models.Good, error) {
	const op = "redis.GetGood"

	res := c.rdb.Get(ctx, key)
	if errors.Is(res.Err(), redis.Nil) {
		// this value is not in the cache
		return models.Good{}, nil
	}

	var good models.Good
	if err := res.Scan(&good); err != nil {
		return models.Good{}, fmt.Errorf("%s: value scanning error: %w", op, err)
	}

	return good, nil
}

func (c *Client) SetGood(ctx context.Context, key string, value models.Good) error {
	const op = "redis.SetGood"

	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("%s: failed marshalling data: %w", op, err)
	}

	if err := c.rdb.Set(ctx, key, data, entryTTL).Err(); err != nil {
		return fmt.Errorf("%s: set value error: %w", op, err)
	}

	return nil
}

package redis

import (
	"fmt"

	"search-radius/go-common/pkg/settings"
)

// NewConnection creates and returns a new Redis client
func NewConnection(cfg *settings.Redis) (*RedisEngine, error) {
	engine := &RedisEngine{
		config: cfg,
	}

	if err := engine.connect(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrConnectionFailed, err)
	}

	return engine, nil
}

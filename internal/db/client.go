package db

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/wolzey/taskboard/internal/scripts"
)

type Redis struct {
	*redis.Client
	Scripts *scripts.Scripts
}

func NewClient(opts *redis.Options) (*Redis, error) {
	client := redis.NewClient(opts)

	ctx := context.Background()
	
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("Failed to connect to Redis: %v\n", err)
	}

	s, err := scripts.LoadScripts(client)

	if err != nil {
		return nil, fmt.Errorf("Failed to load scripts: %v\n", err)
	}

	return &Redis{
		Client: client,
		Scripts: s,
	}, nil
}


package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wolzey/taskboard/internal/api"
	"github.com/wolzey/taskboard/internal/app"
	"github.com/wolzey/taskboard/internal/config"
)

var serveCmd = &cobra.Command{
    Use:   "serve",
    Short: "Starts the taskboard server",
    RunE: func(cmd *cobra.Command, args []string) error {
		// Load configuration
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Convert Redis config to redis.Options
		redisOpts, err := cfg.Redis.ToRedisOptions()
		if err != nil {
			return fmt.Errorf("failed to create Redis options: %w", err)
		}

		a := app.NewApp(&app.AppOptions{
			RedisOpts: redisOpts,
			ApiOptions: &api.ApiOptions{
				Port: cfg.API.Port,
			},
			QueuePrefix: cfg.Queue.Prefix,
		})

		a.Api.Serve(context.Background())

		return nil
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}



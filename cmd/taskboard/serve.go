package main

import (
	"context"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/cobra"
	"github.com/wolzey/taskboard/internal/api"
	"github.com/wolzey/taskboard/internal/app"
)

var serveCmd = &cobra.Command{
    Use:   "serve",
    Short: "Starts the taskboard server",
    RunE: func(cmd *cobra.Command, args []string) error {
		a := app.NewApp(&app.AppOptions{
			RedisOpts: &redis.Options{
				Addr: "localhost:6379",
			},
			ApiOptions: &api.ApiOptions{
				Port: 1337,
			},
		})


		a.Api.Serve(context.Background())

		return nil
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}



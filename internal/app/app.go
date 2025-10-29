package app

import (
	"context"
	"fmt"
	"sort"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/wolzey/taskboard/internal/api"
	"github.com/wolzey/taskboard/internal/db"
)

/**
App is the entrypoint for the entire application. This is where the API will be housed, the redis connection, and
all state / storage will be held.
**/

type queueState []string

var (
	allStates queueState = []string{"active", "wait", "delayed", "completed", "failed"}
)

func (a *queueState) ParseResults(res []int64) map[string]int64 {
	results := make(map[string]int64)

	for i, count := range res {
		results[allStates[i]] = count
	}

	return results
}

type App struct {
	Api         *api.Api
	Redis       *db.Redis
	QueuePrefix string
}

type AppOptions struct {
	RedisOpts   *redis.Options
	ApiOptions  *api.ApiOptions
	QueuePrefix string
}

type QueuesResponse struct {
	Queues []string `json:"queues"`
	Count  int      `json:"count"`
}

type CountsResponse struct {
	Counts map[string]int64 `json:"counts"`
}

type OverviewResponse struct {
	CountsByQueue map[string]map[string]int64 `json:"queue_counts"`
}

func NewApp(opts *AppOptions) *App {
	client, err := db.NewClient(opts.RedisOpts)

	if err != nil {
		fmt.Println("Unable to initialize redis client")
		panic(err)
	}

	routes := api.NewApi(*opts.ApiOptions)

	queuePrefix := opts.QueuePrefix
	if queuePrefix == "" {
		queuePrefix = "bull"
	}

	app := &App{
		Redis:       client,
		Api:         routes,
		QueuePrefix: queuePrefix,
	}

	app.Init()

	return app
}

func (a *App) Init() {
	a.Api.AddAPIHandler("/overview", "GET", a.GetJobsOverview)
	a.Api.AddAPIHandler("/queues", "GET", a.GetQueues)
	a.Api.AddAPIHandler("/queues/:queue", "GET", a.GetQueueDetails)
	a.Api.AddAPIHandler("/queues/:queue/:id", "GET", a.HandleGetJobDetails)
	a.Api.AddAPIHandler("/queues/:queue/:id/promote", "POST", a.HandlePromoteJob)
	a.Api.AddAPIHandler("/queues/:queue/jobs/:state", "GET", a.HandleListJobs)
}

func (a *App) GetQueueDetails(ctx *gin.Context) (int, any, error) {
	queueName := a.withPrefix(ctx.Param("queue"))

	results, err := a.Redis.Scripts.GetJobCounts(context.Background(), queueName, allStates)

	if err != nil {
		return 500, nil, err
	}

	parsed := allStates.ParseResults(results)

	return 200, CountsResponse{Counts: parsed}, nil
}

func (a *App) GetQueues(ctx *gin.Context) (int, any, error) {
	results, err := a.Redis.Scripts.GetQueues(context.Background(), a.QueuePrefix+":")

	if err != nil {
		return 400, nil, err
	}

	sort.Slice(results, func(i int, j int) bool {
		return results[i] < results[j]
	})

	return 200, QueuesResponse{Queues: results, Count: len(results)}, nil
}

func (a *App) GetJobsOverview(ctx *gin.Context) (int, any, error) {
	final := make(map[string]map[string]int64)

	results, err := a.Redis.Scripts.GetQueues(context.Background(), a.QueuePrefix+":")

	if err != nil {
		return 500, nil, err
	}

	for _, v := range results {
		fullKey := a.withPrefix(v)
		counts, err := a.Redis.Scripts.GetJobCounts(context.Background(), fullKey, allStates)

		if err != nil {
			continue
		}

		final[v] = allStates.ParseResults(counts)
	}

	return 200, final, nil
}

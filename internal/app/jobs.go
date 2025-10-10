package app

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type Job struct {
	Name         string `redis:"name" json:"name"`
	ProcessedOn  int64  `redis:"processedOn" json:"processed_on"`
	JobData      string `redis:"data" json:"data"`
	StackTrace   string `redis:"stacktrace" json:"stacktrace"`
	FailedReason string `redis:"failedReason" json:"failed_reason"`
	FinishedOn   string `redis:"finisehdOn" json:"finished_on"`
	Timestamp    string `redis:"timestamp" json:"timestamp"`
	AttemptsMade int    `redis:"attemptsMade" json:"attempts_made"`
	Priority     int    `redis:"priority" json:"priority"`
	Delay        int    `redis:"delay" json:"delay"`
	Opts         string `redis:"opts" json:"options"`
}

type PaginatedJobResponse struct {
	Jobs       []*Job `json:"jobs"`
	TotalCount int64  `json:"totalCount"`
}

type ParsedJobResponse struct {
	Job
	StackTrace []string       `json:"stacktrace"`
	Data       map[string]any `json:"data"`
	Options    map[string]any `json:"options"`
}

func (a *App) GetJobDetails(queue string, id string) (*ParsedJobResponse, error) {
	var res Job
	cmd := a.Redis.Client.HGetAll(context.Background(), fmt.Sprintf("bull:%s:%s", queue, id))

	err := cmd.Scan(&res)

	if err != nil {
		return nil, err
	}

	var jsonData map[string]any
	if err := json.Unmarshal([]byte(res.JobData), &jsonData); err != nil {
		fmt.Printf("Error parsing job data into struct: %v\n", err)
		return nil, err
	}

	var jsonStack []string
	if err := json.Unmarshal([]byte(res.StackTrace), &jsonStack); err != nil {
		jsonStack = nil
	}

	var jsonOptions map[string]any
	if err := json.Unmarshal([]byte(res.Opts), &jsonOptions); err != nil {
		fmt.Printf("Error parsing job options into struct: %v\n", err)
		return nil, err
	}

	return &ParsedJobResponse{
		Job:        res,
		Data:       jsonData,
		StackTrace: jsonStack,
		Options:    jsonOptions,
	}, nil
}

func (a *App) withPrefix(args ...string) string {
	final := slices.Concat([]string{"bull"}, args)

	return strings.Join(final, ":")
}

func (a *App) totalCount(queue string, state string) (int64, error) {
	result := a.Redis.Client.ZCard(context.Background(), a.withPrefix(queue, state))

	count, err := result.Result()

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (a *App) listJobs(queue string, state string, start int64, stop int64) ([]string, error) {
	fmt.Println(a.withPrefix(queue, state))
	result := a.Redis.Client.ZRevRange(context.Background(), a.withPrefix(queue, state), start, stop)

	if result.Err() != nil {
		return nil, result.Err()
	}

	return result.Result()
}

func (a *App) HandleListJobs(ctx *gin.Context) (int, any, error) {
	state := ctx.Param("state")
	queue := ctx.Param("queue")
	start, _ := strconv.Atoi(ctx.Query("start"))
	stop, _ := strconv.Atoi(ctx.Query("stop"))

	if stop == 0 {
		stop = 25
	}

	fmt.Printf("Start: %d\nEnd: %d\nQueue: %s\n", start, stop, queue)
	results, err := a.listJobs(queue, state, int64(start), int64(stop))

	if err != nil {
		return 500, nil, err
	}

	count, err := a.totalCount(queue, state)

	if err != nil {
		return 500, nil, err
	}

	type Results struct {
		Count   int64 `json:"count"`
		Results any   `json:"results"`
	}

	jobs := []*ParsedJobResponse{}

	for _, jobId := range results {
		details, err := a.GetJobDetails(queue, jobId)

		if err != nil {
			fmt.Println("unable to get job details for job id:", jobId)
			continue
		}

		jobs = append(jobs, details)
	}

	return 200, Results{Count: count, Results: jobs}, nil
}

func (a *App) HandleGetJobDetails(ctx *gin.Context) (int, any, error) {
	queue := ctx.Param("queue")
	id := ctx.Param("id")

	results, err := a.GetJobDetails(queue, id)

	if err != nil {
		return 500, nil, err
	}

	return 200, results, nil
}

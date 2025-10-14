package app

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
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

type ListFilter struct {
	Val string `redis:"value"`
}

func Filter[T any](s []T, predicate func(T) bool) []T {
	filtered := make([]T, 0, len(s))

	for _, v := range s {
		if predicate(v) {
			filtered = append(filtered, v)
		}
	}

	return filtered
}

func (a *App) listJobs(queue string, state string, start int64, stop int64, filter *ListFilter) ([]string, error) {
	fmt.Println(a.withPrefix(queue, state))
	result := a.Redis.Client.ZRevRange(context.Background(), a.withPrefix(queue, state), start, stop)

	if result.Err() != nil {
		return nil, result.Err()
	}

	results, err := result.Result()

	if err != nil {
		return nil, err
	}

	if filter != nil {
		results = Filter(results, func(s string) bool {
			details, err := a.GetJobDetails(queue, s)

			if err != nil {
				return false
			}

			ds, err := json.Marshal(details.Data)

			if err != nil {
				return false
			}

			return strings.Contains(string(ds), filter.Val)
		})
	}

	return results, nil
}

func (a *App) HandleListJobs(ctx *gin.Context) (int, any, error) {
	var filter *ListFilter

	state := ctx.Param("state")
	queue := ctx.Param("queue")
	start, _ := strconv.Atoi(ctx.Query("start"))
	stop, _ := strconv.Atoi(ctx.Query("stop"))
	dataFilter := ctx.Query("filter")

	if stop == 0 {
		stop = 25
	}

	if dataFilter != "" {
		filter = &ListFilter{
			Val: dataFilter,
		}
	}

	results, err := a.listJobs(queue, state, int64(start), int64(stop), filter)

	if err != nil {
		fmt.Println("error in listing jobs from redis")
		return 500, nil, err
	}

	count, err := a.totalCount(queue, state)

	if err != nil {
		fmt.Println("error in getting total count")
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

type SerializedId string

func (s SerializedId) IsValid() bool {
	r, err := regexp.Compile("[0-9]+")

	if err != nil {
		fmt.Println("FAILED TO COMPILE REGEXP")
		panic(err)
	}

	return r.MatchString(string(s))
}

func (s SerializedId) String() string {
	return string(s)
}

func (a *App) HandleGetJobDetails(ctx *gin.Context) (int, any, error) {
	queue := ctx.Param("queue")
	id := SerializedId(ctx.Param("id"))

	if !id.IsValid() {
		return 500, nil, fmt.Errorf("invalid job id")
	}

	results, err := a.GetJobDetails(queue, id.String())

	if err != nil {
		return 500, nil, err
	}

	return 200, results, nil
}

package scripts

import (
	"context"
	"embed"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/redis/go-redis/v9"
)

type Scripts struct {
	scripts map[string]*redis.Script
	client  *redis.Client
}

//go:embed lua/*.lua
//go:embed lua/includes/*.lua
var scriptFiles embed.FS

func LoadScripts(client *redis.Client) (*Scripts, error) {
	sc := &Scripts{
		scripts: make(map[string]*redis.Script),
		client:  client,
	}

	entries, _ := scriptFiles.ReadDir("lua")

	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".lua") {
			content, err := scriptFiles.ReadFile(filepath.Join("lua", entry.Name()))

			if err != nil {
				return nil, fmt.Errorf("failed to read %s: %v", entry.Name(), err)
			}

			name := strings.TrimSuffix(entry.Name(), ".lua")
			script := redis.NewScript(string(content))
			sc.scripts[name] = script
		}
	}

	return sc, nil
}

func (s *Scripts) GetQueues(ctx context.Context, prefix string) ([]string, error) {
	script := s.scripts["getQueues"]
	args := []any{fmt.Sprintf("%s*", prefix), "0", "100"}
	cmd := script.Run(ctx, s.client, []string{}, args...)

	if cmd.Err() != nil {
		return nil, cmd.Err()
	}

	res, err := cmd.StringSlice()

	if err != nil {
		return nil, err
	}

	for i, v := range res {
		res[i] = strings.Replace(v, prefix, "", -1)
	}

	return res, nil
}

func (s *Scripts) GetJobCounts(ctx context.Context, queue string, states []string) ([]int64, error) {
	script := s.scripts["getCounts"]

	args := make([]any, len(states))

	for i, state := range states {
		args[i] = state
	}

	cmd := script.Run(ctx, s.client, []string{queue}, args...)

	if cmd.Err() != nil {
		return nil, cmd.Err()
	}

	ret, err := cmd.Int64Slice()

	if err != nil {
		fmt.Printf("Unexpected return type: %v", err)
		return nil, err
	}

	return ret, nil
}

type PaginatedJobOptions struct {
	Start int64
	Stop  int64
	State string
}

type PaginatedJobResponse struct {
	Count   int64 `redis:"count"`
	Results []any `redis:"results"`
}

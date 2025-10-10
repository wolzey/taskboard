package model

import (
	"github.com/redis/go-redis/v9"
)

type Queue struct {
	client *redis.Client
}


type JobCounts struct {
	Waiting int `json:"waiting"`
	Active int `json:"active"`
	Delayed int `json:"delayed"`
	Failed int `json:"failed"`
	Completed int `json:"completed"`
}

/**
	GetJobCounts returns the number of jobs in the queue for each state. Allowing a holistic view of the queue.
	This also allows you to pass keys to get the counts of one specific state.
*/
	
func (q *Queue) GetJobCounts() (*JobCounts, error) {
	return nil, nil
}


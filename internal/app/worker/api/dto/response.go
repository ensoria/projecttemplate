package dto

import "github.com/ensoria/worker/pkg/job"

type JobStatus struct {
	Id     string `json:"jobId"`
	Status string `json:"status"`
}

type JobCancelled struct {
	Id string `json:"jobId"`
}

type JobControlError struct {
	Message string `json:"message"`
}

type DeadLetterJobList struct {
	Jobs  []*job.JobData `json:"jobs"`
	Count int            `json:"count"`
}

type DeadLetterJobRetry struct {
	Id         string `json:"jobId"`
	Message    string `json:"message"`
	RetryCount int    `json:"retriedCount,omitempty"`
}

type DeadLetterJobDeleted struct {
	Id string `json:"jobId"`
}

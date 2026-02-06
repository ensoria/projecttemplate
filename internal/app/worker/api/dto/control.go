package dto

import "github.com/ensoria/worker/pkg/job"

type JobStatus struct {
	Id     string `json:"job_id"`
	Status string `json:"status"`
}

type JobCancelled struct {
	Id string `json:"job_id"`
}

type JobControlError struct {
	Message string `json:"message"`
}

type DeadLetterJobList struct {
	Jobs  []*job.JobData `json:"jobs"`
	Count int            `json:"count"`
}

type DeadLetterJobRetry struct {
	Id         string `json:"job_id"`
	Message    string `json:"message"`
	RetryCount int    `json:"retried_count,omitempty"`
}

type RetryByName struct {
	JobName string `json:"job_name"`
}

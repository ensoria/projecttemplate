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

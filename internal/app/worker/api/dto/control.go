package dto

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

package dto

type JobStatus struct {
	Id     string `json:"job_id"`
	Status string `json:"status"`
}

type JobControlError struct {
	Message string `json:"message"`
}

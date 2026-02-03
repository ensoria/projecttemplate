package dto

type TaskControl struct {
	Message string `json:"message"`
}

type PauseTask struct {
	Reason string `json:"reason"`
}

type DisableTask struct {
	Reason string `json:"reason"`
}

type TaskControlError struct {
	Message string `json:"message"`
}

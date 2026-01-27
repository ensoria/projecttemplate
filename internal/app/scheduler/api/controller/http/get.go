package http

import (
	"net/http"

	"github.com/ensoria/rest/pkg/rest"
	"github.com/ensoria/scheduler/pkg/scheduler"
)

// TODO: schedulerとworkerのGETハンドラを実装

type ListTasks struct {
	scheduler *scheduler.Scheduler
}

func NewListTasks(scheduler *scheduler.Scheduler) *ListTasks {
	return &ListTasks{
		scheduler: scheduler,
	}
}

func (c *ListTasks) Handle(r *rest.Request) *rest.Response {
	ctx := r.Context()
	statuses, err := c.scheduler.ListTaskStates(ctx)
	if err != nil {
		return &rest.Response{
			Code: http.StatusInternalServerError,
			Body: map[string]string{"error": err.Error()},
		}
	}
	return &rest.Response{
		Code: http.StatusOK,
		Body: statuses,
	}
}

type GetStatus struct {
	scheduler *scheduler.Scheduler
}

func NewGetStatus(scheduler *scheduler.Scheduler) *GetStatus {
	return &GetStatus{
		scheduler: scheduler,
	}
}

func (c *GetStatus) Handle(r *rest.Request) *rest.Response {
	taskName, exists := r.PathValue("name")
	if !exists {
		return &rest.Response{
			Code: http.StatusBadRequest,
			Body: map[string]string{"error": "task name is required"},
		}
	}

	ctx := r.Context()
	state, err := c.scheduler.GetTaskState(ctx, taskName)
	if err != nil {
		return &rest.Response{
			Code: http.StatusNotFound,
			Body: map[string]string{"error": err.Error()},
		}
	}

	return &rest.Response{
		Code: http.StatusOK,
		Body: state,
	}
}

package http

import (
	"fmt"
	"net/http"

	"github.com/ensoria/projecttemplate/internal/app/scheduler/api/controller/dto"
	"github.com/ensoria/projecttemplate/internal/plamo/vkit"
	"github.com/ensoria/rest/pkg/rest"
	"github.com/ensoria/scheduler/pkg/scheduler"
	"github.com/ensoria/validator/pkg/rule"
)

type PauseTask struct {
	scheduler *scheduler.Scheduler
}

func NewPauseTask(scheduler *scheduler.Scheduler) *PauseTask {
	return &PauseTask{
		scheduler: scheduler,
	}
}

func (c *PauseTask) Handle(r *rest.Request) *rest.Response {
	taskName, exists := r.PathValue("name")
	if !exists {
		return &rest.Response{
			Code: http.StatusBadRequest,
			Body: map[string]string{"error": "task name is required"},
		}
	}

	pt, msgs := vkit.RestRequestBody[dto.PauseTask](
		r,
		&rule.RuleSet{Field: "reason", Rules: []rule.Rule{vkit.Required()}},
	)
	if msgs != nil {
		return &rest.Response{
			Code: http.StatusBadRequest,
			Body: msgs,
		}
	}

	ctx := r.Context()
	if err := c.scheduler.PauseTask(ctx, taskName, pt.Reason); err != nil {
		return &rest.Response{
			Code: http.StatusInternalServerError,
			Body: map[string]string{"error": err.Error()},
		}
	}

	return &rest.Response{
		Code: http.StatusOK,
		Body: map[string]string{"status": fmt.Sprintf("task [%s] paused", taskName)},
	}

}

type ResumeTask struct {
	scheduler *scheduler.Scheduler
}

func NewResumeTask(scheduler *scheduler.Scheduler) *ResumeTask {
	return &ResumeTask{
		scheduler: scheduler,
	}
}

func (c *ResumeTask) Handle(request *rest.Request) *rest.Response {
	// TODO:
	return nil
}

type DisableTask struct {
	scheduler *scheduler.Scheduler
}

func NewDisableTask(scheduler *scheduler.Scheduler) *DisableTask {
	return &DisableTask{
		scheduler: scheduler,
	}
}

func (c *DisableTask) Handle(request *rest.Request) *rest.Response {
	// TODO:
	return nil
}

type EnableTask struct {
	scheduler *scheduler.Scheduler
}

func NewEnableTask(scheduler *scheduler.Scheduler) *EnableTask {
	return &EnableTask{
		scheduler: scheduler,
	}
}

func (c *EnableTask) Handle(request *rest.Request) *rest.Response {
	// TODO:
	return nil
}

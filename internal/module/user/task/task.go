package task

import (
	"context"
	"log"

	"github.com/ensoria/projecttemplate/internal/module/user/service"
	"github.com/ensoria/projecttemplate/internal/scheduler"
	"github.com/ensoria/scheduler/pkg/cron"
)

// WORKING: まだ実験中
func NewUserTask(task *SimpleTask) (*scheduler.ScheduledTask, error) {
	everyMinutes, err := cron.New("*", "*", "*", "*", "*")
	if err != nil {
		return nil, err
	}
	return &scheduler.ScheduledTask{
		Name: "SimpleUserTask",
		Cron: everyMinutes,
		Task: task.Run,
	}, nil
}

type SimpleTask struct {
	Service service.UserService
}

func NewSimpleTask(service service.UserService) *SimpleTask {
	return &SimpleTask{
		Service: service,
	}
}

func (t *SimpleTask) Run(ctx context.Context) error {
	log.Println("SimpleTask is running")
	return nil
}

package scheduler

import (
	"context"
	"log"

	userService "github.com/ensoria/projecttemplate/internal/module/user/service"
	"github.com/ensoria/scheduler/pkg/cron"
	sched "github.com/ensoria/scheduler/pkg/scheduler"
)

func RegisterTasks(s *sched.Scheduler) error {
	everyMinutes, err := cron.New("*", "*", "*", "*", "*")
	if err != nil {
		return err
	}

	if err := s.SetSchedule("SampleTask", everyMinutes, func(ctx context.Context) error {
		log.Println("Sample Task executed")
		return nil
	}); err != nil {
		return err
	}

	// 他のタスクもここで登録していく

	return nil
}

// TODO: これをimplementしたものを、各モジュールで作成して登録する
type SchedulerTask interface {
	sched.Task
}

type SimpleTask struct{}

func NewSimpleTask(service userService.UserService) *SimpleTask {
	return &SimpleTask{}
}

func (t *SimpleTask) Run(ctx context.Context) error {
	log.Println("SimpleTask is running")
	return nil
}

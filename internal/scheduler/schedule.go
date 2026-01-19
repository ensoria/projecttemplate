package scheduler

import (
	"github.com/ensoria/scheduler/pkg/cron"
	sched "github.com/ensoria/scheduler/pkg/scheduler"
)

type ScheduledTask struct {
	Name string
	Cron *cron.Cron
	Task sched.Task
}

// TODO: 各モジュールで、ScheduledTaskを、fxに登録する
func RegisterTasks(s *sched.Scheduler, tasks []ScheduledTask) error {
	for _, task := range tasks {
		if err := s.SetSchedule(task.Name, task.Cron, task.Task); err != nil {
			return err
		}
	}

	return nil
}

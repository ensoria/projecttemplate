package task

import (
	"github.com/ensoria/scheduler/pkg/cron"
	sched "github.com/ensoria/scheduler/pkg/scheduler"
)

type ScheduledTask struct {
	Name string
	Cron *cron.Cron
	Task sched.Task
}

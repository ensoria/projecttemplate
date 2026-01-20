package job

import workerJob "github.com/ensoria/worker/pkg/job"

type JobHandler struct {
	Name    string
	Handler workerJob.JobHandler
	Options *workerJob.Option
}

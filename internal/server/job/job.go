package job

import libWorker "github.com/ensoria/worker/pkg/job" // TODO: 名前を考え直す

type JobHandler struct {
	Name    string
	Handler libWorker.JobHandler
	Options *libWorker.Option
}

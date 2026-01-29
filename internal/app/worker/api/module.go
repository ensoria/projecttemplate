package system

import (
	"github.com/ensoria/config/pkg/appconfig"
	"github.com/ensoria/config/pkg/registry"
	"github.com/ensoria/projecttemplate/internal/app/worker/api/controller/http"
	"github.com/ensoria/projecttemplate/internal/plamo/dikit"
	"github.com/ensoria/rest/pkg/rest"
)

const ModuleName = "default"

func Params() (*appconfig.Parameters, error) {
	return registry.ModuleParams(ModuleName)
}

func NewListJobsModule(get *http.ListJobs) *rest.Module {
	return &rest.Module{
		Path: "/_/jobs",
		Get:  get,
	}
}

func NewJobStatusModule(get *http.JobStatus) *rest.Module {
	return &rest.Module{
		Path: "/_/jobs/{id}/status",
		Get:  get,
	}
}

// TODO: DELETE /_/jobs/{id}

// TODO: GET /_/dead-letter-jobs
// TODO: POST /_/dead-letter-jobs/{id}/retry"
// TODO: POST /_/dead-letter-jobs/retry-by-name
// TODO: POST /_/dead-letter-jobs/retry-all
// TODO: DELETE /_/dead-letter-jobs/{id}

func init() {
	dikit.AppendConstructors([]any{
		http.NewListJobs,
		dikit.AsHTTPModule(NewListJobsModule),

		http.NewJobStatus,
		dikit.AsHTTPModule(NewJobStatusModule),

		// TODO:
	})
}

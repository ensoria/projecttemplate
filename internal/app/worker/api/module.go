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

func NewCancelJobModule(cancel *http.CancelJob) *rest.Module {
	return &rest.Module{
		Path:   "/_/jobs/{id}",
		Delete: cancel,
	}
}

func NewListDeadLetterJobsModule(list *http.ListDeadLetterJobs) *rest.Module {
	return &rest.Module{
		Path: "/_/dead-letter-jobs",
		Get:  list,
	}
}

func NewGetDeadLetterJobModule(get *http.GetDeadLetterJobs) *rest.Module {
	return &rest.Module{
		Path: "/_/dead-letter-jobs/{id}",
		Get:  get,
	}
}

func NewRetryDeadLetterJobModule(retry *http.RetryDeadLetterJob) *rest.Module {
	return &rest.Module{
		Path: "/_/dead-letter-jobs/{id}/retry",
		Post: retry,
	}
}

func NewRetryDeadLetterJobsByNameModule(retryByName *http.RetryDeadLetterJobsByName) *rest.Module {
	return &rest.Module{
		Path: "/_/dead-letter-jobs/retry-by-name",
		Post: retryByName,
	}
}

// TODO: POST /_/dead-letter-jobs/retry-all
// TODO: DELETE /_/dead-letter-jobs/{id}

func init() {
	dikit.AppendConstructors([]any{
		http.NewListJobs,
		dikit.AsHTTPModule(NewListJobsModule),

		http.NewJobStatus,
		dikit.AsHTTPModule(NewJobStatusModule),

		http.NewCancelJob,
		dikit.AsHTTPModule(NewCancelJobModule),

		http.NewListDeadLetterJobs,
		dikit.AsHTTPModule(NewListDeadLetterJobsModule),

		http.NewGetDeadLetterJobs,
		dikit.AsHTTPModule(NewGetDeadLetterJobModule),

		http.NewRetryDeadLetterJob,
		dikit.AsHTTPModule(NewRetryDeadLetterJobModule),

		http.NewRetryDeadLetterJobsByName,
		dikit.AsHTTPModule(NewRetryDeadLetterJobsByNameModule),

		// TODO:
	})
}

package system

import (
	"github.com/ensoria/config/pkg/appconfig"
	"github.com/ensoria/config/pkg/registry"
	"github.com/ensoria/projecttemplate/internal/app/scheduler/api/controller/http"
	"github.com/ensoria/projecttemplate/internal/plamo/dikit"
	"github.com/ensoria/rest/pkg/rest"
)

const ModuleName = "default"

func Params() (*appconfig.Parameters, error) {
	return registry.ModuleParams(ModuleName)
}

// TODO: 全部middlewareで、特定のクライアントからしかアクセスできないようにする

func NewListTaskModule(listTasks *http.ListTasks) *rest.Module {
	return &rest.Module{
		Path: "/_/tasks",
		Get:  listTasks,
	}
}

func NewTaskStateModule(getStatus *http.GetStatus) *rest.Module {
	return &rest.Module{
		Path: "/_/tasks/{name}",
		Get:  getStatus,
	}
}

func NewPauseTaskModule(pauseTask *http.ResumeTask) *rest.Module {
	return &rest.Module{
		Path: "/_/tasks/{name}/pause",
		Post: pauseTask,
	}
}

func NewResumeTaskModule(resumeTask *http.ResumeTask) *rest.Module {
	return &rest.Module{
		Path: "/_/tasks/{name}/resume",
		Post: resumeTask,
	}
}

func NewDisableTaskModule(disableTask *http.DisableTask) *rest.Module {
	return &rest.Module{
		Path: "/_/tasks/{name}/disable",
		Post: disableTask,
	}
}

func NewEnableTaskModule(enableTask *http.EnableTask) *rest.Module {
	return &rest.Module{
		Path: "/_/tasks/{name}/enable",
		Post: enableTask,
	}
}

func init() {
	dikit.AppendConstructors([]any{
		http.NewListTasks,
		dikit.AsHTTPModule(NewListTaskModule),

		http.NewGetStatus,
		dikit.AsHTTPModule(NewTaskStateModule),

		http.NewPauseTask,
		dikit.AsHTTPModule(NewPauseTaskModule),

		http.NewResumeTask,
		dikit.AsHTTPModule(NewResumeTaskModule),

		http.NewDisableTask,
		dikit.AsHTTPModule(NewDisableTaskModule),

		http.NewEnableTask,
		dikit.AsHTTPModule(NewEnableTaskModule),
	})
}

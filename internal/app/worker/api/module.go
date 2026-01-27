package system

import (
	"github.com/ensoria/config/pkg/appconfig"
	"github.com/ensoria/config/pkg/registry"
	"github.com/ensoria/projecttemplate/internal/plamo/dikit"
	"github.com/ensoria/rest/pkg/rest"
)

const ModuleName = "default"

func Params() (*appconfig.Parameters, error) {
	return registry.ModuleParams(ModuleName)
}

// TODO: inject modules
func NewSchedulerModule() *rest.Module {
	return &rest.Module{
		Path: "/_/jobs",
		// Get:  get,
	}
}

func init() {
	dikit.AppendConstructors([]any{
		// TODO:
	})
}

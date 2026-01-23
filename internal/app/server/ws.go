package server

import (
	"github.com/ensoria/projecttemplate/internal/plamo/dikit"
	"github.com/ensoria/websocket/pkg/wsconfig"
	"github.com/ensoria/websocket/pkg/wsrouter"
	"go.uber.org/fx"
)

func CreateWSRouter(modules []*wsconfig.Module) *wsrouter.Router {
	return &wsrouter.Router{
		Modules: modules,
	}
}

func InjectWSModules(f any) any {
	return fx.Annotate(f, fx.ParamTags(dikit.GroupTagWSModules))
}

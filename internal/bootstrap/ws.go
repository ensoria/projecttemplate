package bootstrap

import (
	"github.com/ensoria/websocket/pkg/wsconfig"
	"github.com/ensoria/websocket/pkg/wsrouter"
)

func CreateWSRouter(modules []*wsconfig.Module) *wsrouter.Router {
	return &wsrouter.Router{
		Modules: modules,
	}
}

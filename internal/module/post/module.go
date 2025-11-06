package post

import (
	"github.com/ensoria/config/pkg/appconfig"
	"github.com/ensoria/config/pkg/registry"
	postgrpc "github.com/ensoria/projecttemplate/internal/module/post/controller/grpc"
	"github.com/ensoria/projecttemplate/internal/module/post/controller/http"
	"github.com/ensoria/projecttemplate/internal/module/post/service"
	"github.com/ensoria/projecttemplate/internal/plamo/dikit"
	pb "github.com/ensoria/projecttemplate/service/adapter/post"
	"github.com/ensoria/rest/pkg/rest"
)

const ModuleName = "post"

func Params() (*appconfig.Parameters, error) {
	return registry.ModuleParams(ModuleName)
}

func NewModule(get *http.Get) *rest.Module {
	return &rest.Module{
		Path: "/post",
		Get:  get,
	}
}

func init() {
	dikit.AppendConstructors([]any{
		dikit.As[service.PostService](service.NewPostService),
		http.NewGet,
		dikit.AsHTTPModule(NewModule),
		// gRPC
		dikit.AsGRPCService(postgrpc.NewPostGRPCService),
		dikit.As[pb.PostServer](postgrpc.NewPostGRPCService),
	})
}

package user

import (
	"github.com/ensoria/config/pkg/appconfig"
	"github.com/ensoria/config/pkg/registry"
	usergrpc "github.com/ensoria/projecttemplate/internal/module/user/controller/grpc"
	"github.com/ensoria/projecttemplate/internal/module/user/controller/http"
	"github.com/ensoria/projecttemplate/internal/module/user/controller/ws"
	"github.com/ensoria/projecttemplate/internal/module/user/service"
	"github.com/ensoria/projecttemplate/internal/plamo/dikit"
	"github.com/ensoria/rest/pkg/rest"
	"github.com/ensoria/websocket/pkg/wsconfig"

	"github.com/ensoria/projecttemplate/internal/infra/connection/grpcclt"
	pbPost "github.com/ensoria/projecttemplate/service/adapter/post"
	pb "github.com/ensoria/projecttemplate/service/adapter/user"
)

// TODO: encliでモジュールを作成したら、このファイルに
// 自動的に、NewModuleと、Constructorsを追加する
// さらに、moduler.goにもimportを追加すること

const ModuleName = "user"

// TODO: 便利機能として、この関数も自動的にencliで生成する
func Params() (*appconfig.Parameters, error) {
	return registry.ModuleParams(ModuleName)
}

func NewModule(get *http.Get, post *http.Post) *rest.Module {
	return &rest.Module{
		Path: "/user",
		Get:  get,
		Post: post,
	}
}

func NewWebSocketModule(onOpen *ws.OnOpen, onMessage *ws.OnMessage) *wsconfig.Module {
	module := wsconfig.NewDefaultModule("/ws/" + ModuleName)
	// for logging
	module.AddOnOpenMiddleware(ws.LogOnOpen)
	module.OnOpen = onOpen.OnOpen()

	// for logging
	module.AddOnMessageMiddleware(ws.LogOnMessage)
	module.OnMessage = onMessage.OnMessage()
	return module
}

func init() {
	dikit.AppendConstructors([]any{
		dikit.Bind[service.UserService](service.NewUserService),
		http.NewGet,
		http.NewPost,
		dikit.AsHTTPModule(NewModule),

		// WebSocket
		ws.NewOnOpen,
		ws.NewOnMessage,
		dikit.AsWSModule(NewWebSocketModule),

		// gRPC server
		dikit.AsGRPCService(usergrpc.NewUserGRPCService),
		dikit.Bind[pb.UserServer](usergrpc.NewUserGRPCService),

		// gRPC client
		// 別のgRPCサーバーのクライアントが必要な場合は、コンストラクタを追加
		// このコンストラクタが必要な`grpc.ClientConnInterface`は、`service/connection`で定義する
		// gRPCクライアントのコンストラクタは、`dikit.InjectNamed`を使って、どの
		// gRPCコネクションを使うかを指定すること
		dikit.InjectNamed(pbPost.NewPostClient, grpcclt.PostConnName),
	})
}

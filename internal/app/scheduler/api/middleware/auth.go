package middleware

import (
	"net/http"

	httpApp "github.com/ensoria/projecttemplate/internal/app/http"
	"github.com/ensoria/rest/pkg/rest"
)

// デフォルトでは、schedulerのAPIには誰もアクセスできない
// 各アプリケーションの実装で、このミドルウェアを上書きして、特定のクライアントからのみアクセスできるようにする
func SysAdminOnly(next rest.Handler) rest.Handler {
	return func(r *rest.Request) *rest.Response {
		return &rest.Response{
			Code: http.StatusForbidden,
			Body: httpApp.GlobalError{Message: "access denied"},
		}

		// return next(r)
	}
}

package service

import (
	"context"
	"fmt"
	"time"

	order "github.com/ensoria/projecttemplate/internal/module/order/service"
	pbPost "github.com/ensoria/projecttemplate/service/adapter/post"
)

// サービスを別のモジュールから使う場合は、
// 直接このサービスを呼び出すのではなく、
// 一度ServiceAdapterを通して呼び出すこと
// serviceの返す値は必ずDTOにすること
// modelを返さないように実装すること
// modelはserviceの中で処理でのみ使う。
type UserService interface {
	Something() string
	GetPostContent(postId string) (string, error)
}

// gRPCクライアントが必要な場合は、クライアントの型を指定する
// order serviceについては、すでにorderのモジュールでAsされているので、
// このmoduleの`init`でAsする必要はなく、dikitが自動的に解決してくれる
func NewUserService(postClient pbPost.PostClient, orderService order.OrderService) UserService {
	return &UserServiceImpl{
		postClient:   postClient,
		orderService: orderService,
	}
}

type UserServiceImpl struct {
	postClient   pbPost.PostClient
	orderService order.OrderService
}

func (s *UserServiceImpl) Something() string {
	fmt.Printf("injected orderService: %T\n", s.orderService)
	s.orderService.GetOrder()
	return "hoge"
}

func (s *UserServiceImpl) GetPostContent(postId string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	p, err := s.postClient.GetPost(ctx, &pbPost.GetPostRequest{
		PostId: postId,
	})
	if err != nil {
		return "", err
	}
	return p.Content, nil

}

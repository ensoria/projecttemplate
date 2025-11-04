package http

import (
	"fmt"
	"net/http"

	"github.com/ensoria/projecttemplate/internal/module/user/dto"
	"github.com/ensoria/projecttemplate/internal/module/user/service"
	"github.com/ensoria/rest/pkg/rest"
)

// RESTでのresponse
// JSON or XML

type Get struct {
	Service service.UserService
}

func NewGet(service service.UserService) *Get {
	return &Get{
		Service: service,
	}
}

func (c *Get) Handle(r *rest.Request) *rest.Response {

	fmt.Println("here? 1")
	c.Service.Something() // DEBUG:
	fmt.Println("here? 2")
	return &rest.Response{
		Xml:        true,
		Code:       http.StatusOK,
		AddHeaders: map[string]string{"Server": "net/http"},
		Body:       &dto.GetUser{Id: 1, Name: "hoge"},
	}
}

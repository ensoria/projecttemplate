package mb

import (
	"github.com/ensoria/projecttemplate/internal/module/user/service"
	"github.com/ensoria/projecttemplate/internal/plamo/logkit"
)

type UserSubscriber struct {
	UserService service.UserService
}

func NewUserSubscriber(us service.UserService) *UserSubscriber {
	return &UserSubscriber{
		UserService: us,
	}
}

func (h *UserSubscriber) OnReceive(data []byte, metadata map[string]string) error {
	logkit.Info("📨 Received message",
		"topic", metadata["topic"],
		"partition", metadata["partition"],
		"offset", metadata["offset"],
		"key", metadata["key"],
		"value", string(data),
		"timestamp", metadata["timestamp"])
	return nil
}

package temporal

import (
	"fmt"

	"github.com/brokeboycoding/tempo/internal/config"

	"go.temporal.io/sdk/client"
)

func NewTemporalClient(cfg *config.Config) (client.Client, error) {
	c, err := client.Dial(client.Options{
		HostPort:  fmt.Sprintf("%s:%d", cfg.Temporal.Host, cfg.Temporal.Port),
		Namespace: "default",
	})
	if err != nil {
		return nil, fmt.Errorf("tao temporal client that bai: %w", err)
	}
	return c, nil
}

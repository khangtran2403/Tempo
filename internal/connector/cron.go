package connector

import (
	"context"
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
)

type CronConnector struct{}

func NewCronConnector() *CronConnector {
	return &CronConnector{}
}

func (c *CronConnector) Name() string {
	return "cron"
}

func (c *CronConnector) Execute(ctx context.Context, config map[string]interface{}, prevRes map[string]interface{}) (interface{}, error) {
	return map[string]interface{}{
		"trigger_type":    "cron",
		"scheduled_at":    time.Now(),
		"cron_expression": config["expression"],
	}, nil
}
func (c *CronConnector) ValidateConfig(config map[string]interface{}) error {

	expression, ok := config["expression"].(string)
	if !ok || expression == "" {
		return fmt.Errorf("cron: 'expression' là bất buộc")
	}

	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	_, err := parser.Parse(expression)
	if err != nil {
		return fmt.Errorf("cron: Định dạng không hợp lệ '%s': %w", expression, err)
	}

	return nil
}

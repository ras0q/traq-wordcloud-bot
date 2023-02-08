package cron

import (
	"fmt"

	"github.com/ras0q/traq-wordcloud-bot/pkg/config"
	"github.com/robfig/cron/v3"
)

type Map map[string]func()

func Setup(cm Map) error {
	c := cron.New(
		cron.WithLocation(config.JST),
		cron.WithChain(cron.Recover(cron.DefaultLogger)),
	)

	for spec, f := range cm {
		if _, err := c.AddFunc(spec, f); err != nil {
			return fmt.Errorf("failed to add cron job: %w", err)
		}
	}

	c.Start()

	return nil
}

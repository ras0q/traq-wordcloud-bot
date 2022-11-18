package cron

import (
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
)

type Map map[string]func()

func Setup(cm Map, loc *time.Location) error {
	c := cron.New(
		cron.WithLocation(loc),
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

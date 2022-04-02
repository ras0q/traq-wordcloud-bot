package cron

import (
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
)

func Setup(f func(), loc *time.Location) error {
	c := cron.New(
		cron.WithLocation(loc),
		cron.WithChain(cron.Recover(cron.DefaultLogger)),
	)

	if _, err := c.AddFunc("50 23 * * *", f); err != nil {
		return fmt.Errorf("failed to add cron job: %w", err)
	}

	c.Start()

	return nil
}

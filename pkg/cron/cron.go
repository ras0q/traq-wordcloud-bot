package cron

import (
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
)

func Setup(f func()) error {
	c := cron.New(
		cron.WithLocation(time.FixedZone("Asia/Tokyo", 9*60*60)),
		cron.WithChain(cron.Recover(cron.DefaultLogger)),
	)

	if _, err := c.AddFunc("* * * * *", f); err != nil {
		return fmt.Errorf("failed to add cron job: %w", err)
	}

	c.Start()

	return nil
}

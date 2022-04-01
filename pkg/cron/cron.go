package cron

import (
	"fmt"
	"time"

	"github.com/robfig/cron"
)

func Setup(f func()) error {
	c := cron.NewWithLocation(time.FixedZone("Asia/Tokyo", 9*60*60))
	if err := c.AddFunc("0 50 23 * *", f); err != nil {
		return fmt.Errorf("failed to add cron job: %w", err)
	}

	c.Start()

	return nil
}

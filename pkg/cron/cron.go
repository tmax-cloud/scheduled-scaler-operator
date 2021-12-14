package cron

import (
	"time"

	robfigCron "github.com/robfig/cron"
	"github.com/tmax-cloud/scheduled-scaler-operator/pkg/scaler"
)

type Cron struct {
	timeZone string
	cronImpl *robfigCron.Cron
	scalers  []scaler.Scaler
}

func NewCron(timeZone string) *Cron {
	return &Cron{
		timeZone: timeZone,
		scalers:  make([]scaler.Scaler, 0),
	}
}

func (c *Cron) Push(scaler scaler.Scaler) {
	c.scalers = append(c.scalers, scaler)
}

func (c *Cron) Start() error {
	if err := c.init(); err != nil {
		return err
	}

	c.cronImpl.Start()
	return nil
}

func (c *Cron) init() error {
	if c.cronImpl != nil {
		c.cronImpl.Stop()
	}

	if c.timeZone == "none" {
		c.cronImpl = robfigCron.New()
	} else {
		tz, err := time.LoadLocation(c.timeZone)
		if err != nil {
			return err
		}

		c.cronImpl = robfigCron.NewWithLocation(tz)
	}

	for _, scaler := range c.scalers {
		c.cronImpl.AddJob(scaler.Schedule().Runat, scaler)
	}
	return nil
}

func (c *Cron) Stop() {
	c.cronImpl.Stop()
}

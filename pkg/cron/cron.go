package cron

import (
	"time"

	robfigCron "github.com/robfig/cron"
	"github.com/tmax-cloud/scheduled-scaler-operator/pkg/scaler"
)

type Cron interface {
	Push(scaler.Scaler)
	Start() error
	Stop()
}

type CronImpl struct {
	timeZone     string
	internalCron *robfigCron.Cron
	scalers      []scaler.Scaler
}

func NewCron(timeZone string) Cron {
	return &CronImpl{
		timeZone: timeZone,
		scalers:  make([]scaler.Scaler, 0),
	}
}

func (c *CronImpl) Push(scaler scaler.Scaler) {
	c.scalers = append(c.scalers, scaler)
}

func (c *CronImpl) Start() error {
	if err := c.init(); err != nil {
		return err
	}

	c.internalCron.Start()
	return nil
}

func (c *CronImpl) init() error {
	if c.internalCron != nil {
		c.internalCron.Stop()
	}

	if c.timeZone == "none" {
		c.internalCron = robfigCron.New()
	} else {
		tz, err := time.LoadLocation(c.timeZone)
		if err != nil {
			return err
		}

		c.internalCron = robfigCron.NewWithLocation(tz)
	}

	for _, scaler := range c.scalers {
		c.internalCron.AddJob(scaler.Schedule().Runat, scaler)
	}
	return nil
}

func (c *CronImpl) Stop() {
	c.internalCron.Stop()
}

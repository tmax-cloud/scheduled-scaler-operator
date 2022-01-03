package cron

import (
	"testing"

	"github.com/golang/mock/gomock"
	scscv1 "github.com/tmax-cloud/scheduled-scaler-operator/api/v1"
	"github.com/tmax-cloud/scheduled-scaler-operator/pkg/scaler/fake"
)

func TestCronManager_UpdateCron(t *testing.T) {
	tc := map[string]struct {
		timezone string
	}{
		"default": {
			timezone: "none",
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			// set test case
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			m := fake.NewMockScaler(ctrl)
			done := make(chan bool)
			m.EXPECT().
				Run().Do(func() {
				done <- true
			})
			m.EXPECT().
				Schedule().DoAndReturn(func() scscv1.Schedule {
				return scscv1.Schedule{
					Runat: "* * * * * *",
				}
			})

			testCron := NewCron(c.timezone)

			// do testing function
			testCron.Push(m)
			testCron.Start()
			<-done
			testCron.Stop()
		})
	}
}

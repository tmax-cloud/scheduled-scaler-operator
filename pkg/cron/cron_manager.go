package cron

import (
	"fmt"

	tmaxiov1 "github.com/tmax-cloud/scheduled-scaler-operator/api/v1"
	"github.com/tmax-cloud/scheduled-scaler-operator/pkg/hpamanager"
	"github.com/tmax-cloud/scheduled-scaler-operator/pkg/scaler"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type CronManager struct {
	client.Client
	scheduleCron map[string]*Cron
}

func NewCronManager(cl client.Client) *CronManager {
	return &CronManager{
		Client:       cl,
		scheduleCron: make(map[string]*Cron),
	}
}

func (m *CronManager) UpdateCron(scheduledScaler *tmaxiov1.ScheduledScaler) error {
	key := fmt.Sprintf("%s-%s", scheduledScaler.Namespace, scheduledScaler.Name)
	previousCron, exist := m.scheduleCron[key]
	if exist {
		previousCron.Stop()
	}

	if err := hpamanager.DeleteHpa(m.Client, hpamanager.GetHpaName(scheduledScaler.Name), scheduledScaler.Namespace); err != nil {
		return fmt.Errorf("Couldn't delete previous hpa during update cron by %v", err)
	}

	tz := "none"
	if scheduledScaler.Spec.TimeZone != "" {
		tz = scheduledScaler.Spec.TimeZone
	}
	newCron := NewCron(tz)
	m.scheduleCron[key] = newCron

	for _, schedule := range scheduledScaler.Spec.Schedule {
		scalerImpl, err := scaler.NewScaler(m.Client, scheduledScaler.Name, scheduledScaler.Namespace, scheduledScaler.Spec.Target.Name, schedule)
		if err != nil {
			return err
		}

		newCron.Push(scalerImpl)
	}

	if err := newCron.Start(); err != nil {
		return err
	}

	return nil
}

func (m *CronManager) RemoveCron(namespace, name string) error {
	key := fmt.Sprintf("%s-%s", namespace, name)
	targetCron, ok := m.scheduleCron[key]
	if !ok {
		return nil
	}

	targetCron.Stop()
	delete(m.scheduleCron, key)
	if err := hpamanager.DeleteHpa(m.Client, hpamanager.GetHpaName(name), namespace); err != nil {
		return err
	}

	return nil
}

func (m *CronManager) IsEqual(scheduledScaler *tmaxiov1.ScheduledScaler) bool {
	key := fmt.Sprintf("%s-%s", scheduledScaler.Namespace, scheduledScaler.Name)
	targetCron, ok := m.scheduleCron[key]
	if !ok {
		return false
	}

	if len(targetCron.scalers) != len(scheduledScaler.Spec.Schedule) {
		return false
	}

	scheduleMap := make(map[string]tmaxiov1.Schedule)
	for _, schedule := range scheduledScaler.Spec.Schedule {
		scheduleMap[schedule.Runat] = schedule
	}

	for _, scaler := range targetCron.scalers {
		schedule := scheduleMap[scaler.Schedule().Runat]
		if schedule.Type != scaler.Schedule().Type {
			return false
		}

		if schedule.Type == "fixed" {
			if *schedule.Replicas != *scaler.Schedule().Replicas {
				return false
			}
		} else {
			if *schedule.MinReplicas != *scaler.Schedule().MinReplicas {
				return false
			}

			if *schedule.MaxReplicas != *scaler.Schedule().MaxReplicas {
				return false
			}
		}
	}

	return true
}

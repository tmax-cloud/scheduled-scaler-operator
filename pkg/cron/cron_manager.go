package cron

import (
	"fmt"

	tmaxiov1 "github.com/tmax-cloud/scheduled-scaler-operator/api/v1"
	"github.com/tmax-cloud/scheduled-scaler-operator/internal"
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

	if err := internal.DeleteHpaByOwner(m.Client, scheduledScaler.Name); err != nil {
		return fmt.Errorf("Couldn't delete previous hpa during update cron by %v", err)
	}

	tz := "none"
	if scheduledScaler.Spec.TimeZone != "" {
		tz = scheduledScaler.Spec.TimeZone
	}
	newCron := NewCron(tz)
	m.scheduleCron[key] = newCron
	targetDeploy, err := internal.GetTargetDeployment(m.Client, scheduledScaler.Spec.Target.Name, scheduledScaler.Namespace)
	if err != nil {
		return err
	}

	for _, schedule := range scheduledScaler.Spec.Schedule {
		scalerImpl, err := scaler.NewScaler(scheduledScaler.Name, schedule, targetDeploy)
		if err != nil {
			return err
		}

		newCron.Push(schedule.Runat, scalerImpl.Run)
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
	if err := internal.DeleteHpaByOwner(m.Client, name); err != nil {
		return err
	}

	return nil
}

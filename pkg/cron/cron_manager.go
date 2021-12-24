package cron

import (
	"fmt"

	scscv1 "github.com/tmax-cloud/scheduled-scaler-operator/api/v1"
	"github.com/tmax-cloud/scheduled-scaler-operator/internal/k8s"
	"github.com/tmax-cloud/scheduled-scaler-operator/pkg/apimanager"
	"github.com/tmax-cloud/scheduled-scaler-operator/pkg/scaler"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type CronManager interface {
	UpdateCron(*scscv1.ScheduledScaler) error
	RemoveCron(*scscv1.ScheduledScaler) error
}

type CronManagerImpl struct {
	client.Client
	scheduleCron map[string]Cron
}

func NewCronManager(cl client.Client) CronManager {
	return &CronManagerImpl{
		Client:       cl,
		scheduleCron: make(map[string]Cron),
	}
}

func (m *CronManagerImpl) UpdateCron(scheduledScaler *scscv1.ScheduledScaler) error {
	key := fmt.Sprintf("%s-%s", scheduledScaler.Namespace, scheduledScaler.Name)
	previousCron, exist := m.scheduleCron[key]
	if exist {
		previousCron.Stop()
	}

	if err := k8s.DeleteHpa(m.Client, k8s.GetHpaName(scheduledScaler.Name), scheduledScaler.Namespace); err != nil {
		return fmt.Errorf("Couldn't delete previous hpa during update cron by %v", err)
	}

	tz := "none"
	if scheduledScaler.Spec.TimeZone != "" {
		tz = scheduledScaler.Spec.TimeZone
	}
	newCron := NewCron(tz)
	m.scheduleCron[key] = newCron

	for _, schedule := range scheduledScaler.Spec.Schedule {
		scalerImpl, err := scaler.New(m.Client, scheduledScaler.Name, scheduledScaler.Namespace, scheduledScaler.Spec.Target.Name, schedule)
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

func (m *CronManagerImpl) RemoveCron(scsc *scscv1.ScheduledScaler) error {
	key := apimanager.GetNamespacedName(*scsc)
	targetCron, ok := m.scheduleCron[key]
	if !ok {
		return nil
	}

	targetCron.Stop()
	delete(m.scheduleCron, key)
	if err := k8s.DeleteHpa(m.Client, k8s.GetHpaName(scsc.Name), scsc.Namespace); err != nil {
		return err
	}

	return nil
}

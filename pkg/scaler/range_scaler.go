package scaler

import (
	"github.com/tmax-cloud/scheduled-scaler-operator/pkg/hpamanager"
)

// RangeScaler is ..
type RangeScaler struct {
	ScalerImpl
}

func (s *RangeScaler) Run() {
	logger.Info("RangeScaler start running")
	if err := hpamanager.UpdateHpa(s.cl, &hpamanager.HpaOptions{
		Namespace:           s.namespace,
		Target:              s.target,
		ScheduledScalerName: s.scheduledScaler,
		MinReplicas:         s.schedule.MinReplicas,
		MaxReplicas:         s.schedule.MaxReplicas,
	}); err != nil {
		logger.Error(err, "Creating Hpa failed in Range scaler")
		return
	}

	logger.Info("scaling done")
}

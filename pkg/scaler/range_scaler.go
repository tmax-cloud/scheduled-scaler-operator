package scaler

import (
	"github.com/tmax-cloud/scheduled-scaler-operator/internal/k8s"
)

// RangeScaler is ..
type RangeScaler struct {
	ScalerImpl
}

func (s *RangeScaler) Run() {
	logger.Info("RangeScaler start running")
	targetDeploy, err := k8s.GetTargetDeployment(s.cl, s.target, s.namespace)
	if err != nil {
		logger.Error(err, "Getting deployment error in RangeScaler")
		return
	}

	if err = k8s.ScaleDeploymentReplicas(s.cl, targetDeploy, s.schedule.MinReplicas); err != nil {
		logger.Error(err, "Patching deployment error in RangeScaler")
		return
	}

	if err := k8s.UpdateHpa(s.cl, &k8s.HpaValidationOptions{
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

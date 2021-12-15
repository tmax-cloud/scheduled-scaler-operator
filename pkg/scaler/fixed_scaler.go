package scaler

import (
	"context"

	"github.com/tmax-cloud/scheduled-scaler-operator/internal/k8s"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// FixedScaler is ...
type FixedScaler struct {
	ScalerImpl
}

func (s *FixedScaler) Run() {
	logger.Info("FixedScaler start running")
	if err := k8s.DeleteHpa(s.cl, k8s.GetHpaName(s.scheduledScaler), s.namespace); err != nil {
		logger.Error(err, "Cleaning HPA failed in FixedScaler")
		return
	}
	replicas := s.schedule.DeepCopy().Replicas
	targetDeploy, err := k8s.GetTargetDeployment(s.cl, s.target, s.namespace)
	if err != nil {
		logger.Error(err, "Getting deployment error in FixedScaler")
		return
	}
	patch := targetDeploy.DeepCopy()
	patch.Spec.Replicas = replicas
	if err = s.cl.Patch(context.Background(), patch, client.MergeFrom(targetDeploy)); err != nil {
		logger.Error(err, "Patching deployment error in FixedScaler")
		return
	}

	logger.Info("scaling done")
}

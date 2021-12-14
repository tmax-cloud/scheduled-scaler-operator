package scaler

import (
	"context"

	"github.com/tmax-cloud/scheduled-scaler-operator/internal"
	"github.com/tmax-cloud/scheduled-scaler-operator/pkg/hpamanager"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// FixedScaler is ...
type FixedScaler struct {
	ScalerImpl
}

func (s *FixedScaler) Run() {
	logger.Info("FixedScaler start running")
	if err := hpamanager.DeleteHpa(s.cl, hpamanager.GetHpaName(s.scheduledScaler), s.namespace); err != nil {
		logger.Error(err, "Cleaning HPA failed in FixedScaler")
		return
	}
	replicas := s.schedule.DeepCopy().Replicas
	targetDeploy, err := internal.GetTargetDeployment(s.cl, s.target, s.namespace)
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

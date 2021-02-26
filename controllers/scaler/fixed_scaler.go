package scaler

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// FixedScaler is ...
type FixedScaler struct {
	scaler ScalerImpl
}

func (s *FixedScaler) Run() {
	logger.Info("FixedScaler start running")
	targetDeploy := s.scaler.target.DeepCopy()
	targetDeploy.Spec.Replicas = s.scaler.schedule.Replicas
	if err := s.scaler.cl.Patch(context.Background(), targetDeploy, client.MergeFrom(s.scaler.target)); err != nil {
		logger.Error(err, "Patching deployment error in FixedScaler")
		return
	}

	logger.Info("scaling done")
}

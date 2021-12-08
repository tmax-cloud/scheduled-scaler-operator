package scaler

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// FixedScaler is ...
type FixedScaler struct {
	ScalerImpl
}

func (s *FixedScaler) Run() {
	logger.Info("FixedScaler start running")
	targetDeploy := s.target.DeepCopy()
	targetDeploy.Spec.Replicas = s.schedule.Replicas
	if err := s.cl.Patch(context.Background(), targetDeploy, client.MergeFrom(s.target)); err != nil {
		logger.Error(err, "Patching deployment error in FixedScaler")
		return
	}

	logger.Info("scaling done")
}

package scaler

import (
	scscv1 "github.com/tmax-cloud/scheduled-scaler-operator/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var logger = logf.Log.WithName("scaler")

// Scaler is ...
type Scaler interface {
	Schedule() scscv1.Schedule
	Run()
}

type ScalerImpl struct {
	scheduledScaler string
	target          string
	namespace       string
	schedule        scscv1.Schedule
	cl              client.Client
}

func (s *ScalerImpl) Schedule() scscv1.Schedule {
	return s.schedule
}

func NewScaler(cl client.Client, name, namespace, targetDeploy string, schedule scscv1.Schedule) (Scaler, error) {
	var scaler Scaler
	scalerImpl := ScalerImpl{
		scheduledScaler: name,
		target:          targetDeploy,
		namespace:       namespace,
		schedule:        schedule,
		cl:              cl,
	}

	switch schedule.Type {
	case "fixed":
		scaler = &FixedScaler{
			scalerImpl,
		}
	case "range":
		scaler = &RangeScaler{
			scalerImpl,
		}
	}

	return scaler, nil
}

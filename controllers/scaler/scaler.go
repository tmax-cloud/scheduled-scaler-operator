package scaler

import (
	tmaxiov1 "github.com/tmax-cloud/scheduled-scaler-operator/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var logger = logf.Log.WithName("scaler")

// Scaler is ...
type Scaler interface {
	Run()
}

type ScalerImpl struct {
	scheduledScaler string
	target          *appsv1.Deployment
	schedule        tmaxiov1.Schedule
	cl              client.Client
}

func NewScaler(name string, schedule tmaxiov1.Schedule, targetDeploy *appsv1.Deployment) (Scaler, error) {
	var scaler Scaler
	cl, err := client.New(config.GetConfigOrDie(), client.Options{})
	if err != nil {
		return nil, err
	}

	scalerImpl := ScalerImpl{
		scheduledScaler: name,
		target:          targetDeploy,
		schedule:        schedule,
		cl:              cl,
	}

	switch schedule.Type {
	case "fixed":
		scaler = &FixedScaler{
			scaler: scalerImpl,
		}
	case "range":
		scaler = &RangeScaler{
			scaler: scalerImpl,
		}
	}

	return scaler, nil
}

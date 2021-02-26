package scaler

import (
	"context"
	"fmt"
	"math/rand"

	"k8s.io/api/autoscaling/v2beta2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RangeScaler is ..
type RangeScaler struct {
	scaler ScalerImpl
}

func (s *RangeScaler) Run() {
	logger.Info("RangeScaler start running")
	utilization := int32(50)
	b := make([]byte, 6)
	rand.Read(b)

	hpaName := fmt.Sprintf("%s-%s-%x", s.scaler.scheduledScaler, s.scaler.target.Name, b)
	hpa := &v2beta2.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      hpaName,
			Namespace: s.scaler.target.Namespace,
			Labels: map[string]string{
				"owner": s.scaler.scheduledScaler,
			},
		},
		Spec: v2beta2.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: v2beta2.CrossVersionObjectReference{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
				Name:       s.scaler.target.Name,
			},
			MinReplicas: s.scaler.schedule.MinReplicas,
			MaxReplicas: s.scaler.schedule.MaxReplicas,
			Metrics: []v2beta2.MetricSpec{
				{
					Type: v2beta2.ResourceMetricSourceType,
					Resource: &v2beta2.ResourceMetricSource{
						Name: v1.ResourceCPU,
						Target: v2beta2.MetricTarget{
							Type:               v2beta2.UtilizationMetricType,
							AverageUtilization: &utilization,
						},
					},
				},
			},
		},
	}

	if err := s.scaler.cl.Create(context.Background(), hpa); err != nil {
		logger.Error(err, "Failed to create hpa in range scaler")
		return
	}

	logger.Info("scaling done")
}

package hpamanager

import (
	"context"
	"fmt"

	autov2beta2 "k8s.io/api/autoscaling/v2beta2"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type HpaOptions struct {
	Namespace           string
	Target              string
	ScheduledScalerName string
	MinReplicas         *int32
	MaxReplicas         *int32
}

func (o *HpaOptions) validate() bool {
	if o.Namespace == "" ||
		o.Target == "" ||
		o.ScheduledScalerName == "" ||
		o.MinReplicas == nil ||
		o.MaxReplicas == nil {
		return false
	}

	return true
}

func GetHpaName(scheduledScalerName string) string {
	return fmt.Sprintf("%s-hpa", scheduledScalerName)
}

func GetHpa(cl client.Client, name, namespace string, hpa *autov2beta2.HorizontalPodAutoscaler) (bool, error) {
	if err := cl.Get(context.Background(), types.NamespacedName{Name: name, Namespace: namespace}, hpa); err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		} else {
			return false, fmt.Errorf("Getting HPA failed")
		}
	}
	return true, nil
}

func UpdateHpa(cl client.Client, options *HpaOptions) error {
	if !options.validate() {
		return fmt.Errorf("Required options validation failed in CreateHpa")
	}

	hpaName := GetHpaName(options.ScheduledScalerName)
	hpa := &autov2beta2.HorizontalPodAutoscaler{}
	if ok, err := GetHpa(cl, hpaName, options.Namespace, hpa); err != nil {
		return fmt.Errorf("Getting HPA failed in UpdateHPA")
	} else if ok {
		newHpa := hpa.DeepCopy()
		newHpa.Spec.MinReplicas = options.MinReplicas
		newHpa.Spec.MaxReplicas = *options.MaxReplicas
		if err = cl.Patch(context.TODO(), newHpa, client.MergeFrom(hpa)); err != nil {
			return fmt.Errorf("Patch Hpa failed: %v", err)
		}
	} else {
		utilization := int32(50)
		newHpa := &autov2beta2.HorizontalPodAutoscaler{
			ObjectMeta: metav1.ObjectMeta{
				Name:      hpaName,
				Namespace: options.Namespace,
				Labels: map[string]string{
					"owner": options.ScheduledScalerName,
				},
			},
			Spec: autov2beta2.HorizontalPodAutoscalerSpec{
				ScaleTargetRef: autov2beta2.CrossVersionObjectReference{
					APIVersion: "apps/v1",
					Kind:       "Deployment",
					Name:       options.Target,
				},
				MinReplicas: options.MinReplicas,
				MaxReplicas: *options.MaxReplicas,
				Metrics: []autov2beta2.MetricSpec{
					{
						Type: autov2beta2.ResourceMetricSourceType,
						Resource: &autov2beta2.ResourceMetricSource{
							Name: v1.ResourceCPU,
							Target: autov2beta2.MetricTarget{
								Type:               autov2beta2.UtilizationMetricType,
								AverageUtilization: &utilization,
							},
						},
					},
				},
			},
		}
		if err := cl.Create(context.Background(), newHpa); err != nil {
			return fmt.Errorf("Creating Hpa failed: %v", err)
		}
	}

	return nil
}

func DeleteHpa(cl client.Client, name, namespace string) error {
	hpa := &autov2beta2.HorizontalPodAutoscaler{}
	if ok, err := GetHpa(cl, name, namespace, hpa); err != nil {
		return fmt.Errorf("Getting Hpa failed in DeleteHpa")
	} else if ok {
		if err = cl.Delete(context.Background(), hpa); err != nil {
			return fmt.Errorf("Delete Hpa failed by: %v", err)
		}
	}

	return nil
}

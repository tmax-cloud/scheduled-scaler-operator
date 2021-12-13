package internal

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	autoscaling "k8s.io/api/autoscaling/v2beta2"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetTargetDeployment(cl client.Client, name, namespace string) (*appsv1.Deployment, error) {
	targetDeploy := &appsv1.Deployment{}
	if err := cl.Get(context.Background(), types.NamespacedName{Name: name, Namespace: namespace}, targetDeploy); err != nil {
		return nil, err
	}

	return targetDeploy, nil
}

func DeleteHpaByOwner(cl client.Client, owner string) error {
	hpaList := &autoscaling.HorizontalPodAutoscalerList{}
	selector := labels.SelectorFromSet(map[string]string{
		"owner": owner,
	})

	if err := cl.List(context.Background(), hpaList, &client.ListOptions{LabelSelector: selector}); err != nil {
		return fmt.Errorf("couldn't list hpas: %v", err)
	}

	for _, hpa := range hpaList.Items {
		if err := cl.Delete(context.Background(), &hpa); err != nil {
			return fmt.Errorf("couldn't delete hpa named %s: %v", hpa.Name, err)
		}
	}

	return nil
}

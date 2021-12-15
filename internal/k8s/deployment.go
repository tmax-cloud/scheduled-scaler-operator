package k8s

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
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

package controllers

import (
	"context"
	"fmt"

	tmaxiov1 "github.com/tmax-cloud/scheduled-scaler-operator/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func UpdateStatus(cl client.Client, origin *tmaxiov1.ScheduledScaler, status tmaxiov1.Status, msg string, reason string) error {
	fmt.Println("Update status runned")
	originObject := client.MergeFrom(origin)
	patch := origin.DeepCopy()
	patch.Status = tmaxiov1.ScheduledScalerStatus{
		Phase:   string(status),
		Message: msg,
		Reason:  reason,
	}

	if err := cl.Status().Patch(context.TODO(), patch, originObject); err != nil {
		return fmt.Errorf("Couldn't update status: %v", err)
	}

	return nil
}

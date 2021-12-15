package apimanager

import (
	"context"
	"fmt"

	scscv1 "github.com/tmax-cloud/scheduled-scaler-operator/api/v1"
	"github.com/tmax-cloud/scheduled-scaler-operator/pkg/validator"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

/*
* API Manager is a set helper functions for API(Custom Resource)
 */

func GetNamespacedName(scsc scscv1.ScheduledScaler) string {
	if scsc.Name == "" || scsc.Namespace == "" {
		return ""
	}

	return fmt.Sprintf("%s-%s", scsc.Namespace, scsc.Name)
}

func UpdateStatus(cl client.Client, scsc *scscv1.ScheduledScaler, status scscv1.ScheduledScalerStatus) error {
	origin := client.MergeFrom(scsc)
	patch := scsc.DeepCopy()
	patch.Status = status

	if err := cl.Status().Patch(context.TODO(), patch, origin); err != nil {
		return fmt.Errorf("Couldn't update status: %v", err)
	}

	return nil
}

func Validate(scsc *scscv1.ScheduledScaler) bool {
	return validator.NewValidator(*scsc).Validate()
}

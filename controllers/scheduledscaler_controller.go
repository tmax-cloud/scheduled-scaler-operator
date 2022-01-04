/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	scscv1 "github.com/tmax-cloud/scheduled-scaler-operator/api/v1"
	"github.com/tmax-cloud/scheduled-scaler-operator/internal/util"
	"github.com/tmax-cloud/scheduled-scaler-operator/pkg/apimanager"
	"github.com/tmax-cloud/scheduled-scaler-operator/pkg/cache"
	"github.com/tmax-cloud/scheduled-scaler-operator/pkg/cron"
)

const finalizer = "finalizer.scheduledscaler.tmax.io"

// ScheduledScalerReconciler reconciles a ScheduledScaler object
type ScheduledScalerReconciler struct {
	client.Client
	Log         logr.Logger
	Scheme      *runtime.Scheme
	cache       cache.ScheduledScalerCache
	cronManager cron.CronManager
}

// +kubebuilder:rbac:groups=tmax.io,resources=scheduledscalers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=tmax.io,resources=scheduledscalers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=autoscaling,resources=horizontalpodautoscalers,verbs=get;list;watch;create;update;patch;delete
func (r *ScheduledScalerReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("scheduledscaler", req.NamespacedName)

	// get scheduled scaler resource
	scheduledScaler := &scscv1.ScheduledScaler{}
	if err := r.Get(ctx, req.NamespacedName, scheduledScaler); err != nil {
		if errors.IsNotFound(err) {
			// Not-found error isn't handled by error, because it is always occured in deleting phase
			log.Info(fmt.Sprintf("Couldn't find %s ScheduledScaler", req.NamespacedName))
			return ctrl.Result{}, nil
		}
		log.Error(err, "Unable to fetch resource ScheduledScaler")
		return ctrl.Result{}, err
	}

	// set deleting flag to default: false
	isDeleting := false
	// deferring to manage cache
	defer func() {
		if isDeleting {
			r.cache.Remove(scheduledScaler)
			return
		} else if apimanager.GetNamespacedName(*scheduledScaler) == "" {
			return
		}
		r.cache.Put(scheduledScaler)
	}()

	// handle finalizer to check deleting event
	if scheduledScaler.ObjectMeta.DeletionTimestamp.IsZero() {
		if !util.ContainsString(scheduledScaler.ObjectMeta.Finalizers, finalizer) {
			// add finalizer if not set to remove cron after deleting CR
			scheduledScaler.ObjectMeta.Finalizers = append(scheduledScaler.ObjectMeta.Finalizers, finalizer)
			if err := r.Update(ctx, scheduledScaler); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		if util.ContainsString(scheduledScaler.ObjectMeta.Finalizers, finalizer) {
			isDeleting = true // set deleting flag to remove scsc from cache
			log.Info("deleting CR")
			r.cronManager.RemoveCron(scheduledScaler) // remove cron of scsc
			scheduledScaler.ObjectMeta.Finalizers = util.RemoveString(scheduledScaler.ObjectMeta.Finalizers, finalizer)
			if err := r.Update(ctx, scheduledScaler); err != nil {
				return ctrl.Result{}, err
			}

			return ctrl.Result{}, nil
		}
	}

	// When reconciled scsc is failed status and has reason InvalidSpecError, validate it again to check if it is modified
	if scheduledScaler.Status.Phase == scscv1.StatusFailed && scheduledScaler.Status.Reason == scscv1.ValidationFailedError {
		if !apimanager.Validate(scheduledScaler) {
			return ctrl.Result{}, nil
		}
	}

	// When scsc has no status(creating) or failed status, update status to updating status to reconcile again
	if scheduledScaler.Status.Phase == "" || scheduledScaler.Status.Phase == scscv1.StatusFailed {
		if err := apimanager.UpdateStatus(
			r.Client,
			scheduledScaler,
			scscv1.ScheduledScalerStatus{
				Phase:   scscv1.StatusUpdating,
				Message: "Scheduled Scaler is running",
				Reason:  scscv1.NeedToReconcile,
			}); err != nil {
			log.Error(err, "Updating status failed")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, nil
	}

	// When scsc has updating status, do reconciling logic: validate scsc and update cron
	if scheduledScaler.Status.Phase == scscv1.StatusUpdating {
		if !apimanager.Validate(scheduledScaler) {
			r.cronManager.RemoveCron(scheduledScaler)
			if err := apimanager.UpdateStatus(r.Client, scheduledScaler, scscv1.ScheduledScalerStatus{
				Phase:   scscv1.StatusFailed,
				Message: "Scheduled Scaler spec is invalid",
				Reason:  scscv1.ValidationFailedError,
			}); err != nil {
				log.Error(err, "Updating status failed")
				return ctrl.Result{}, nil
			}
			log.Error(fmt.Errorf("Invalid Spec is entered"), "Invalid Spec is entered")
			return ctrl.Result{}, nil
		}

		if err := r.cronManager.UpdateCron(scheduledScaler); err != nil {
			log.Error(err, "Couldn't update cron")
			if err = apimanager.UpdateStatus(r.Client, scheduledScaler, scscv1.ScheduledScalerStatus{
				Phase:   scscv1.StatusFailed,
				Message: "Scheduled Scaler is failed",
				Reason:  scscv1.InternalLogicError,
			}); err != nil {
				log.Error(err, "Updating status failed")
				return ctrl.Result{}, nil
			}
			return ctrl.Result{}, err
		}

		log.Info("Reconciling done")
		if err := apimanager.UpdateStatus(r.Client, scheduledScaler, scscv1.ScheduledScalerStatus{
			Phase:   scscv1.StatusRunning,
			Message: "Scheduled Scaler is running",
			Reason:  scscv1.ReconcileDone,
		}); err != nil {
			log.Error(err, "Updating status failed")
			return ctrl.Result{}, nil
		}
	} else {
		// In else case (Running status), check if scsc is modified. If it's modified, update status to Updating to reconcile again
		if r.cache.HasChanged(scheduledScaler) {
			if err := apimanager.UpdateStatus(r.Client, scheduledScaler, scscv1.ScheduledScalerStatus{
				Phase:   scscv1.StatusUpdating,
				Message: "Scheduled Scaler is running",
				Reason:  scscv1.NeedToReconcile,
			}); err != nil {
				log.Error(err, "Updating status failed")
				return ctrl.Result{}, nil
			}
			return ctrl.Result{}, nil
		}
	}

	return ctrl.Result{}, nil
}

// Init is for initiating member components: cron manager and cache
func (r *ScheduledScalerReconciler) Init() *ScheduledScalerReconciler {
	r.cronManager = cron.NewCronManager(r.Client)
	r.cache = cache.New()
	return r
}

func (r *ScheduledScalerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&scscv1.ScheduledScaler{}).
		Complete(r)
}

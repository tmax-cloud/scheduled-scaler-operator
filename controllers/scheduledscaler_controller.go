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
	appsv1 "k8s.io/api/apps/v1"
	autoscaling "k8s.io/api/autoscaling/v2beta2"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	tmaxiov1 "github.com/tmax-cloud/scheduled-scaler-operator/api/v1"
	"github.com/tmax-cloud/scheduled-scaler-operator/internal/util"
	"github.com/tmax-cloud/scheduled-scaler-operator/pkg/cron"
	"github.com/tmax-cloud/scheduled-scaler-operator/pkg/scaler"
)

// ScheduledScalerReconciler reconciles a ScheduledScaler object
type ScheduledScalerReconciler struct {
	client.Client
	Log          logr.Logger
	Scheme       *runtime.Scheme
	scheduleCron map[string]*cron.Cron
}

// +kubebuilder:rbac:groups=tmax.io,resources=scheduledscalers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=tmax.io,resources=scheduledscalers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=autoscaling,resources=horizontalpodautoscalers,verbs=get;list;watch;create;update;patch;delete

func (r *ScheduledScalerReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("scheduledscaler", req.NamespacedName)
	namespacedNameString := fmt.Sprintf("%s-%s", req.Namespace, req.Name)

	// get scheduled scaler resource
	scheduledScaler := &tmaxiov1.ScheduledScaler{}
	if err := r.Get(ctx, req.NamespacedName, scheduledScaler); err != nil {
		if errors.IsNotFound(err) {
			log.Error(err, fmt.Sprintf("Couldn't find %s ScheduledScaler", req.NamespacedName))
			return ctrl.Result{}, err
		}
		log.Error(err, "Unable to fetch resource ScheduledScaler")
		return ctrl.Result{}, err
	}

	myFinalizerName := "finalizer.scheduledscaler.tmax.io"
	if scheduledScaler.ObjectMeta.DeletionTimestamp.IsZero() {
		if !util.ContainsString(scheduledScaler.ObjectMeta.Finalizers, myFinalizerName) {
			// add finalizer to remove cron after deleting CR
			scheduledScaler.ObjectMeta.Finalizers = append(scheduledScaler.ObjectMeta.Finalizers, myFinalizerName)
			if err := r.Update(ctx, scheduledScaler); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		if util.ContainsString(scheduledScaler.ObjectMeta.Finalizers, myFinalizerName) {
			log.Info("deleting CR")
			targetCron, ok := r.scheduleCron[namespacedNameString]
			if ok {
				log.Info("stop cron")
				targetCron.Stop()
				delete(r.scheduleCron, namespacedNameString)
				if err := r.deleteHpa(scheduledScaler); err != nil {
					log.Error(err, "Error occured during deleting hpa")
				}
			}

			scheduledScaler.ObjectMeta.Finalizers = util.RemoveString(scheduledScaler.ObjectMeta.Finalizers, myFinalizerName)
			if err := r.Update(ctx, scheduledScaler); err != nil {
				return ctrl.Result{}, err
			}

			return ctrl.Result{}, nil
		}
	}

	if scheduledScaler.Status.Phase == string(tmaxiov1.StatusFailed) {
		return ctrl.Result{}, nil
	}

	if scheduledScaler.Status.Phase == "" {
		r.updateStatus(r.Client, scheduledScaler, tmaxiov1.StatusCreating, "Scheduled Scaler is creating", "InitializingProcess")
		return ctrl.Result{}, nil
	}

	if err := r.updateCron(scheduledScaler); err != nil {
		log.Error(err, "Couldn't update cron")
		r.updateStatus(r.Client, scheduledScaler, tmaxiov1.StatusFailed, "Scheduled Scaler is failed", "InternalLogicError")
		return ctrl.Result{}, err
	}

	r.updateStatus(r.Client, scheduledScaler, tmaxiov1.StatusRunning, "Scheduled Scaler is running", "Running")
	log.Info("Reconciling done")

	return ctrl.Result{}, nil
}

func (r *ScheduledScalerReconciler) Init() *ScheduledScalerReconciler {
	r.scheduleCron = make(map[string]*cron.Cron)
	return r
}

func (r *ScheduledScalerReconciler) updateCron(scheduledScaler *tmaxiov1.ScheduledScaler) error {
	key := fmt.Sprintf("%s-%s", scheduledScaler.Namespace, scheduledScaler.Name)
	previousCron, exist := r.scheduleCron[key]
	if exist {
		previousCron.Stop()
	}

	if err := r.deleteHpa(scheduledScaler); err != nil {
		return fmt.Errorf("Couldn't delete previous hpa during update cron by %v", err)
	}

	tz := "none"
	if scheduledScaler.Spec.TimeZone != "" {
		tz = scheduledScaler.Spec.TimeZone
	}
	newCron := cron.NewCron(tz)
	r.scheduleCron[key] = newCron
	targetDeploy, err := getTargetDeployment(scheduledScaler.Spec.Target.Name, scheduledScaler.Namespace)
	if err != nil {
		return err
	}

	for _, schedule := range scheduledScaler.Spec.Schedule {
		scalerImpl, err := scaler.NewScaler(scheduledScaler.Name, schedule, targetDeploy)
		if err != nil {
			return err
		}

		newCron.Push(schedule.Runat, scalerImpl.Run)
	}

	if err := newCron.Start(); err != nil {
		return err
	}

	return nil
}

func (r *ScheduledScalerReconciler) updateStatus(cl client.Client, origin *tmaxiov1.ScheduledScaler, status tmaxiov1.Status, msg string, reason string) error {
	originObject := client.MergeFrom(origin)
	patch := origin.DeepCopy()
	patch.Status = tmaxiov1.ScheduledScalerStatus{
		Phase:   string(status),
		Message: msg,
		Reason:  reason,
	}

	if err := r.Client.Status().Patch(context.TODO(), patch, originObject); err != nil {
		return fmt.Errorf("Couldn't update status: %v", err)
	}

	return nil
}

func (r *ScheduledScalerReconciler) deleteHpa(scheduledScaler *tmaxiov1.ScheduledScaler) error {
	hpaList := &autoscaling.HorizontalPodAutoscalerList{}
	selector := labels.SelectorFromSet(map[string]string{
		"owner": scheduledScaler.Name,
	})

	if err := r.Client.List(context.Background(), hpaList, &client.ListOptions{LabelSelector: selector}); err != nil {
		return fmt.Errorf("couldn't list hpas: %v", err)
	}

	for _, hpa := range hpaList.Items {
		if err := r.Client.Delete(context.Background(), &hpa); err != nil {
			return fmt.Errorf("couldn't delete hpa named %s: %v", hpa.Name, err)
		}
	}

	return nil
}

func getTargetDeployment(name, namespace string) (*appsv1.Deployment, error) {
	cl, err := client.New(config.GetConfigOrDie(), client.Options{})
	if err != nil {
		return nil, err
	}

	targetDeploy := &appsv1.Deployment{}
	if err := cl.Get(context.Background(), types.NamespacedName{Name: name, Namespace: namespace}, targetDeploy); err != nil {
		return nil, err
	}

	return targetDeploy, nil
}

func (r *ScheduledScalerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&tmaxiov1.ScheduledScaler{}).
		Complete(r)
}

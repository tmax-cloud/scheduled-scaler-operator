package controllers

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	scscv1 "github.com/tmax-cloud/scheduled-scaler-operator/api/v1"
	"github.com/tmax-cloud/scheduled-scaler-operator/internal/test"
	"github.com/tmax-cloud/scheduled-scaler-operator/pkg/cache"
	cronFake "github.com/tmax-cloud/scheduled-scaler-operator/pkg/cron/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	cRuntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestScheduledScalerController_Reconcile(t *testing.T) {
	s := runtime.NewScheme()
	utilruntime.Must(scscv1.AddToScheme(s))

	scaledReplica := int32(2)
	nowTime := metav1.Now()

	tc := map[string]struct {
		scsc *scscv1.ScheduledScaler

		expectedFinalizer []string
		expectedStatus    scscv1.ScheduledScalerStatus
		isCronUpdated     bool
		isCronRemoved     bool
		cronUpdateFailed  bool
		inCache           *scscv1.ScheduledScaler
	}{
		"scheduled scaler first created": {
			scsc: &scscv1.ScheduledScaler{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-scsc",
					Namespace: "test-ns",
				},
				Spec: scscv1.ScheduledScalerSpec{
					Target: scscv1.SchedulingTarget{
						Name: "test-deploy",
					},
					Schedule: []scscv1.Schedule{
						{
							Type:     "fixed",
							Runat:    "* * * * *",
							Replicas: &scaledReplica,
						},
					},
				},
			},
			expectedFinalizer: []string{finalizer},
			expectedStatus: scscv1.ScheduledScalerStatus{
				Phase:   scscv1.StatusUpdating,
				Message: "Scheduled Scaler is running",
				Reason:  scscv1.NeedToReconcile,
			},
			isCronUpdated: false,
			isCronRemoved: false,
		},
		"scheduled scaler in updating status and done well": {
			scsc: &scscv1.ScheduledScaler{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "test-scsc",
					Namespace:  "test-ns",
					Finalizers: []string{finalizer},
				},
				Spec: scscv1.ScheduledScalerSpec{
					Target: scscv1.SchedulingTarget{
						Name: "test-deploy",
					},
					Schedule: []scscv1.Schedule{
						{
							Type:     "fixed",
							Runat:    "* * * * *",
							Replicas: &scaledReplica,
						},
					},
				},
				Status: scscv1.ScheduledScalerStatus{
					Phase:   scscv1.StatusUpdating,
					Message: "Scheduled Scaler is running",
					Reason:  scscv1.NeedToReconcile,
				},
			},
			expectedFinalizer: []string{finalizer},
			expectedStatus: scscv1.ScheduledScalerStatus{
				Phase:   scscv1.StatusRunning,
				Message: "Scheduled Scaler is running",
				Reason:  scscv1.ReconcileDone,
			},
			isCronUpdated: true,
			isCronRemoved: false,
		},
		"scheduled scaler in updating status and validation failed": {
			scsc: &scscv1.ScheduledScaler{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "test-scsc",
					Namespace:  "test-ns",
					Finalizers: []string{finalizer},
				},
				Spec: scscv1.ScheduledScalerSpec{
					Target: scscv1.SchedulingTarget{
						Name: "test-deploy",
					},
					Schedule: []scscv1.Schedule{
						{
							Type:  "fixed",
							Runat: "* * * * *",
							// replica missing => invalid
						},
					},
				},
				Status: scscv1.ScheduledScalerStatus{
					Phase:   scscv1.StatusUpdating,
					Message: "Scheduled Scaler is running",
					Reason:  scscv1.NeedToReconcile,
				},
			},
			expectedFinalizer: []string{finalizer},
			expectedStatus: scscv1.ScheduledScalerStatus{
				Phase:   scscv1.StatusFailed,
				Message: "Scheduled Scaler spec is invalid",
				Reason:  scscv1.ValidationFailedError,
			},
			isCronUpdated: false,
			isCronRemoved: true,
		},
		"scheduled scaler in updating status and cron updating failed": {
			scsc: &scscv1.ScheduledScaler{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "test-scsc",
					Namespace:  "test-ns",
					Finalizers: []string{finalizer},
				},
				Spec: scscv1.ScheduledScalerSpec{
					Target: scscv1.SchedulingTarget{
						Name: "test-deploy",
					},
					Schedule: []scscv1.Schedule{
						{
							Type:     "fixed",
							Runat:    "* * * * *",
							Replicas: &scaledReplica,
						},
					},
				},
				Status: scscv1.ScheduledScalerStatus{
					Phase:   scscv1.StatusUpdating,
					Message: "Scheduled Scaler is running",
					Reason:  scscv1.NeedToReconcile,
				},
			},
			expectedFinalizer: []string{finalizer},
			expectedStatus: scscv1.ScheduledScalerStatus{
				Phase:   scscv1.StatusFailed,
				Message: "Scheduled Scaler is failed",
				Reason:  scscv1.InternalLogicError,
			},
			isCronUpdated:    false,
			isCronRemoved:    false,
			cronUpdateFailed: true,
		},
		"scheduled scaler in failed status": {
			scsc: &scscv1.ScheduledScaler{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "test-scsc",
					Namespace:  "test-ns",
					Finalizers: []string{finalizer},
				},
				Spec: scscv1.ScheduledScalerSpec{
					Target: scscv1.SchedulingTarget{
						Name: "test-deploy",
					},
					Schedule: []scscv1.Schedule{
						{
							Type:     "fixed",
							Runat:    "* * * * *",
							Replicas: &scaledReplica,
						},
					},
				},
				Status: scscv1.ScheduledScalerStatus{
					Phase:   scscv1.StatusFailed,
					Message: "Scheduled Scaler is failed",
					Reason:  scscv1.InternalLogicError,
				},
			},
			expectedFinalizer: []string{finalizer},
			expectedStatus: scscv1.ScheduledScalerStatus{
				Phase:   scscv1.StatusUpdating,
				Message: "Scheduled Scaler is running",
				Reason:  scscv1.NeedToReconcile,
			},
			isCronUpdated:    false,
			isCronRemoved:    false,
			cronUpdateFailed: false,
		},
		"scheduled scaler in running status when not changed": {
			scsc: &scscv1.ScheduledScaler{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "test-scsc",
					Namespace:  "test-ns",
					Finalizers: []string{finalizer},
					Generation: 1,
				},
				Spec: scscv1.ScheduledScalerSpec{
					Target: scscv1.SchedulingTarget{
						Name: "test-deploy",
					},
					Schedule: []scscv1.Schedule{
						{
							Type:     "fixed",
							Runat:    "* * * * *",
							Replicas: &scaledReplica,
						},
					},
				},
				Status: scscv1.ScheduledScalerStatus{
					Phase:   scscv1.StatusRunning,
					Message: "Scheduled Scaler is running",
					Reason:  scscv1.ReconcileDone,
				},
			},
			expectedFinalizer: []string{finalizer},
			expectedStatus: scscv1.ScheduledScalerStatus{
				Phase:   scscv1.StatusRunning,
				Message: "Scheduled Scaler is running",
				Reason:  scscv1.ReconcileDone,
			},
			isCronUpdated:    false,
			isCronRemoved:    false,
			cronUpdateFailed: false,
			inCache: &scscv1.ScheduledScaler{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "test-scsc",
					Namespace:  "test-ns",
					Finalizers: []string{finalizer},
					Generation: 1,
				},
				Spec: scscv1.ScheduledScalerSpec{
					Target: scscv1.SchedulingTarget{
						Name: "test-deploy",
					},
					Schedule: []scscv1.Schedule{
						{
							Type:     "fixed",
							Runat:    "* * * * *",
							Replicas: &scaledReplica,
						},
					},
				},
				Status: scscv1.ScheduledScalerStatus{
					Phase:   scscv1.StatusRunning,
					Message: "Scheduled Scaler is running",
					Reason:  scscv1.ReconcileDone,
				},
			},
		},
		"scheduled scaler in running status when changed": {
			scsc: &scscv1.ScheduledScaler{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "test-scsc",
					Namespace:  "test-ns",
					Finalizers: []string{finalizer},
					Generation: 1,
				},
				Spec: scscv1.ScheduledScalerSpec{
					Target: scscv1.SchedulingTarget{
						Name: "test-deploy",
					},
					Schedule: []scscv1.Schedule{
						{
							Type:     "fixed",
							Runat:    "* * * * *",
							Replicas: &scaledReplica,
						},
					},
				},
				Status: scscv1.ScheduledScalerStatus{
					Phase:   scscv1.StatusRunning,
					Message: "Scheduled Scaler is running",
					Reason:  scscv1.ReconcileDone,
				},
			},
			expectedFinalizer: []string{finalizer},
			expectedStatus: scscv1.ScheduledScalerStatus{
				Phase:   scscv1.StatusUpdating,
				Message: "Scheduled Scaler is running",
				Reason:  scscv1.NeedToReconcile,
			},
			isCronUpdated:    false,
			isCronRemoved:    false,
			cronUpdateFailed: false,
			inCache: &scscv1.ScheduledScaler{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "test-scsc",
					Namespace:  "test-ns",
					Finalizers: []string{finalizer},
					Generation: 2, // generation changed
				},
				Spec: scscv1.ScheduledScalerSpec{
					Target: scscv1.SchedulingTarget{
						Name: "test-deploy",
					},
					Schedule: []scscv1.Schedule{
						{
							Type:     "fixed",
							Runat:    "* * * * *",
							Replicas: &scaledReplica,
						},
					},
				},
				Status: scscv1.ScheduledScalerStatus{
					Phase:   scscv1.StatusRunning,
					Message: "Scheduled Scaler is running",
					Reason:  scscv1.ReconcileDone,
				},
			},
		},
		"scheduled scaler in deleting process": {
			scsc: &scscv1.ScheduledScaler{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "test-scsc",
					Namespace:         "test-ns",
					Finalizers:        []string{finalizer},
					DeletionTimestamp: &nowTime,
				},
				Spec: scscv1.ScheduledScalerSpec{
					Target: scscv1.SchedulingTarget{
						Name: "test-deploy",
					},
					Schedule: []scscv1.Schedule{
						{
							Type:     "fixed",
							Runat:    "* * * * *",
							Replicas: &scaledReplica,
						},
					},
				},
			},
			isCronUpdated:    false,
			isCronRemoved:    true,
			cronUpdateFailed: false,
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			// set test cases
			fakeCli := fake.NewFakeClientWithScheme(s)

			// mocking cron manager
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockCronManager := cronFake.NewMockCronManager(ctrl)
			cronUpdated := false
			if c.isCronUpdated {
				mockCronManager.EXPECT().UpdateCron(gomock.Any()).DoAndReturn(func(*scscv1.ScheduledScaler) error {
					cronUpdated = true
					return nil
				})
			} else if c.cronUpdateFailed {
				mockCronManager.EXPECT().UpdateCron(gomock.Any()).DoAndReturn(func(*scscv1.ScheduledScaler) error {
					return errors.New("cron update fail")
				})
			}

			cronRemoved := false
			if c.isCronRemoved {
				mockCronManager.EXPECT().RemoveCron(gomock.Any()).DoAndReturn(func(*scscv1.ScheduledScaler) error {
					cronRemoved = true
					return nil
				})
			}

			testController := &ScheduledScalerReconciler{
				Client:      fakeCli,
				Log:         &test.FakeLogger{},
				Scheme:      s,
				cronManager: mockCronManager,
				cache:       cache.New(),
			}

			req := cRuntime.Request{
				NamespacedName: types.NamespacedName{
					Namespace: "test-ns",
					Name:      "test-scsc",
				},
			}
			fakeCli.Create(context.Background(), c.scsc)
			if c.inCache != nil {
				testController.cache.Put(c.inCache)
			}

			// do testing function
			_, err := testController.Reconcile(req)
			result := &scscv1.ScheduledScaler{}
			gettingErr := fakeCli.Get(context.Background(), req.NamespacedName, result)

			// verify
			require.NoError(t, gettingErr)
			require.NoError(t, err)
			require.Equal(t, c.expectedFinalizer, result.ObjectMeta.Finalizers)
			require.Equal(t, c.expectedStatus, result.Status)
			require.Equal(t, c.isCronRemoved, cronRemoved)
			require.Equal(t, c.isCronUpdated, cronUpdated)
		})
	}
}

package cron

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	scscv1 "github.com/tmax-cloud/scheduled-scaler-operator/api/v1"
	"github.com/tmax-cloud/scheduled-scaler-operator/internal/k8s"
	"github.com/tmax-cloud/scheduled-scaler-operator/pkg/apimanager"
	"github.com/tmax-cloud/scheduled-scaler-operator/pkg/cron/fake"
	autov2beta2 "k8s.io/api/autoscaling/v2beta2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	fakeCli "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestCronManager_UpdateCron(t *testing.T) {
	s := runtime.NewScheme()
	utilruntime.Must(scscv1.AddToScheme(s))
	utilruntime.Must(autov2beta2.AddToScheme(s))

	scaledReplica := int32(2)
	min := int32(1)
	max := int32(3)
	tc := map[string]struct {
		scsc          *scscv1.ScheduledScaler
		hpa           *autov2beta2.HorizontalPodAutoscaler
		previosExists bool
	}{
		"first initiated": {
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
		},
		"when previos cron exists": {
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
			previosExists: true,
		},
		"when range scaling exists": {
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
			hpa: &autov2beta2.HorizontalPodAutoscaler{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-scsc-hpa",
					Namespace: "test-ns",
				},
				Spec: autov2beta2.HorizontalPodAutoscalerSpec{
					MinReplicas: &min,
					MaxReplicas: max,
				},
			},
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			// set test case
			fakeClient := fakeCli.NewFakeClientWithScheme(s)
			key := apimanager.GetNamespacedName(*c.scsc)
			testCronManager := &CronManagerImpl{
				Client:       fakeClient,
				scheduleCron: make(map[string]Cron),
			}

			previosStop := false
			if c.previosExists {
				ctrl := gomock.NewController(t)
				defer ctrl.Finish()
				m := fake.NewMockCron(ctrl)
				m.EXPECT().Stop().Do(func() {
					previosStop = true
				})
				testCronManager.scheduleCron[key] = m
			}

			if c.hpa != nil {
				require.NoError(t, fakeClient.Create(context.Background(), c.hpa))
			}

			// do testing function
			err := testCronManager.UpdateCron(c.scsc)

			// verify by cases
			require.NoError(t, err)
			require.Len(t, testCronManager.scheduleCron, 1)
			_, exist := testCronManager.scheduleCron[key]
			require.True(t, exist)

			if c.previosExists {
				require.True(t, previosStop)
			}

			if c.hpa != nil {
				hpaName := c.hpa.Name
				hpaNamespace := c.hpa.Namespace
				test, err := k8s.GetHpa(fakeClient, hpaName, hpaNamespace)
				require.Nil(t, test)
				require.Nil(t, err)
			}
		})
	}
}

func TestCronManager_RemoveCron(t *testing.T) {
	s := runtime.NewScheme()
	utilruntime.Must(scscv1.AddToScheme(s))
	utilruntime.Must(autov2beta2.AddToScheme(s))

	scaledReplica := int32(2)
	min := int32(1)
	max := int32(3)
	tc := map[string]struct {
		scsc  *scscv1.ScheduledScaler
		hpa   *autov2beta2.HorizontalPodAutoscaler
		exist bool
	}{
		"Cron exists": {
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
		},
		"Cron & HPA exists": {
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
			hpa: &autov2beta2.HorizontalPodAutoscaler{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-scsc-hpa",
					Namespace: "test-ns",
				},
				Spec: autov2beta2.HorizontalPodAutoscalerSpec{
					MinReplicas: &min,
					MaxReplicas: max,
				},
			},
			exist: true,
		},
		"Cron doesn't exist": {
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
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			// set test case
			fakeClient := fakeCli.NewFakeClientWithScheme(s)
			key := apimanager.GetNamespacedName(*c.scsc)
			testCronManager := &CronManagerImpl{
				Client:       fakeClient,
				scheduleCron: make(map[string]Cron),
			}

			previosStop := false
			if c.exist {
				ctrl := gomock.NewController(t)
				defer ctrl.Finish()
				m := fake.NewMockCron(ctrl)
				m.EXPECT().Stop().Do(func() {
					previosStop = true
				})
				testCronManager.scheduleCron[key] = m
			}

			if c.hpa != nil {
				require.NoError(t, fakeClient.Create(context.Background(), c.hpa))
			}

			// do testing function
			err := testCronManager.RemoveCron(c.scsc)

			// verify by cases
			require.NoError(t, err)
			require.Len(t, testCronManager.scheduleCron, 0)
			_, exist := testCronManager.scheduleCron[key]
			require.False(t, exist)

			if c.exist {
				require.True(t, previosStop)
			}

			if c.hpa != nil {
				hpaName := c.hpa.Name
				hpaNamespace := c.hpa.Namespace
				test, err := k8s.GetHpa(fakeClient, hpaName, hpaNamespace)
				require.Nil(t, test)
				require.Nil(t, err)
			}
		})
	}
}

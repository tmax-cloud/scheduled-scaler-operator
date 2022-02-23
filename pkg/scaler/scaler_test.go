package scaler

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	scscv1 "github.com/tmax-cloud/scheduled-scaler-operator/api/v1"
	"github.com/tmax-cloud/scheduled-scaler-operator/internal/k8s"
	appsv1 "k8s.io/api/apps/v1"
	autov2beta2 "k8s.io/api/autoscaling/v2beta2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestScaler_Run(t *testing.T) {
	s := runtime.NewScheme()
	utilruntime.Must(scscv1.AddToScheme(s))
	utilruntime.Must(appsv1.AddToScheme(s))
	utilruntime.Must(autov2beta2.AddToScheme(s))

	replica := int32(1)
	scaledReplica := int32(2)
	min := int32(1)
	max := int32(3)

	tc := map[string]struct {
		scsc          *scscv1.ScheduledScaler
		target        *appsv1.Deployment
		hpa           *autov2beta2.HorizontalPodAutoscaler
		types         string
		multiSchedule bool
	}{
		"fixed scaling": {
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
			target: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-deploy",
					Namespace: "test-ns",
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: &replica,
				},
			},
			types: "fixed",
		},
		"fixed scaling after range scaling": {
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
			target: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-deploy",
					Namespace: "test-ns",
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: &replica,
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
			types:         "fixed",
			multiSchedule: true,
		},
		"range scaling": {
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
							Type:        "range",
							Runat:       "* * * * *",
							MinReplicas: &min,
							MaxReplicas: &max,
						},
					},
				},
			},
			target: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-deploy",
					Namespace: "test-ns",
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: &replica,
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
			types: "range",
		},
		"range scaling after fixed scaling": {
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
							Type:        "range",
							Runat:       "* * * * *",
							MinReplicas: &min,
							MaxReplicas: &max,
						},
					},
				},
			},
			target: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-deploy",
					Namespace: "test-ns",
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: &scaledReplica,
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
			types: "range",
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			// set test case
			fakeCli := fake.NewFakeClientWithScheme(s)
			require.NoError(t, fakeCli.Create(context.Background(), c.target))
			testScaler, err := New(fakeCli, c.scsc.Name, c.scsc.Namespace, c.scsc.Spec.Target.Name, c.scsc.Spec.Schedule[0])
			require.NoError(t, err)
			if c.multiSchedule {
				if c.types == "fixed" {
					// in multi schedule, hpa already exists before fixed scaling
					k8s.UpdateHpa(fakeCli, &k8s.HpaValidationOptions{
						Namespace:           c.scsc.Namespace,
						Target:              c.target.Name,
						ScheduledScalerName: c.scsc.Name,
						MinReplicas:         c.hpa.Spec.MinReplicas,
						MaxReplicas:         &c.hpa.Spec.MaxReplicas,
					})
					hpa, err := k8s.GetHpa(fakeCli, k8s.GetHpaName(c.scsc.Name), c.scsc.Namespace)
					require.NoError(t, err)
					require.Equal(t, c.hpa.Name, hpa.Name)
					require.Equal(t, c.hpa.Namespace, hpa.Namespace)
					require.Equal(t, c.hpa.Spec.MinReplicas, hpa.Spec.MinReplicas)
					require.Equal(t, c.hpa.Spec.MaxReplicas, hpa.Spec.MaxReplicas)
				}
			}

			// do testing function
			testScaler.Run()

			// verify by cases
			if c.types == "fixed" {
				// when fixed scaling
				scaled, _ := k8s.GetTargetDeployment(fakeCli, c.target.Name, c.target.Namespace)
				require.Equal(t, scaledReplica, *(scaled.Spec.Replicas))
				if c.multiSchedule {
					// after fixed scaling, hpa must be deleted
					hpa, err := k8s.GetHpa(fakeCli, k8s.GetHpaName(c.scsc.Name), c.scsc.Namespace)
					require.Nil(t, hpa)
					require.Nil(t, err)
				}
			} else {
				// when range scaling
				// after range scaling, current replicas must be min replicas at first
				scaled, _ := k8s.GetTargetDeployment(fakeCli, c.target.Name, c.target.Namespace)
				require.Equal(t, min, *(scaled.Spec.Replicas))
				// after range scaling, hpa must be created
				hpa, err := k8s.GetHpa(fakeCli, k8s.GetHpaName(c.scsc.Name), c.scsc.Namespace)
				require.NoError(t, err)
				require.Equal(t, c.hpa.Name, hpa.Name)
				require.Equal(t, c.hpa.Namespace, hpa.Namespace)
				require.Equal(t, c.hpa.Spec.MinReplicas, hpa.Spec.MinReplicas)
				require.Equal(t, c.hpa.Spec.MaxReplicas, hpa.Spec.MaxReplicas)
			}
		})
	}
}

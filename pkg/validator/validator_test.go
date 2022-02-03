package validator

import (
	"testing"

	"github.com/stretchr/testify/require"
	scscv1 "github.com/tmax-cloud/scheduled-scaler-operator/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestValidator_Validate(t *testing.T) {
	replica := int32(1)
	min := int32(1)
	max := int32(3)
	tc := map[string]struct {
		scsc  *scscv1.ScheduledScaler
		valid bool
	}{
		"fixed valid": {
			scsc: &scscv1.ScheduledScaler{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-scsc",
					Namespace: "test-ns",
				},
				Spec: scscv1.ScheduledScalerSpec{
					Schedule: []scscv1.Schedule{
						{
							Type:     "fixed",
							Runat:    "* * * * *",
							Replicas: &replica,
						},
					},
				},
			},
			valid: true,
		},
		"fixed invalid: no replicas": {
			scsc: &scscv1.ScheduledScaler{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-scsc",
					Namespace: "test-ns",
				},
				Spec: scscv1.ScheduledScalerSpec{
					Schedule: []scscv1.Schedule{
						{
							Type:  "fixed",
							Runat: "* * * * *",
						},
					},
				},
			},
			valid: false,
		},
		"fixed invalid: range spec is input": {
			scsc: &scscv1.ScheduledScaler{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-scsc",
					Namespace: "test-ns",
				},
				Spec: scscv1.ScheduledScalerSpec{
					Schedule: []scscv1.Schedule{
						{
							Type:        "fixed",
							Runat:       "* * * * *",
							MinReplicas: &min,
							MaxReplicas: &max,
						},
					},
				},
			},
			valid: false,
		},
		"range valid": {
			scsc: &scscv1.ScheduledScaler{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-scsc",
					Namespace: "test-ns",
				},
				Spec: scscv1.ScheduledScalerSpec{
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
			valid: true,
		},
		"range invalid: missing minReplicas": {
			scsc: &scscv1.ScheduledScaler{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-scsc",
					Namespace: "test-ns",
				},
				Spec: scscv1.ScheduledScalerSpec{
					Schedule: []scscv1.Schedule{
						{
							Type:        "range",
							Runat:       "* * * * *",
							MaxReplicas: &max,
						},
					},
				},
			},
			valid: false,
		},
		"range invalid: missing maxReplicas": {
			scsc: &scscv1.ScheduledScaler{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-scsc",
					Namespace: "test-ns",
				},
				Spec: scscv1.ScheduledScalerSpec{
					Schedule: []scscv1.Schedule{
						{
							Type:        "range",
							Runat:       "* * * * *",
							MinReplicas: &min,
						},
					},
				},
			},
			valid: false,
		},
		"range invalid: missing spec": {
			scsc: &scscv1.ScheduledScaler{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-scsc",
					Namespace: "test-ns",
				},
				Spec: scscv1.ScheduledScalerSpec{
					Schedule: []scscv1.Schedule{
						{
							Type:  "range",
							Runat: "* * * * *",
						},
					},
				},
			},
			valid: false,
		},
		"range invalid: fixed spec is input": {
			scsc: &scscv1.ScheduledScaler{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-scsc",
					Namespace: "test-ns",
				},
				Spec: scscv1.ScheduledScalerSpec{
					Schedule: []scscv1.Schedule{
						{
							Type:     "range",
							Runat:    "* * * * *",
							Replicas: &replica,
						},
					},
				},
			},
			valid: false,
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			// set test case
			testValidator := New(*c.scsc)

			// do testing function
			valid := testValidator.Validate()

			// verify by cases
			if c.valid {
				// when valid scsc is input
				require.True(t, valid)
			} else {
				require.False(t, valid)
			}
		})
	}
}

package cache

import (
	"testing"

	"github.com/stretchr/testify/require"
	scscv1 "github.com/tmax-cloud/scheduled-scaler-operator/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCache_Put(t *testing.T) {
	tc := map[string]struct {
		scsc         *scscv1.ScheduledScaler
		errorOccurs  bool
		errorMessage string
	}{
		"success": {
			scsc: &scscv1.ScheduledScaler{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-scsc",
					Namespace: "test-ns",
				},
			},
		},
		"failed": {
			scsc:         &scscv1.ScheduledScaler{},
			errorOccurs:  true,
			errorMessage: "There's empty namespaced name in scheduledScaler instance",
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			// set test case
			testCache := &ScheduledScalerCacheImpl{
				cached: make(map[string]*scscv1.ScheduledScaler),
			}

			// do testing function
			err := testCache.Put(c.scsc)

			// verify by cases
			if c.errorOccurs {
				// when invalid scsc is input
				require.Error(t, err)
				require.Equal(t, c.errorMessage, err.Error())
			} else {
				// when valid scsc is input
				require.NoError(t, err)
				require.Len(t, testCache.cached, 1)
			}
		})
	}
}

func TestCache_Get(t *testing.T) {
	tc := map[string]struct {
		scsc  *scscv1.ScheduledScaler
		exist bool
	}{
		"exist": {
			scsc: &scscv1.ScheduledScaler{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-scsc",
					Namespace: "test-ns",
				},
			},
			exist: true,
		},
		"failed": {
			scsc: &scscv1.ScheduledScaler{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-scsc",
					Namespace: "test-ns",
				},
			},
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			// set test case
			testCache := New()
			if c.exist {
				testCache.Put(c.scsc)
			}

			// do testing function
			get := testCache.Get(c.scsc)

			// verify by cases
			if c.exist {
				require.Equal(t, c.scsc, get)
			} else {
				require.Nil(t, get)
			}
		})
	}
}

func TestCache_Remove(t *testing.T) {
	tc := map[string]struct {
		scsc        *scscv1.ScheduledScaler
		anotherScsc *scscv1.ScheduledScaler
		exist       bool
	}{
		"remove existing scsc": {
			scsc: &scscv1.ScheduledScaler{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-scsc",
					Namespace: "test-ns",
				},
			},
			exist: true,
		},
		"remove none existing scsc": {
			scsc: &scscv1.ScheduledScaler{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-scsc",
					Namespace: "test-ns",
				},
			},
			anotherScsc: &scscv1.ScheduledScaler{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-scsc2",
					Namespace: "test-ns",
				},
			},
			exist: false,
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			// set test cases
			testCache := &ScheduledScalerCacheImpl{
				cached: make(map[string]*scscv1.ScheduledScaler),
			}
			if c.exist {
				testCache.Put(c.scsc)
			} else {
				testCache.Put(c.anotherScsc)
			}
			require.Len(t, testCache.cached, 1)

			// do testing function
			testCache.Remove(c.scsc)

			// verify by cases
			if c.exist {
				require.Len(t, testCache.cached, 0)
			} else {
				require.Len(t, testCache.cached, 1)
			}
		})
	}
}

func TestCache_HasChanged(t *testing.T) {
	tc := map[string]struct {
		scsc        *scscv1.ScheduledScaler
		anotherScsc *scscv1.ScheduledScaler
		changed     bool
	}{
		"changed scsc": {
			scsc: &scscv1.ScheduledScaler{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "test-scsc",
					Namespace:  "test-ns",
					Generation: 1,
				},
			},
			anotherScsc: &scscv1.ScheduledScaler{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "test-scsc",
					Namespace:  "test-ns",
					Generation: 2,
				},
			},
			changed: true,
		},
		"not changed scsc": {
			scsc: &scscv1.ScheduledScaler{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "test-scsc",
					Namespace:  "test-ns",
					Generation: 1,
				},
			},
			anotherScsc: &scscv1.ScheduledScaler{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "test-scsc",
					Namespace:  "test-ns",
					Generation: 1,
				},
			},
			changed: false,
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			// set test cases
			testCache := New()
			testCache.Put(c.scsc)

			// do testing function
			changed := testCache.HasChanged(c.anotherScsc)

			// verify by cases
			if c.changed {
				require.True(t, changed)
			} else {
				require.False(t, changed)
			}
		})
	}
}

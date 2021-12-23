package cache

import (
	"fmt"

	scscv1 "github.com/tmax-cloud/scheduled-scaler-operator/api/v1"
	"github.com/tmax-cloud/scheduled-scaler-operator/pkg/apimanager"
)

type ScheduledScalerCache struct {
	cached map[string]*scscv1.ScheduledScaler
}

func New() *ScheduledScalerCache {
	return &ScheduledScalerCache{
		cached: make(map[string]*scscv1.ScheduledScaler),
	}
}

func (c *ScheduledScalerCache) Put(scsc *scscv1.ScheduledScaler) error {
	if key := apimanager.GetNamespacedName(*scsc); key == "" {
		return fmt.Errorf("There's empty namespaced name in scheduledScaler instance")
	} else {
		c.cached[key] = scsc
		return nil
	}
}

func (c *ScheduledScalerCache) Get(scsc *scscv1.ScheduledScaler) *scscv1.ScheduledScaler {
	return c.cached[apimanager.GetNamespacedName(*scsc)]
}

func (c *ScheduledScalerCache) Remove(scsc *scscv1.ScheduledScaler) {
	if c.exist(*scsc) {
		delete(c.cached, apimanager.GetNamespacedName(*scsc))
	}
}

func (c *ScheduledScalerCache) HasChanged(scsc *scscv1.ScheduledScaler) bool {
	if !c.exist(*scsc) {
		return true
	}

	oldScsc := c.Get(scsc)
	return scsc.ObjectMeta.Generation != oldScsc.ObjectMeta.Generation
}

func (c *ScheduledScalerCache) exist(scsc scscv1.ScheduledScaler) bool {
	_, exist := c.cached[apimanager.GetNamespacedName(scsc)]
	return exist
}

package validator

import scscv1 "github.com/tmax-cloud/scheduled-scaler-operator/api/v1"

type Validator interface {
	Validate() bool
}

type ValidatorImpl struct {
	source scscv1.ScheduledScaler
}

func NewValidator(source scscv1.ScheduledScaler) Validator {
	return &ValidatorImpl{
		source: source,
	}
}

func (v *ValidatorImpl) Validate() bool {
	for _, schedule := range v.source.Spec.Schedule {
		if schedule.Type == "fixed" {
			if !v.fixedScheduleValidate(schedule) {
				return false
			}
		} else {
			if !v.rangeScheduleValidate(schedule) {
				return false
			}
		}
	}

	return true
}

func (v *ValidatorImpl) fixedScheduleValidate(schedule scscv1.Schedule) bool {
	if schedule.Replicas == nil {
		return false
	}

	if schedule.MinReplicas != nil || schedule.MaxReplicas != nil {
		return false
	}

	return true
}

func (v *ValidatorImpl) rangeScheduleValidate(schedule scscv1.Schedule) bool {
	if schedule.Replicas != nil {
		return false
	}

	if schedule.MinReplicas == nil || schedule.MaxReplicas == nil {
		return false
	}

	return true
}

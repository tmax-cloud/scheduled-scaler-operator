package validator

import scscv1 "github.com/tmax-cloud/scheduled-scaler-operator/api/v1"

type Validator struct {
	source scscv1.ScheduledScaler
}

func NewValidator(source scscv1.ScheduledScaler) *Validator {
	return &Validator{
		source: source,
	}
}

func (v *Validator) Validate() bool {
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

func (v *Validator) fixedScheduleValidate(schedule scscv1.Schedule) bool {
	if schedule.Replicas == nil {
		return false
	}

	if schedule.MinReplicas != nil || schedule.MaxReplicas != nil {
		return false
	}

	return true
}

func (v *Validator) rangeScheduleValidate(schedule scscv1.Schedule) bool {
	if schedule.Replicas != nil {
		return false
	}

	if schedule.MinReplicas == nil || schedule.MaxReplicas == nil {
		return false
	}

	return true
}
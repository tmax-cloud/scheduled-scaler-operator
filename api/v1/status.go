package v1

type Status string
type Reason string

const (
	StatusUpdating = Status("Updating")
	StatusRunning  = Status("Running")
	StatusFailed   = Status("Failed")
)

const (
	NeedToReconcile = Reason("Updating")
	ReconcileDone   = Reason("Done")
)

const (
	InternalLogicError    = Reason("InternalLogicError")
	ValidationFailedError = Reason("InvalidSpecError")
)

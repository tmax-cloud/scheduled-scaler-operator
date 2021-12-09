package v1

type Status string

const (
	StatusRunning = Status("Running")
	StatusFailed  = Status("Failed")
)

package v1

type Status string

const (
	StatusRunning     = Status("Running")
	StatusCreating    = Status("Creating")
	StatusFailed      = Status("Failed")
	StatusTerminating = Status("Terminating")
)

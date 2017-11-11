package job

type JobState int

// Job states
const (
	Unknown = iota
	Created
	Queued
	Starting
	Started
	Paused
	Cancelled
	Finished
	Failed
)

type ProcessState int

// Process states
const (
	NonExisting = iota
	Running
	Complete
	Terminated
)

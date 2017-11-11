package scheduler

type Policy int

const (
	Fairest = iota
	Graceful
	GroupBased
	Preemptive
)

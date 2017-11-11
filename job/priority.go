package job

type Priority int

const (
	Lowest = iota
	Low
	Normal
	High
	Realtime
)

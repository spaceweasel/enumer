package testpkg

// RunStatus represents a flag-based enum with composite values
type RunStatus int

const (
	Pending   RunStatus = 1 << iota
	Running
	Success
	Failure
	Skipped
	Completed RunStatus = Success | Failure | Skipped
)

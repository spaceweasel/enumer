package testpkg

// Status represents a simple enum using iota
type Status int

const (
	Pending Status = iota
	Running
	Success
	Failure
)

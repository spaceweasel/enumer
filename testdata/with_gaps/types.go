package testpkg

// Priority represents an enum with explicit values and gaps
type Priority int

const (
	Low    Priority = 1
	Medium Priority = 5
	High   Priority = 10
	Urgent Priority = 20
)

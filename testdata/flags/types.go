package testpkg

// Permission represents a flag-based enum using bit shifts
type Permission int

const (
	Read    Permission = 1 << iota
	Write
	Execute
	Delete
)

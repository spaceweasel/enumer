package testpkg

// Direction represents an enum with a common prefix
type Direction int

const (
	DirectionNorth Direction = iota
	DirectionEast
	DirectionSouth
	DirectionWest
)

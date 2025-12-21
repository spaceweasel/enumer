package testpkg

import "testing"

func TestDirectionTrimPrefix(t *testing.T) {
	// String should have prefix trimmed
	if DirectionNorth.String() != "North" {
		t.Errorf("Expected 'North', got %q", DirectionNorth.String())
	}
	if DirectionEast.String() != "East" {
		t.Errorf("Expected 'East', got %q", DirectionEast.String())
	}

	// Parse using trimmed name
	dir, err := DirectionString("South")
	if err != nil {
		t.Fatalf("Failed to parse 'South': %v", err)
	}
	if dir != DirectionSouth {
		t.Error("Parsed 'South' doesn't match DirectionSouth")
	}

	// Full name should not work
	_, err = DirectionString("DirectionWest")
	if err == nil {
		t.Error("Should not be able to parse 'DirectionWest' with trimprefix")
	}
}

func TestDirectionNumericValues(t *testing.T) {
	if int(DirectionNorth) != 0 {
		t.Errorf("DirectionNorth should be 0, got %d", DirectionNorth)
	}
	if int(DirectionWest) != 3 {
		t.Errorf("DirectionWest should be 3, got %d", DirectionWest)
	}
}

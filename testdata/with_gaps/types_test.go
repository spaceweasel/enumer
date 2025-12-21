package testpkg

import "testing"

func TestPriorityValues(t *testing.T) {
	// Test exact numeric values with gaps
	if int(Low) != 1 {
		t.Errorf("Low should be 1, got %d", Low)
	}
	if int(Medium) != 5 {
		t.Errorf("Medium should be 5, got %d", Medium)
	}
	if int(High) != 10 {
		t.Errorf("High should be 10, got %d", High)
	}
	if int(Urgent) != 20 {
		t.Errorf("Urgent should be 20, got %d", Urgent)
	}
}

func TestPriorityParse(t *testing.T) {
	p, err := PriorityString("Medium")
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}
	if int(p) != 5 {
		t.Errorf("Medium should have value 5, got %d", p)
	}
}

func TestPriorityInvalid(t *testing.T) {
	invalid := Priority(7)
	if invalid.Valid() {
		t.Error("Priority(7) should not be valid")
	}
	if invalid.String() != "Priority(7)" {
		t.Errorf("Invalid value string incorrect: %s", invalid.String())
	}
}

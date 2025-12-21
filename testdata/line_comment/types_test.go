package testpkg

import "testing"

func TestColorLineComments(t *testing.T) {
	// With linecomment, string values should come from comments
	if Red.String() != "red" {
		t.Errorf("Expected 'red', got %q", Red.String())
	}
	if Green.String() != "green" {
		t.Errorf("Expected 'green', got %q", Green.String())
	}

	// Parse using comment text
	color, err := ColorString("blue")
	if err != nil {
		t.Fatalf("Failed to parse 'blue': %v", err)
	}
	if color != Blue {
		t.Error("Parsed 'blue' doesn't match Blue constant")
	}

	// Original names should not work
	_, err = ColorString("Blue")
	if err == nil {
		t.Error("Should not be able to parse 'Blue' when using linecomment")
	}
}

func TestColorNumericValues(t *testing.T) {
	if int(Red) != 0 {
		t.Errorf("Red should be 0, got %d", Red)
	}
	if int(Green) != 1 {
		t.Errorf("Green should be 1, got %d", Green)
	}
}

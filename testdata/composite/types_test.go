package testpkg

import (
	"encoding/json"
	"testing"
)

func TestRunStatusValues(t *testing.T) {
	// Test flag values
	if int(Pending) != 1 {
		t.Errorf("Pending should be 1, got %d", Pending)
	}
	if int(Running) != 2 {
		t.Errorf("Running should be 2, got %d", Running)
	}
	if int(Success) != 4 {
		t.Errorf("Success should be 4, got %d", Success)
	}
	if int(Failure) != 8 {
		t.Errorf("Failure should be 8, got %d", Failure)
	}
	if int(Skipped) != 16 {
		t.Errorf("Skipped should be 16, got %d", Skipped)
	}

	// Test composite value
	expectedCompleted := 4 + 8 + 16 // Success | Failure | Skipped
	if int(Completed) != expectedCompleted {
		t.Errorf("Completed should be %d, got %d", expectedCompleted, Completed)
	}
}

func TestRunStatusComposite(t *testing.T) {
	// Verify Completed equals the OR of its components
	combined := Success | Failure | Skipped
	if Completed != combined {
		t.Errorf("Completed (%d) should equal Success|Failure|Skipped (%d)", Completed, combined)
	}
}

func TestRunStatusParse(t *testing.T) {
	// Parse composite value
	status, err := RunStatusString("Completed")
	if err != nil {
		t.Fatalf("Failed to parse Completed: %v", err)
	}
	if status != Completed {
		t.Error("Parsed Completed doesn't match constant")
	}
	if int(status) != 28 {
		t.Errorf("Completed should have value 28, got %d", status)
	}
}

func TestRunStatusValidation(t *testing.T) {
	// Named constants should be valid
	if !Pending.Valid() {
		t.Error("Pending should be valid")
	}
	if !Completed.Valid() {
		t.Error("Completed should be valid")
	}

	// Arbitrary combinations should not be valid
	arbitrary := Pending | Running
	if arbitrary.Valid() {
		t.Error("Arbitrary combination should not be valid")
	}
}

func TestRunStatusHas(t *testing.T) {
	// Test Has method - checking if a specific flag is set
	rs := Completed // Completed = Success | Failure | Skipped

	if !rs.Has(Success) {
		t.Error("Completed should have Success flag")
	}
	if !rs.Has(Failure) {
		t.Error("Completed should have Failure flag")
	}
	if !rs.Has(Skipped) {
		t.Error("Completed should have Skipped flag")
	}
	if rs.Has(Pending) {
		t.Error("Completed should not have Pending flag")
	}
	if rs.Has(Running) {
		t.Error("Completed should not have Running flag")
	}

	// Test with individual flag
	if !Success.Has(Success) {
		t.Error("Success should have Success flag")
	}
	if Success.Has(Failure) {
		t.Error("Success should not have Failure flag")
	}
}

func TestRunStatusHasAny(t *testing.T) {
	rs := Success | Failure // Not a named constant, but valid combination

	if !rs.HasAny(Success, Running) {
		t.Error("Should have at least one of Success or Running")
	}
	if !rs.HasAny(Failure) {
		t.Error("Should have Failure")
	}
	if rs.HasAny(Pending, Running, Skipped) {
		t.Error("Should not have any of Pending, Running, or Skipped")
	}
}

func TestRunStatusHasAll(t *testing.T) {
	rs := Completed // Success | Failure | Skipped

	if !rs.HasAll(Success, Failure, Skipped) {
		t.Error("Completed should have all three flags")
	}
	if rs.HasAll(Success, Failure, Skipped, Pending) {
		t.Error("Completed should not have Pending")
	}
	if !Success.HasAll(Success) {
		t.Error("Success should have Success flag")
	}
}

func TestRunStatusSet(t *testing.T) {
	// Start with Pending
	rs := Pending

	// Set Success flag
	rs = rs.Set(Success)
	if !rs.Has(Success) {
		t.Error("Should have Success after Set")
	}
	if !rs.Has(Pending) {
		t.Error("Should still have Pending")
	}

	// Set multiple flags
	rs = Pending.Set(Success, Failure)
	if !rs.HasAll(Success, Failure, Pending) {
		t.Error("Should have all three flags")
	}
}

func TestRunStatusClear(t *testing.T) {
	// Start with Completed (Success | Failure | Skipped)
	rs := Completed

	// Clear one flag
	rs = rs.Clear(Success)
	if rs.Has(Success) {
		t.Error("Should not have Success after Clear")
	}
	if !rs.Has(Failure) {
		t.Error("Should still have Failure")
	}

	// Clear multiple flags
	rs = Completed.Clear(Success, Skipped)
	if !rs.Has(Failure) {
		t.Error("Should only have Failure remaining")
	}
	if rs.Has(Success) || rs.Has(Skipped) {
		t.Error("Should not have Success or Skipped")
	}
}

func TestRunStatusToggle(t *testing.T) {
	rs := Success

	// Toggle off
	rs = rs.Toggle(Success)
	if rs.Has(Success) {
		t.Error("Success should be toggled off")
	}

	// Toggle on
	rs = rs.Toggle(Success)
	if !rs.Has(Success) {
		t.Error("Success should be toggled back on")
	}

	// Toggle multiple
	rs = Success.Toggle(Success, Failure)
	if rs.Has(Success) {
		t.Error("Success should be off")
	}
	if !rs.Has(Failure) {
		t.Error("Failure should be on")
	}
}

func TestRunStatusJSON(t *testing.T) {
	// Marshal composite value
	data, err := json.Marshal(Completed)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	if string(data) != `"Completed"` {
		t.Errorf("Expected \"Completed\", got %s", data)
	}

	// Unmarshal composite value
	var status RunStatus
	err = json.Unmarshal([]byte(`"Completed"`), &status)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if status != Completed {
		t.Error("Unmarshaled value doesn't match")
	}
	if int(status) != 28 {
		t.Errorf("Unmarshaled Completed should be 28, got %d", status)
	}
}

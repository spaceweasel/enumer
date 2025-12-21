package testpkg

import (
	"database/sql/driver"
	"encoding/json"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestStatusValues(t *testing.T) {
	// Test exact numeric values
	if int(Pending) != 0 {
		t.Errorf("Pending should be 0, got %d", Pending)
	}
	if int(Running) != 1 {
		t.Errorf("Running should be 1, got %d", Running)
	}
	if int(Success) != 2 {
		t.Errorf("Success should be 2, got %d", Success)
	}
	if int(Failure) != 3 {
		t.Errorf("Failure should be 3, got %d", Failure)
	}
}

func TestStatusString(t *testing.T) {
	if Pending.String() != "Pending" {
		t.Errorf("Expected 'Pending', got %q", Pending.String())
	}
	if Running.String() != "Running" {
		t.Errorf("Expected 'Running', got %q", Running.String())
	}
}

func TestStatusParse(t *testing.T) {
	s, err := StatusString("Success")
	if err != nil {
		t.Fatalf("Failed to parse 'Success': %v", err)
	}
	if s != Success {
		t.Errorf("Expected Success, got %v", s)
	}
	if int(s) != 2 {
		t.Errorf("Expected numeric value 2, got %d", s)
	}

	_, err = StatusString("Invalid")
	if err == nil {
		t.Error("Expected error for invalid string")
	}
}

func TestStatusValues_All(t *testing.T) {
	values := StatusValues()
	if len(values) != 4 {
		t.Fatalf("Expected 4 values, got %d", len(values))
	}
	if values[0] != Pending || values[1] != Running || values[2] != Success || values[3] != Failure {
		t.Error("Values in wrong order")
	}
}

func TestStatusValid(t *testing.T) {
	if !Pending.Valid() {
		t.Error("Pending should be valid")
	}

	invalid := Status(99)
	if invalid.Valid() {
		t.Error("Status(99) should not be valid")
	}
}

func TestStatusJSON(t *testing.T) {
	data, err := json.Marshal(Success)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	if string(data) != `"Success"` {
		t.Errorf("Expected \"Success\", got %s", data)
	}

	var s Status
	err = json.Unmarshal([]byte(`"Failure"`), &s)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if s != Failure {
		t.Errorf("Expected Failure, got %v", s)
	}
	if int(s) != 3 {
		t.Errorf("Expected numeric value 3, got %d", s)
	}
}

func TestStatusYAML(t *testing.T) {
	// Marshal
	data, err := yaml.Marshal(Running)
	if err != nil {
		t.Fatalf("YAML marshal failed: %v", err)
	}
	expected := "Running\n"
	if string(data) != expected {
		t.Errorf("Expected %q, got %q", expected, string(data))
	}

	// Unmarshal
	var s Status
	err = yaml.Unmarshal([]byte("Success"), &s)
	if err != nil {
		t.Fatalf("YAML unmarshal failed: %v", err)
	}
	if s != Success {
		t.Errorf("Expected Success, got %v", s)
	}
	if int(s) != 2 {
		t.Errorf("Expected numeric value 2, got %d", s)
	}

	// Unmarshal invalid
	var s2 Status
	err = yaml.Unmarshal([]byte("InvalidStatus"), &s2)
	if err == nil {
		t.Error("Expected error for invalid YAML value")
	}
}

func TestStatusSQL(t *testing.T) {
	// Test Scan with string
	var s Status
	err := s.Scan("Pending")
	if err != nil {
		t.Fatalf("Scan string failed: %v", err)
	}
	if s != Pending {
		t.Errorf("Expected Pending, got %v", s)
	}

	// Test Scan with []byte
	var s2 Status
	err = s2.Scan([]byte("Running"))
	if err != nil {
		t.Fatalf("Scan bytes failed: %v", err)
	}
	if s2 != Running {
		t.Errorf("Expected Running, got %v", s2)
	}

	// Test Scan with nil
	var s3 Status
	err = s3.Scan(nil)
	if err != nil {
		t.Errorf("Scan(nil) should succeed, got: %v", err)
	}

	// Test Scan with invalid type
	var s4 Status
	err = s4.Scan(123)
	if err == nil {
		t.Error("Expected error when scanning int")
	}

	// Test Scan with invalid value
	var s5 Status
	err = s5.Scan("InvalidStatus")
	if err == nil {
		t.Error("Expected error for invalid status value")
	}

	// Test Value
	val, err := Success.Value()
	if err != nil {
		t.Fatalf("Value failed: %v", err)
	}
	if val != driver.Value("Success") {
		t.Errorf("Expected \"Success\", got %v", val)
	}
}

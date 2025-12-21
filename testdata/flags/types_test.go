package testpkg

import (
	"encoding/json"
	"testing"
)

func TestPermissionValues(t *testing.T) {
	// Test flag values are powers of 2
	if int(Read) != 1 {
		t.Errorf("Read should be 1, got %d", Read)
	}
	if int(Write) != 2 {
		t.Errorf("Write should be 2, got %d", Write)
	}
	if int(Execute) != 4 {
		t.Errorf("Execute should be 4, got %d", Execute)
	}
	if int(Delete) != 8 {
		t.Errorf("Delete should be 8, got %d", Delete)
	}
}

func TestPermissionParse(t *testing.T) {
	p, err := PermissionString("Execute")
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}
	if int(p) != 4 {
		t.Errorf("Execute should have value 4, got %d", p)
	}
}

func TestPermissionCombinations(t *testing.T) {
	// Test that arbitrary combinations are not valid
	readWrite := Read | Write
	if int(readWrite) != 3 {
		t.Errorf("Read|Write should be 3, got %d", readWrite)
	}
	if readWrite.Valid() {
		t.Error("Combined Read|Write should not be valid (not a named constant)")
	}

	// Individual flags should be valid
	if !Read.Valid() {
		t.Error("Read should be valid")
	}
}

func TestPermissionHas(t *testing.T) {
	// Test Has method
	perms := Read | Write | Execute // 1 | 2 | 4 = 7

	if !perms.Has(Read) {
		t.Error("Should have Read permission")
	}
	if !perms.Has(Write) {
		t.Error("Should have Write permission")
	}
	if !perms.Has(Execute) {
		t.Error("Should have Execute permission")
	}
	if perms.Has(Delete) {
		t.Error("Should not have Delete permission")
	}
}

func TestPermissionHasAny(t *testing.T) {
	perms := Read | Write

	if !perms.HasAny(Read, Delete) {
		t.Error("Should have at least Read")
	}
	if perms.HasAny(Execute, Delete) {
		t.Error("Should not have Execute or Delete")
	}
}

func TestPermissionHasAll(t *testing.T) {
	perms := Read | Write | Execute

	if !perms.HasAll(Read, Write) {
		t.Error("Should have both Read and Write")
	}
	if perms.HasAll(Read, Write, Delete) {
		t.Error("Should not have all including Delete")
	}
}

func TestPermissionJSON(t *testing.T) {
	data, err := json.Marshal(Write)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	if string(data) != `"Write"` {
		t.Errorf("Expected \"Write\", got %s", data)
	}
}

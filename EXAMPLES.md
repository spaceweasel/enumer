# Enumer Examples

This document provides practical examples of using enumer with different enum patterns.

## Example 1: Simple Sequential Enum

```go
package myapp

//go:generate enumer -type=Status -json
type Status int

const (
    Pending Status = iota
    Active
    Completed
    Cancelled
)
```

**Generated file**: `status_enumer.go`

**Usage**:
```go
s := Active
fmt.Println(s.String()) // "Active"

parsed, err := StatusString("Pending")
if err != nil {
    log.Fatal(err)
}
fmt.Println(parsed) // Pending (value 0)

// JSON
data, _ := json.Marshal(s)
fmt.Println(string(data)) // "Active"

var s2 Status
json.Unmarshal([]byte(`"Completed"`), &s2)
fmt.Println(s2) // Completed (value 2)
```

## Example 2: Flag-Based Enum (Bit Masks)

```go
package myapp

//go:generate enumer -type=Permission -json
type Permission int

const (
    Read    Permission = 1 << iota // 1
    Write                         // 2
    Execute                       // 4
    Delete                        // 8

    ReadWrite Permission = Read | Write           // 3
    FullAccess Permission = Read | Write | Execute | Delete // 15
)
```

**Usage**:
```go
p := Read | Write
fmt.Println(int(p)) // 3

// Check if Read permission is set
if p & Read != 0 {
    fmt.Println("Has read permission")
}

// Named composite values work too
admin := FullAccess
fmt.Println(admin.String())  // "FullAccess"
fmt.Println(int(admin))      // 15
fmt.Println(admin.Valid()) // true

// Individual flags are valid
fmt.Println(Read.Valid())  // true
fmt.Println(Write.Valid()) // true

// Named composites are valid
fmt.Println(FullAccess.Valid()) // true

// Arbitrary combinations not in const block are not valid
arbitrary := Read | Delete // 9
fmt.Println(arbitrary.Valid()) // false (not a named constant)
```

## Example 3: Using Line Comments for Display

```go
package myapp

//go:generate enumer -type=HTTPStatus -linecomment -json
type HTTPStatus int

const (
    StatusOK                   HTTPStatus = 200 // OK
    StatusBadRequest           HTTPStatus = 400 // Bad Request
    StatusUnauthorized         HTTPStatus = 401 // Unauthorized
    StatusForbidden            HTTPStatus = 403 // Forbidden
    StatusNotFound             HTTPStatus = 404 // Not Found
    StatusInternalServerError  HTTPStatus = 500 // Internal Server Error
)
```

**Usage**:
```go
s := StatusNotFound
fmt.Println(s.String()) // "Not Found" (from line comment)

parsed, err := HTTPStatusString("Bad Request")
fmt.Println(parsed) // StatusBadRequest (value 400)
```

## Example 4: Trimming Prefixes

```go
package myapp

//go:generate enumer -type=Color -trimprefix=Color -json
type Color int

const (
    ColorRed   Color = iota
    ColorGreen
    ColorBlue
    ColorYellow
)
```

**Usage**:
```go
c := ColorRed
fmt.Println(c.String()) // "Red" (prefix trimmed)

parsed, err := ColorString("Green")
fmt.Println(parsed) // ColorGreen
```

## Example 5: Database Integration (SQL)

```go
package myapp

import "database/sql"

//go:generate enumer -type=OrderStatus -sql -json
type OrderStatus int

const (
    OrderPending OrderStatus = iota
    OrderProcessing
    OrderShipped
    OrderDelivered
)

type Order struct {
    ID     int
    Status OrderStatus
}
```

**Usage**:
```go
// Insert
_, err := db.Exec(
    "INSERT INTO orders (status) VALUES ($1)",
    OrderPending, // Automatically converted to "OrderPending" string
)

// Query
var order Order
err := db.QueryRow("SELECT id, status FROM orders WHERE id = $1", 1).
    Scan(&order.ID, &order.Status) // Automatically scanned from string

fmt.Println(order.Status) // OrderPending

// Also works with JSON APIs
data, _ := json.Marshal(order)
// {"ID":1,"Status":"OrderPending"}
```

## Example 6: YAML Configuration

```go
package myapp

//go:generate enumer -type=LogLevel -yaml
type LogLevel int

const (
    Debug LogLevel = iota
    Info
    Warning
    Error
)

type Config struct {
    Level LogLevel `yaml:"log_level"`
    File  string   `yaml:"log_file"`
}
```

**YAML file** (`config.yaml`):
```yaml
log_level: Warning
log_file: /var/log/app.log
```

**Usage**:
```go
var cfg Config
data, _ := os.ReadFile("config.yaml")
yaml.Unmarshal(data, &cfg)

fmt.Println(cfg.Level) // Warning
```

## Example 7: Multiple Types in One Command

```go
package myapp

//go:generate enumer -type=Status,Priority,Category -output=enums_generated.go -json

type Status int
const (
    Active Status = iota
    Inactive
)

type Priority int
const (
    Low Priority = iota
    High
)

type Category int
const (
    Personal Category = iota
    Work
)
```

All three types will be generated in a single file: `enums_generated.go`

## Example 8: Complex Flag Enum (Real-World)

Here's a more complex example based on your use case:

```go
package workflow

//go:generate enumer -type=RunStatus -json -sql
type RunStatus int

const (
    Pending   RunStatus = 1 << iota                   // Pending
    Running                                           // Running
    Success                                           // Success
    Failure                                           // Failure
    Skipped                                           // Skipped
    Completed RunStatus = Success | Failure | Skipped // Completed
)
```

**Usage**:
```go
// Track individual statuses
run := Running
fmt.Println(run.String()) // "Running"

// NEW: Use flag methods for working with bitmasks
run = Pending

// Set flags (add flags to the value)
run = run.Set(Success)           // Pending | Success
run = run.Set(Failure, Skipped) // Pending | Success | Failure | Skipped

// Clear flags (remove flags from the value)
run = run.Clear(Pending)        // Success | Failure | Skipped (= Completed)

// Toggle flags (flip the bit)
run = run.Toggle(Success)       // Failure | Skipped (Success removed)
run = run.Toggle(Success)       // Success | Failure | Skipped (Success added back)

// Check if specific flag is set
if run.Has(Failure) {
    fmt.Println("Run has Failure flag")
}

// Check if any of the flags are set
if run.HasAny(Success, Skipped) {
    fmt.Println("Run has at least one terminal state")
}

// Check if all flags are set
if run.HasAll(Success, Failure, Skipped) {
    fmt.Println("All terminal states present")
}

// Traditional equality check (exact match only)
if run == Completed {
    // This is true only if run equals exactly Completed (28)
}

// Values
fmt.Println(int(Pending))   // 1
fmt.Println(int(Running))   // 2
fmt.Println(int(Success))   // 4
fmt.Println(int(Failure))   // 8
fmt.Println(int(Skipped))   // 16
fmt.Println(int(Completed)) // 28 (4 + 8 + 16)

// Database storage
db.Exec("INSERT INTO runs (status) VALUES ($1)", Success)
// Stores "Success" as string

var status RunStatus
db.QueryRow("SELECT status FROM runs WHERE id = 1").Scan(&status)
// Reads "Success" and converts back to enum

// JSON API
type Run struct {
    ID     int       `json:"id"`
    Status RunStatus `json:"status"`
}

run := Run{ID: 1, Status: Success}
data, _ := json.Marshal(run)
// {"id":1,"status":"Success"}
```

## Testing Generated Code

You can test the generated methods:

```go
func TestStatusEnum(t *testing.T) {
    // Test String()
    if Active.String() != "Active" {
        t.Error("Expected 'Active'")
    }

    // Test parsing
    s, err := StatusString("Completed")
    if err != nil || s != Completed {
        t.Error("Failed to parse")
    }

    // Test Valid
    if !Pending.Valid() {
        t.Error("Pending should be valid")
    }

    if Status(999).Valid() {
        t.Error("999 should not be valid")
    }

    // Test Values
    all := StatusValues()
    if len(all) != 4 {
        t.Errorf("Expected 4 values, got %d", len(all))
    }
}
```

## Tips

1. **Always use `-json` for APIs**: If your enum is used in REST APIs, always generate JSON methods
2. **Use `-sql` for database types**: For enums stored in databases, generate SQL methods
3. **Use `-linecomment` for user-facing strings**: When the constant name differs from display text
4. **Use `-trimprefix` to avoid redundancy**: Keeps string representations clean
5. **Flag enums**: Perfect for permission systems, feature flags, and status combinations
6. **Test your enums**: Write tests to verify marshaling, parsing, and validation work as expected

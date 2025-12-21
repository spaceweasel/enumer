# Enumer

A Go code generation tool for creating enum helper methods. This tool is inspired by [alvaroloes/enumer](https://github.com/alvaroloes/enumer) but with key improvements, including proper support for flag-based enums using bit shifts (`1 << iota`) and composite values.

## Features

- **Complete enum support**: Works with simple `iota`, bit-shifted flags (`1 << iota`), explicit numeric values, and composite expressions
- **String conversion**: Automatic `String()` method and reverse parsing
- **JSON marshaling**: Optional JSON marshal/unmarshal methods
- **YAML marshaling**: Optional YAML marshal/unmarshal methods
- **SQL support**: Optional `sql.Scanner` and `driver.Valuer` implementations
- **Bitwise methods**: Optional bitwise methods for flag enums `Has()`, `HasAny()`, `HasAll()`, `Set()`, `Clear()` and `Toggle()`
- **Template-based**: Uses Go templates for clear, maintainable code generation
- **Helper methods**: `Values()`, `Valid()`, and type-safe parsing functions

## Installation

```bash
go install github.com/spaceweasel/enumer@latest
```

## Usage

```bash
enumer -type=TypeName[,OtherType] [options]
```

### Options

- `type`: (required)
Comma-separated list of type names to generate code for.

- `output`: Output filename. Defaults:
    - single type: `<type>_enumer.go`
    - multiple types: `enums_gen.go` (or `flags_gen.go` when `-bitmask` flag is set)
- `trimprefix`: Prefix to trim from constant names in string representation.

- `linecomment`: Use line comment text as the string value when present (when present and non-empty).

- `json`: Generate `MarshalJSON`/`UnmarshalJSON` using the string representation.

- `yaml`: Generate YAML `Marshal`/`Unmarshal` using the string representation.

- `sql`: Generate `Scan` and `Value` for database/sql usage.

- `bitmask`: Generate bitwise methods:
`Has`, `HasAny`, `HasAll`, `Set`, `Clear`, `Toggle`. _Note: These methods will be generated even for non-flag type enums, which although they will compile, they will be semantically meaningless._


### Typical Usage

Add a `//go:generate` comment above your enum type:

```go
//go:generate enumer -type=Status -json -yaml -sql -trimprefix=Status
type Status int

const (
    StatusPending Status = iota
    StatusRunning
    StatusSuccess
    StatusFailure
)
```

Then run:

```bash
go generate ./...
```

## Generated Methods

For a type named `Status`, enumer generates:

### Basic Methods

```go
// StatusString retrieves an enum value from string
func StatusString(s string) (Status, error)

// StatusValues returns all enum values
func StatusValues() []Status

// Valid checks if value is valid
func (i Status) Valid() bool

// String returns the string representation
func (i Status) String() string
```

### JSON Methods (with `-json` flag)

```go
func (i Status) MarshalJSON() ([]byte, error)
func (i *Status) UnmarshalJSON(data []byte) error
```

### YAML Methods (with `-yaml` flag)

```go
func (i Status) MarshalYAML() (any, error)
func (i *Status) UnmarshalYAML(node *yaml.Node) error
```

### SQL Methods (with `-sql` flag)

```go
func (i *Status) Scan(value any) error
func (i Status) Value() (driver.Value, error)
```

### Bitwise Methods (with `-bitmask` flag)

For a type named RunStatus:
```go
//go:generate enumer -type=RunStatus -bitmask
type RunStatus int

const (
    Pending   RunStatus = 1 << iota
    Running
    Success
    Failure
    Skipped
    Completed RunStatus = Success | Failure | Skipped
)
```

```go
// Checking flags
func (i RunStatus) Has(flag RunStatus) bool
func (i RunStatus) HasAny(flags ...RunStatus) bool
func (i RunStatus) HasAll(flags ...RunStatus) bool

// Manipulating flags (returns new value, doesn't modify original)
func (i RunStatus) Set(flags ...RunStatus) RunStatus
func (i RunStatus) Clear(flags ...RunStatus) RunStatus
func (i RunStatus) Toggle(flags ...RunStatus) RunStatus
```

**Example usage:**
```go
status := Pending

// Check flags
if status.Has(Failure) {
    // Handle failure case
}

// Set flags (returns new value)
status = status.Set(Success, Skipped)  // Now has Pending | Success | Skipped

// Clear flags
status = status.Clear(Pending)  // Remove Pending flag

// Toggle flags
status = status.Toggle(Failure)  // Add Failure if not present, remove if present

// Check multiple
if status.HasAny(Success, Skipped) {
    // At least one is set
}

if status.HasAll(Success, Skipped) {
    // Both are set
}
```

## Examples

### Simple Enum

```go
type Priority int

const (
    Low Priority = iota
    Medium
    High
)

// Generated usage:
p := High
fmt.Println(p.String()) // "High"

parsed, err := PriorityString("Medium")
// parsed == Medium

values := PriorityValues() // []Priority{Low, Medium, High}
```

### With Line Comments

```go
//go:generate enumer -type=Color -linecomment
type Color int

const (
    ColorRed   Color = iota // red
    ColorGreen              // green
    ColorBlue               // blue
)

// String() returns "red", "green", "blue" instead of "Red", "Green", "Blue"
```

### Flag-Based Enum

```go
type Permission int

const (
    Read    Permission = 1 << iota              // 1
    Write                                       // 2
    Execute                                     // 4
    Admin   Permission = Read | Write | Execute // 7
)

// All values including composite ones work correctly:
p := Admin
fmt.Println(p.String())  // "Admin"
fmt.Println(p.Valid())   // true
fmt.Println(int(p))      // 7

// JSON marshaling preserves the named value:
data, _ := json.Marshal(Admin) // "Admin"
```

## Comparison with alvaroloes/enumer

This implementation differs from alvaroloes/enumer in several ways:

1. **Flag enum support**: Properly handles `1 << iota` and composite values like `Success | Failure`
2. **Template-based**: Uses Go templates for clearer code organization (slight performance trade-off for better maintainability)
3. **Simplified**: Focuses on core functionality without text marshaling or transform methods
4. **Active maintenance**: Built to be maintained and extended

### Migrating from alvaroloes/enumer
Obviously the generated code will be different, but functionality remains the practically the same and you should be able to replace the enumer with little or no other changes required. The only method that might cause issues is validation. In the alvaroloes/enumer, each type had a customised method name based on the type, for example a `RunStatus` type would have a validation method named `IsARunStatus()` whereas we generate the more idiomatic `Valid()`.

## Testing

The repository includes comprehensive tests demonstrating:

- Simple iota-based enums
- Flag-based enums with bit shifts
- Composite enum values
- Enums with gaps in values
- Line comment support
- Prefix trimmimg
- JSON marshaling/unmarshaling
- YAML marshaling/unmarshaling
- SQL Scanner/Valuer implementations
- Flag manipulation methods (Has, Set, Clear, Toggle)

Run tests:

```bash
go test ./...
```

The test runner automatically discovers and runs all test cases in the `testdata/` directory. Each subdirectory represents a test case with:
- `types.go` - enum type definitions
- `types_test.go` - tests for generated methods
- `enumer.args` - command-line arguments for generation

## License

MIT License - see [LICENSE](LICENSE) file for details

## Contributing

Contributions welcome! Please open an issue or pull request.

package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"log"
	"os"
	"os/exec"
	"sort"
	"strings"
	"text/template"

	"golang.org/x/tools/go/packages"
)

var (
	typeNames   = flag.String("type", "", "comma-separated list of type names; must be set")
	output      = flag.String("output", "", "output file name; default is <type>_enumer.go for single type")
	trimPrefix  = flag.String("trimprefix", "", "prefix to be trimmed from the name of each constant")
	lineComment = flag.Bool("linecomment", false, "use line comment text as printed text when present")
	sqlFlag     = flag.Bool("sql", false, "enable SQL Scanner and Valuer interface generation")
	jsonFlag    = flag.Bool("json", false, "enable JSON marshaling methods")
	yamlFlag    = flag.Bool("yaml", false, "enable YAML marshaling methods")
	bitmaskFlag = flag.Bool("bitmask", false, "enable bitmask methods for flag based enums")
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("enumer: ")
	flag.Parse()

	if len(*typeNames) == 0 {
		flag.Usage()
		os.Exit(2)
	}

	types := strings.Split(*typeNames, ",")
	for i := range types {
		types[i] = strings.TrimSpace(types[i])
	}
	sort.Strings(types)

	// Determine output file name
	outputName := *output
	if outputName == "" {
		if len(types) == 1 {
			outputName = fmt.Sprintf("%s_enumer.go", strings.ToLower(types[0]))
		} else if *bitmaskFlag {
			outputName = "flags_gen.go"
		} else {
			outputName = "enums_gen.go"
		}
	}

	// Load the package
	cfg := &packages.Config{
		Mode: packages.NeedTypes | packages.NeedTypesInfo | packages.NeedSyntax | packages.NeedName,
	}
	pkgs, err := packages.Load(cfg, ".")
	if err != nil {
		log.Fatalf("Failed to load package: %v", err)
	}
	if len(pkgs) != 1 {
		log.Fatalf("Expected exactly one package, got %d", len(pkgs))
	}
	pkg := pkgs[0]

	if len(pkg.Errors) > 0 {
		for _, err := range pkg.Errors {
			log.Printf("Package error: %v", err)
		}
		os.Exit(1)
	}

	// Process each type
	allElements := make(map[string][]Element)
	for _, typeName := range types {
		elements, err := processType(pkg, typeName)
		if err != nil {
			log.Fatalf("Failed to process type %s: %v", typeName, err)
		}
		if len(elements) == 0 {
			log.Fatalf("No constants found for type %s", typeName)
		}
		allElements[typeName] = elements
	}

	// Build command string
	cmdStr := buildCommandString(types, outputName)

	// Generate code
	data := TemplateData{
		PackageName: pkg.Name,
		Types:       types,
		Elements:    allElements,
		TrimPrefix:  *trimPrefix,
		SQL:         *sqlFlag,
		JSON:        *jsonFlag,
		YAML:        *yamlFlag,
		Bitmask:     *bitmaskFlag,
		Command:     cmdStr,
	}

	if err := generateCode(outputName, data); err != nil {
		log.Fatalf("Failed to generate code: %v", err)
	}
}

// buildCommandString constructs the command line used to generate the code
func buildCommandString(types []string, outputName string) string {
	var parts []string
	parts = append(parts, "enumer")
	parts = append(parts, fmt.Sprintf("-type=%s", strings.Join(types, ",")))

	if *trimPrefix != "" {
		parts = append(parts, fmt.Sprintf("-trimprefix=%s", *trimPrefix))
	}
	if *lineComment {
		parts = append(parts, "-linecomment")
	}
	if *jsonFlag {
		parts = append(parts, "-json")
	}
	if *yamlFlag {
		parts = append(parts, "-yaml")
	}
	if *sqlFlag {
		parts = append(parts, "-sql")
	}
	if *output != "" {
		parts = append(parts, fmt.Sprintf("-output=%s", *output))
	}
	if *bitmaskFlag {
		parts = append(parts, "-bitmask")
	}

	return strings.Join(parts, " ")
}

// Element represents a single enum constant
type Element struct {
	Name        string
	Value       string
	StringValue string
}

// TemplateData holds all data needed for template execution
type TemplateData struct {
	PackageName string
	Types       []string
	Elements    map[string][]Element
	TrimPrefix  string
	SQL         bool
	JSON        bool
	YAML        bool
	Bitmask     bool
	Command     string
}

// processType extracts all constants for a given type
func processType(pkg *packages.Package, typeName string) ([]Element, error) {
	// Find the type
	obj := pkg.Types.Scope().Lookup(typeName)
	if obj == nil {
		return nil, fmt.Errorf("type %s not found", typeName)
	}

	targetType := obj.Type()
	var elements []Element

	// Iterate through all files in the package
	for _, file := range pkg.Syntax {
		for _, decl := range file.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok || gd.Tok != token.CONST {
				continue
			}

			for _, spec := range gd.Specs {
				vspec := spec.(*ast.ValueSpec)

				for _, name := range vspec.Names {
					// Check if this constant matches our target type
					constObj := pkg.TypesInfo.Defs[name]
					if constObj == nil || !types.Identical(constObj.Type(), targetType) {
						continue
					}

					// Get the constant value
					constValue := constObj.(*types.Const).Val()

					// Get string value (trim prefix if required)
					stringValue := name.Name
					if *trimPrefix != "" {
						stringValue = strings.TrimPrefix(stringValue, *trimPrefix)
					}

					// Override string value with comment if present
					if *lineComment && vspec.Comment != nil {
						comment := strings.TrimSpace(vspec.Comment.Text())
						if comment != "" {
							stringValue = comment
						}
					}

					elements = append(elements, Element{
						Name:        name.Name,
						Value:       constValue.ExactString(),
						StringValue: stringValue,
					})

				}
			}
		}
	}

	return elements, nil
}

// generateCode creates the output file from the template
func generateCode(filename string, data TemplateData) error {
	tmpl, err := template.New("enumer").Funcs(template.FuncMap{
		"lower": strings.ToLower,
	}).Parse(codeTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	// Write to file
	if err := os.WriteFile(filename, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	// Format with gofmt
	cmd := exec.Command("gofmt", "-w", filename)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to format file: %w", err)
	}

	return nil
}

const codeTemplate = `// Code generated by enumer; DO NOT EDIT.
// See: https://github.com/spaceweasel/enumer
// Command: {{.Command}}

package {{.PackageName}}

import (
{{- if .SQL}}
	"database/sql/driver"
{{- end}}
	"fmt"
{{- if .JSON}}
	"encoding/json"
{{- end}}
{{- if .YAML}}
	"gopkg.in/yaml.v3"
{{- end}}
)

{{range $typeName := .Types}}
{{$elements := index $.Elements $typeName}}
{{$trimPrefix := $.TrimPrefix}}

var _{{$typeName}}Map = map[{{$typeName}}]string{
{{- range $elements}}
	{{.Name}}: "{{.StringValue}}",
{{- end}}
}

var _{{$typeName}}Values = []{{$typeName}}{
{{- range $elements}}
	{{.Name}},
{{- end}}
}

var _{{$typeName}}NameToValueMap = map[string]{{$typeName}}{
{{- range $elements}}
	"{{.StringValue}}": {{.Name}},
{{- end}}
}

// String returns the string representation of the {{$typeName}} value
func (i {{$typeName}}) String() string {
	if str, ok := _{{$typeName}}Map[i]; ok {
		return str
	}
	return fmt.Sprintf("{{$typeName}}(%d)", i)
}

// {{$typeName}}Values returns all values of the enum
func {{$typeName}}Values() []{{$typeName}} {
	return _{{$typeName}}Values
}

// {{$typeName}}String retrieves an enum value from the string representation
func {{$typeName}}String(s string) ({{$typeName}}, error) {
	if val, ok := _{{$typeName}}NameToValueMap[s]; ok {
		return val, nil
	}
	return 0, fmt.Errorf("%s is not a valid {{$typeName}}", s)
}

// Valid returns true if the value is a valid {{$typeName}}
func (i {{$typeName}}) Valid() bool {
	_, ok := _{{$typeName}}Map[i]
	return ok
}

{{if $.Bitmask}}
// Has returns true if the flag is set in the {{$typeName}} value
func (i {{$typeName}}) Has(flag {{$typeName}}) bool {
	return i&flag != 0
}

// HasAny returns true if any of the provided flags are set in the {{$typeName}} value
func (i {{$typeName}}) HasAny(flags ...{{$typeName}}) bool {
	for _, flag := range flags {
		if i&flag != 0 {
			return true
		}
	}
	return false
}

// HasAll returns true if all of the provided flags are set in the {{$typeName}} value
func (i {{$typeName}}) HasAll(flags ...{{$typeName}}) bool {
	for _, flag := range flags {
		if i&flag == 0 {
			return false
		}
	}
	return true
}

// Set returns a new {{$typeName}} with the specified flags set
func (i {{$typeName}}) Set(flags ...{{$typeName}}) {{$typeName}} {
	result := i
	for _, flag := range flags {
		result |= flag
	}
	return result
}

// Clear returns a new {{$typeName}} with the specified flags cleared
func (i {{$typeName}}) Clear(flags ...{{$typeName}}) {{$typeName}} {
	result := i
	for _, flag := range flags {
		result &^= flag
	}
	return result
}

// Toggle returns a new {{$typeName}} with the specified flags toggled
func (i {{$typeName}}) Toggle(flags ...{{$typeName}}) {{$typeName}} {
	result := i
	for _, flag := range flags {
		result ^= flag
	}
	return result
}
{{end}}

{{if $.JSON}}
// MarshalJSON implements the json.Marshaler interface for {{$typeName}}
func (i {{$typeName}}) MarshalJSON() ([]byte, error) {
	return json.Marshal(i.String())
}

// UnmarshalJSON implements the json.Unmarshaler interface for {{$typeName}}
func (i *{{$typeName}}) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("{{$typeName}} should be a string, got %s", data)
	}

	var err error
	*i, err = {{$typeName}}String(s)
	return err
}
{{end}}

{{if $.YAML}}
// MarshalYAML implements the yaml.Marshaler interface for {{$typeName}}
func (i {{$typeName}}) MarshalYAML() (any, error) {
	return i.String(), nil
}

// UnmarshalYAML implements the yaml.Unmarshaler interface for {{$typeName}}
func (i *{{$typeName}}) UnmarshalYAML(node *yaml.Node) error {
	var s string
	if err := node.Decode(&s); err != nil {
		return fmt.Errorf("{{$typeName}} should be a string, got %v", node.Value)
	}

	var err error
	*i, err = {{$typeName}}String(s)
	return err
}
{{end}}

{{if $.SQL}}
// Scan implements the sql.Scanner interface for {{$typeName}}
func (i *{{$typeName}}) Scan(value any) error {
	if value == nil {
		return nil
	}

	var s string
	switch v := value.(type) {
	case string:
		s = v
	case []byte:
		s = string(v)
	default:
		return fmt.Errorf("cannot scan type %T into {{$typeName}}", value)
	}

	var err error
	*i, err = {{$typeName}}String(s)
	return err
}

// Value implements the driver.Valuer interface for {{$typeName}}
func (i {{$typeName}}) Value() (driver.Value, error) {
	return i.String(), nil
}
{{end}}

{{end}}
`

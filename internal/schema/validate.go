package schema

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/dlclark/regexp2"
	"github.com/santhosh-tekuri/jsonschema/v6"

	aipkg "github.com/tjespers/aipkg"
)

var (
	compiledSchema *jsonschema.Schema
	compileOnce    sync.Once
	compileErr     error

	compiledProjectSchema *jsonschema.Schema
	compileProjectOnce    sync.Once
	compileProjectErr     error
)

// regexp2Regexp adapts regexp2.Regexp to the jsonschema.Regexp interface.
type regexp2Regexp regexp2.Regexp

// MatchString implements jsonschema.Regexp using the regexp2 engine.
func (re *regexp2Regexp) MatchString(s string) bool {
	matched, err := (*regexp2.Regexp)(re).MatchString(s)
	return err == nil && matched
}

func (re *regexp2Regexp) String() string {
	return (*regexp2.Regexp)(re).String()
}

func regexp2Compile(s string) (jsonschema.Regexp, error) {
	re, err := regexp2.Compile(s, regexp2.ECMAScript)
	if err != nil {
		return nil, err
	}
	return (*regexp2Regexp)(re), nil
}

func compiled() (*jsonschema.Schema, error) {
	compileOnce.Do(func() {
		doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(aipkg.PackageSchemaJSON))
		if err != nil {
			compileErr = fmt.Errorf("unmarshaling package schema: %w", err)
			return
		}
		c := jsonschema.NewCompiler()
		c.DefaultDraft(jsonschema.Draft2020)
		c.UseRegexpEngine(regexp2Compile)
		if err := c.AddResource("package.json", doc); err != nil {
			compileErr = fmt.Errorf("adding schema resource: %w", err)
			return
		}
		compiledSchema, compileErr = c.Compile("package.json")
	})
	return compiledSchema, compileErr
}

func compiledProject() (*jsonschema.Schema, error) {
	compileProjectOnce.Do(func() {
		doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(aipkg.ProjectSchemaJSON))
		if err != nil {
			compileProjectErr = fmt.Errorf("unmarshaling project schema: %w", err)
			return
		}
		c := jsonschema.NewCompiler()
		c.DefaultDraft(jsonschema.Draft2020)
		c.UseRegexpEngine(regexp2Compile)
		if err := c.AddResource("project.json", doc); err != nil {
			compileProjectErr = fmt.Errorf("adding project schema resource: %w", err)
			return
		}
		compiledProjectSchema, compileProjectErr = c.Compile("project.json")
	})
	return compiledProjectSchema, compileProjectErr
}

// Validate validates a JSON-encoded manifest against the package schema.
// The input should be the raw JSON bytes of an aipkg.json file.
func Validate(jsonData []byte) error {
	root, err := compiled()
	if err != nil {
		return fmt.Errorf("compiling schema: %w", err)
	}

	doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("parsing manifest JSON: %w", err)
	}

	return root.Validate(doc)
}

// ValidateProject validates a JSON-encoded project file against the project schema.
// The input should be the raw JSON bytes of an aipkg-project.json file.
func ValidateProject(jsonData []byte) error {
	root, err := compiledProject()
	if err != nil {
		return fmt.Errorf("compiling project schema: %w", err)
	}

	doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("parsing project JSON: %w", err)
	}

	return root.Validate(doc)
}

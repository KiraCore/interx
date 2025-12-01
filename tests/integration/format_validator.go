package integration

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

// FieldType represents expected JSON types
type FieldType string

const (
	TypeString  FieldType = "string"
	TypeInt     FieldType = "int"
	TypeFloat   FieldType = "float"
	TypeBool    FieldType = "bool"
	TypeArray   FieldType = "array"
	TypeObject  FieldType = "object"
	TypeAny     FieldType = "any"
)

// FieldSpec defines expected field properties
type FieldSpec struct {
	Name     string
	Type     FieldType
	Required bool
}

// ValidationResult contains test results for reporting
type ValidationResult struct {
	Field    string
	Issue    string
	IssueRef string // GitHub issue reference if known
}

// FormatValidator validates JSON responses against expected formats
type FormatValidator struct {
	t       *testing.T
	rawJSON map[string]interface{}
	results []ValidationResult
}

// NewFormatValidator creates a validator from JSON bytes
func NewFormatValidator(t *testing.T, body []byte) (*FormatValidator, error) {
	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}
	return &FormatValidator{
		t:       t,
		rawJSON: raw,
		results: make([]ValidationResult, 0),
	}, nil
}

// GetRaw returns the raw JSON map
func (v *FormatValidator) GetRaw() map[string]interface{} {
	return v.rawJSON
}

// HasField checks if a field exists
func (v *FormatValidator) HasField(name string) bool {
	_, exists := v.rawJSON[name]
	return exists
}

// GetField returns a field value
func (v *FormatValidator) GetField(name string) (interface{}, bool) {
	val, exists := v.rawJSON[name]
	return val, exists
}

// ValidateFieldExists checks field existence and reports issues
func (v *FormatValidator) ValidateFieldExists(name string, issueRef string) bool {
	if !v.HasField(name) {
		v.t.Errorf("%s: Missing expected field '%s'", issueRef, name)
		v.results = append(v.results, ValidationResult{
			Field:    name,
			Issue:    "missing field",
			IssueRef: issueRef,
		})
		return false
	}
	return true
}

// ValidateFieldNotExists checks that a field does NOT exist (for detecting wrong naming)
func (v *FormatValidator) ValidateFieldNotExists(wrongName, correctName, issueRef string) {
	if v.HasField(wrongName) && !v.HasField(correctName) {
		v.t.Errorf("%s: Using '%s' instead of expected '%s'", issueRef, wrongName, correctName)
		v.results = append(v.results, ValidationResult{
			Field:    wrongName,
			Issue:    fmt.Sprintf("wrong field name, should be '%s'", correctName),
			IssueRef: issueRef,
		})
	}
}

// ValidateFieldType checks if a field has the expected type
func (v *FormatValidator) ValidateFieldType(name string, expectedType FieldType, issueRef string) bool {
	val, exists := v.rawJSON[name]
	if !exists {
		return false
	}

	var actualType FieldType
	switch val.(type) {
	case string:
		actualType = TypeString
	case float64:
		// JSON numbers are float64, check if it's an integer
		f := val.(float64)
		if f == float64(int64(f)) {
			actualType = TypeInt
		} else {
			actualType = TypeFloat
		}
	case bool:
		actualType = TypeBool
	case []interface{}:
		actualType = TypeArray
	case map[string]interface{}:
		actualType = TypeObject
	case nil:
		return true // nil is acceptable for optional fields
	default:
		actualType = TypeAny
	}

	// Special case: int expected but got float that's actually integer
	if expectedType == TypeInt && actualType == TypeFloat {
		f := val.(float64)
		if f == float64(int64(f)) {
			actualType = TypeInt
		}
	}

	// Special case: expecting string but got number (common Issue #16)
	if expectedType == TypeString && (actualType == TypeInt || actualType == TypeFloat) {
		v.t.Errorf("%s: Field '%s' should be string, got number", issueRef, name)
		v.results = append(v.results, ValidationResult{
			Field:    name,
			Issue:    fmt.Sprintf("expected string, got %s", actualType),
			IssueRef: issueRef,
		})
		return false
	}

	// Special case: expecting int but got string (common Issue #16)
	if expectedType == TypeInt && actualType == TypeString {
		v.t.Errorf("%s: Field '%s' should be int, got string", issueRef, name)
		v.results = append(v.results, ValidationResult{
			Field:    name,
			Issue:    "expected int, got string",
			IssueRef: issueRef,
		})
		return false
	}

	if expectedType != TypeAny && actualType != expectedType {
		v.t.Errorf("%s: Field '%s' expected type %s, got %s", issueRef, name, expectedType, actualType)
		v.results = append(v.results, ValidationResult{
			Field:    name,
			Issue:    fmt.Sprintf("expected %s, got %s", expectedType, actualType),
			IssueRef: issueRef,
		})
		return false
	}

	return true
}

// ValidateSnakeCase checks all top-level field names use snake_case
func (v *FormatValidator) ValidateSnakeCase(issueRef string) {
	for key := range v.rawJSON {
		if key == "@type" {
			continue // Special case for protobuf type field
		}
		if isCamelCase(key) {
			v.t.Errorf("%s: Field '%s' uses camelCase, expected snake_case", issueRef, key)
			v.results = append(v.results, ValidationResult{
				Field:    key,
				Issue:    "uses camelCase instead of snake_case",
				IssueRef: issueRef,
			})
		}
	}
}

// ValidateNestedSnakeCase checks nested object field names
func (v *FormatValidator) ValidateNestedSnakeCase(fieldName, issueRef string) {
	val, exists := v.rawJSON[fieldName]
	if !exists {
		return
	}

	nested, ok := val.(map[string]interface{})
	if !ok {
		return
	}

	for key := range nested {
		if key == "@type" {
			continue
		}
		if isCamelCase(key) {
			v.t.Errorf("%s: Nested field '%s.%s' uses camelCase, expected snake_case", issueRef, fieldName, key)
			v.results = append(v.results, ValidationResult{
				Field:    fmt.Sprintf("%s.%s", fieldName, key),
				Issue:    "uses camelCase instead of snake_case",
				IssueRef: issueRef,
			})
		}
	}
}

// ValidateArrayFieldType validates types within an array
func (v *FormatValidator) ValidateArrayFieldType(arrayName, fieldName string, expectedType FieldType, issueRef string) {
	val, exists := v.rawJSON[arrayName]
	if !exists {
		return
	}

	arr, ok := val.([]interface{})
	if !ok || len(arr) == 0 {
		return
	}

	for i, item := range arr {
		obj, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		fieldVal, exists := obj[fieldName]
		if !exists {
			v.t.Errorf("%s: Array '%s[%d]' missing field '%s'", issueRef, arrayName, i, fieldName)
			continue
		}

		var actualType FieldType
		switch fieldVal.(type) {
		case string:
			actualType = TypeString
		case float64:
			f := fieldVal.(float64)
			if f == float64(int64(f)) {
				actualType = TypeInt
			} else {
				actualType = TypeFloat
			}
		case bool:
			actualType = TypeBool
		}

		if expectedType == TypeString && (actualType == TypeInt || actualType == TypeFloat) {
			v.t.Errorf("%s: Array '%s[%d].%s' should be string, got number", issueRef, arrayName, i, fieldName)
			return // Only report first occurrence
		}

		if expectedType == TypeInt && actualType == TypeString {
			v.t.Errorf("%s: Array '%s[%d].%s' should be int, got string", issueRef, arrayName, i, fieldName)
			return
		}
	}
}

// GetResults returns all validation results
func (v *FormatValidator) GetResults() []ValidationResult {
	return v.results
}

// HasIssues returns true if any validation issues were found
func (v *FormatValidator) HasIssues() bool {
	return len(v.results) > 0
}

// isCamelCase checks if a string uses camelCase (lowercase start, has uppercase)
func isCamelCase(s string) bool {
	if len(s) == 0 {
		return false
	}
	// Skip if starts with uppercase (PascalCase)
	if s[0] >= 'A' && s[0] <= 'Z' {
		return false
	}
	// Check for uppercase letters after the first character
	for i := 1; i < len(s); i++ {
		if s[i] >= 'A' && s[i] <= 'Z' {
			return true
		}
	}
	return false
}

// toSnakeCase converts camelCase to snake_case
func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				result.WriteRune('_')
			}
			result.WriteRune(r + 32) // lowercase
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// ValidateRequiredFields validates multiple required fields exist
func (v *FormatValidator) ValidateRequiredFields(fields []string, issueRef string) {
	for _, field := range fields {
		v.ValidateFieldExists(field, issueRef)
	}
}

// ValidateFieldTypes validates multiple field types
func (v *FormatValidator) ValidateFieldTypes(specs []FieldSpec, issueRef string) {
	for _, spec := range specs {
		if spec.Required {
			if !v.ValidateFieldExists(spec.Name, issueRef) {
				continue
			}
		}
		if v.HasField(spec.Name) {
			v.ValidateFieldType(spec.Name, spec.Type, issueRef)
		}
	}
}

// GetNestedValidator returns a validator for a nested object
func (v *FormatValidator) GetNestedValidator(fieldName string) (*FormatValidator, bool) {
	val, exists := v.rawJSON[fieldName]
	if !exists {
		return nil, false
	}

	nested, ok := val.(map[string]interface{})
	if !ok {
		return nil, false
	}

	return &FormatValidator{
		t:       v.t,
		rawJSON: nested,
		results: v.results, // Share results
	}, true
}

// GetArrayLength returns the length of an array field
func (v *FormatValidator) GetArrayLength(fieldName string) int {
	val, exists := v.rawJSON[fieldName]
	if !exists {
		return 0
	}

	arr, ok := val.([]interface{})
	if !ok {
		return 0
	}

	return len(arr)
}

// ValidateTimestampIsInt checks if a timestamp field is an integer (not string, not ISO)
func (v *FormatValidator) ValidateTimestampIsInt(fieldName, issueRef string) bool {
	val, exists := v.rawJSON[fieldName]
	if !exists {
		v.t.Errorf("%s: Missing timestamp field '%s'", issueRef, fieldName)
		return false
	}

	switch val.(type) {
	case float64:
		return true // JSON numbers are float64, this is correct for int
	case string:
		v.t.Errorf("%s: Timestamp '%s' should be Unix int, got string (possibly ISO format)", issueRef, fieldName)
		return false
	default:
		v.t.Errorf("%s: Timestamp '%s' has unexpected type %T", issueRef, fieldName, val)
		return false
	}
}

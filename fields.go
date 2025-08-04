package simsdk

import (
	"fmt"
	"strings"
)

type FieldType string

const (
	FieldString    FieldType = "string"    // A plain string
	FieldInt       FieldType = "int"       // A signed integer
	FieldUint      FieldType = "uint"      // An unsigned integer
	FieldFloat     FieldType = "float"     // A floating point number
	FieldBool      FieldType = "bool"      // A boolean true/false value
	FieldEnum      FieldType = "enum"      // A fixed set of named values
	FieldTimestamp FieldType = "timestamp" // A time value (e.g., RFC3339)
	FieldRepeated  FieldType = "repeated"  // A repeated field (meta-type)
	FieldObject    FieldType = "object"    // A nested structure
)

// AllFieldTypes is a list of valid field types for lookup or documentation
var AllFieldTypes = []FieldType{
	FieldString,
	FieldInt,
	FieldUint,
	FieldFloat,
	FieldBool,
	FieldEnum,
	FieldTimestamp,
	FieldRepeated,
	FieldObject,
}

// IsValid returns true if the FieldType is one of the known constants.
func (f FieldType) IsValid() bool {
	for _, v := range AllFieldTypes {
		if f == v {
			return true
		}
	}
	return false
}

// String returns the FieldType as a string.
func (f FieldType) String() string {
	return string(f)
}

// ParseFieldType returns a valid FieldType from a string.
// It is case-insensitive and returns an error if the type is unrecognized.
func ParseFieldType(s string) (FieldType, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	for _, ft := range AllFieldTypes {
		if s == string(ft) {
			return ft, nil
		}
	}
	return "", fmt.Errorf("unrecognized field type: %q", s)
}

// MustValidate panics if the FieldType is invalid. Use for strict assertions.
func (f FieldType) MustValidate() {
	if !f.IsValid() {
		panic(fmt.Sprintf("invalid FieldType: %s", f))
	}
}

// MustParseFieldType is like ParseFieldType but panics on error.
func MustParseFieldType(s string) FieldType {
	ft, err := ParseFieldType(s)
	if err != nil {
		panic(err)
	}
	return ft
}

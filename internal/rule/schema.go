package rule

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

// Schema defines the interface for options validators. It supports validating
// raw configuration maps/slices, applying default values, and exporting
// TypeScript declarations for the generated TypeScript type definitions.
type Schema interface {
	Validate(raw any) (any, error)
	TSType() string
}

// AnySchema matches any value
type AnySchema struct {
	def any
}

func Any() *AnySchema {
	return &AnySchema{}
}

func (s *AnySchema) Default(def any) *AnySchema {
	s.def = def
	return s
}

func (s *AnySchema) Validate(raw any) (any, error) {
	if raw == nil {
		return s.def, nil
	}
	return raw, nil
}

func (s *AnySchema) TSType() string {
	return "any"
}

// BoolSchema validates boolean values
type BoolSchema struct {
	def *bool
}

func Bool() *BoolSchema {
	return &BoolSchema{}
}

func (s *BoolSchema) Default(def bool) *BoolSchema {
	s.def = &def
	return s
}

func (s *BoolSchema) Validate(raw any) (any, error) {
	if raw == nil {
		if s.def != nil {
			return *s.def, nil
		}
		return false, nil
	}
	b, ok := raw.(bool)
	if !ok {
		return nil, fmt.Errorf("expected bool, got %T", raw)
	}
	return b, nil
}

func (s *BoolSchema) TSType() string {
	return "boolean"
}

// IntSchema validates integer values
type IntSchema struct {
	def *int
	min *int
	max *int
}

func Int() *IntSchema {
	return &IntSchema{}
}

func (s *IntSchema) Default(def int) *IntSchema {
	s.def = &def
	return s
}

func (s *IntSchema) Min(minVal int) *IntSchema {
	s.min = &minVal
	return s
}

func (s *IntSchema) Max(maxVal int) *IntSchema {
	s.max = &maxVal
	return s
}

func (s *IntSchema) Validate(raw any) (any, error) {
	if raw == nil {
		if s.def != nil {
			return *s.def, nil
		}
		return 0, nil
	}

	var val int
	switch n := raw.(type) {
	case int:
		val = n
	case float64:
		val = int(n)
	default:
		return nil, fmt.Errorf("expected int, got %T", raw)
	}

	if s.min != nil && val < *s.min {
		return nil, fmt.Errorf("value %d is less than min %d", val, *s.min)
	}
	if s.max != nil && val > *s.max {
		return nil, fmt.Errorf("value %d is greater than max %d", val, *s.max)
	}
	return val, nil
}

func (s *IntSchema) TSType() string {
	return "number"
}

// StringSchema validates string values
type StringSchema struct {
	def *string
}

func String() *StringSchema {
	return &StringSchema{}
}

func (s *StringSchema) Default(def string) *StringSchema {
	s.def = &def
	return s
}

func (s *StringSchema) Validate(raw any) (any, error) {
	if raw == nil {
		if s.def != nil {
			return *s.def, nil
		}
		return "", nil
	}
	str, ok := raw.(string)
	if !ok {
		return nil, fmt.Errorf("expected string, got %T", raw)
	}
	return str, nil
}

func (s *StringSchema) TSType() string {
	return "string"
}

// EnumSchema validates string values matching a set of options
type EnumSchema struct {
	allowed []string
	def     *string
}

func Enum(allowed ...string) *EnumSchema {
	return &EnumSchema{allowed: allowed}
}

func (s *EnumSchema) Default(def string) *EnumSchema {
	s.def = &def
	return s
}

func (s *EnumSchema) Validate(raw any) (any, error) {
	if raw == nil {
		if s.def != nil {
			return *s.def, nil
		}
		return "", nil
	}
	str, ok := raw.(string)
	if !ok {
		return nil, fmt.Errorf("expected string, got %T", raw)
	}
	for _, a := range s.allowed {
		if a == str {
			return str, nil
		}
	}
	return nil, fmt.Errorf("expected one of %v, got %q", s.allowed, str)
}

func (s *EnumSchema) TSType() string {
	var parts []string
	for _, a := range s.allowed {
		parts = append(parts, fmt.Sprintf("%q", a))
	}
	return strings.Join(parts, " | ")
}

// ArraySchema validates slices of values
type ArraySchema struct {
	item Schema
}

func Array(item Schema) *ArraySchema {
	return &ArraySchema{item: item}
}

func (s *ArraySchema) Validate(raw any) (any, error) {
	if raw == nil {
		return []any{}, nil
	}
	var arr []any
	val := reflect.ValueOf(raw)
	if val.Kind() == reflect.Slice {
		arr = make([]any, val.Len())
		for i := range val.Len() {
			arr[i] = val.Index(i).Interface()
		}
	} else {
		return nil, fmt.Errorf("expected slice, got %T", raw)
	}

	res := make([]any, len(arr))
	for i, item := range arr {
		v, err := s.item.Validate(item)
		if err != nil {
			return nil, fmt.Errorf("at index %d: %w", i, err)
		}
		res[i] = v
	}
	return res, nil
}

func (s *ArraySchema) TSType() string {
	return fmt.Sprintf("Array<%s>", s.item.TSType())
}

// ObjectSchema validates map of key-value pairs
type ObjectSchema struct {
	properties map[string]Schema
}

func Object(properties map[string]Schema) *ObjectSchema {
	return &ObjectSchema{properties: properties}
}

func (s *ObjectSchema) Validate(raw any) (any, error) {
	var m map[string]any
	if raw == nil {
		m = make(map[string]any)
	} else {
		val := reflect.ValueOf(raw)
		if val.Kind() == reflect.Map {
			m = make(map[string]any)
			for _, k := range val.MapKeys() {
				if keyStr, ok := k.Interface().(string); ok {
					m[keyStr] = val.MapIndex(k).Interface()
				}
			}
		} else {
			return nil, fmt.Errorf("expected map, got %T", raw)
		}
	}

	res := make(map[string]any)
	for k, propSchema := range s.properties {
		val := m[k]
		v, err := propSchema.Validate(val)
		if err != nil {
			return nil, fmt.Errorf("at property %q: %w", k, err)
		}
		res[k] = v
	}
	return res, nil
}

func (s *ObjectSchema) TSType() string {
	var parts []string
	for k, v := range s.properties {
		parts = append(parts, fmt.Sprintf("%s?: %s", k, v.TSType()))
	}
	sort.Strings(parts)
	if len(parts) == 0 {
		return "Record<string, any>"
	}
	return fmt.Sprintf("{ %s }", strings.Join(parts, "; "))
}

// TupleSchema validates ordered element shapes in a slice
type TupleSchema struct {
	items []Schema
}

func Tuple(items ...Schema) *TupleSchema {
	return &TupleSchema{items: items}
}

func (s *TupleSchema) Validate(raw any) (any, error) {
	if raw == nil {
		raw = []any{}
	}
	var arr []any
	val := reflect.ValueOf(raw)
	if val.Kind() == reflect.Slice {
		arr = make([]any, val.Len())
		for i := range val.Len() {
			arr[i] = val.Index(i).Interface()
		}
	} else {
		// Single element wrapped in array
		arr = []any{raw}
	}

	res := make([]any, len(s.items))
	for i, schema := range s.items {
		var item any
		if i < len(arr) {
			item = arr[i]
		}
		v, err := schema.Validate(item)
		if err != nil {
			return nil, fmt.Errorf("at element %d: %w", i, err)
		}
		res[i] = v
	}
	return res, nil
}

func (s *TupleSchema) TSType() string {
	var parts []string
	for _, item := range s.items {
		parts = append(parts, item.TSType())
	}
	return fmt.Sprintf("[%s]", strings.Join(parts, ", "))
}

// UnionSchema validates raw input against multiple alternatives
type UnionSchema struct {
	schemas []Schema
	def     any
}

func Union(schemas ...Schema) *UnionSchema {
	return &UnionSchema{schemas: schemas}
}

func (s *UnionSchema) Default(def any) *UnionSchema {
	s.def = def
	return s
}

func (s *UnionSchema) Validate(raw any) (any, error) {
	if raw == nil {
		return s.def, nil
	}
	var errors []error
	for _, schema := range s.schemas {
		v, err := schema.Validate(raw)
		if err == nil {
			return v, nil
		}
		errors = append(errors, err)
	}
	return nil, fmt.Errorf("union failed: %v", errors)
}

func (s *UnionSchema) TSType() string {
	var parts []string
	for _, schema := range s.schemas {
		parts = append(parts, schema.TSType())
	}
	return strings.Join(parts, " | ")
}

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
	IsOptional() bool
}

// DefaultSchema wraps an inner schema and returns a default value if the input is nil.
type DefaultSchema struct {
	inner Schema
	def   any
}

// Default wraps a schema with a default value.
func Default(inner Schema, def any) *DefaultSchema {
	return &DefaultSchema{
		inner: inner,
		def:   def,
	}
}

func (s *DefaultSchema) Validate(raw any) (any, error) {
	if raw == nil {
		if s.def == nil {
			return nil, nil
		}
		return s.inner.Validate(s.def)
	}
	return s.inner.Validate(raw)
}

func (s *DefaultSchema) TSType() string {
	return s.inner.TSType()
}

func (s *DefaultSchema) IsOptional() bool {
	return true
}

// AnySchema matches any value
type AnySchema struct{}

func Any() *AnySchema {
	return &AnySchema{}
}

func (s *AnySchema) Default(def any) *DefaultSchema {
	return Default(s, def)
}

func (s *AnySchema) Validate(raw any) (any, error) {
	return raw, nil
}

func (s *AnySchema) TSType() string {
	return "any"
}

func (s *AnySchema) IsOptional() bool {
	return true
}

// BoolSchema validates boolean values
type BoolSchema struct{}

func Bool() *BoolSchema {
	return &BoolSchema{}
}

func (s *BoolSchema) Default(def bool) *DefaultSchema {
	return Default(s, def)
}

func (s *BoolSchema) Validate(raw any) (any, error) {
	if raw == nil {
		return nil, fmt.Errorf("expected bool, got nil")
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

func (s *BoolSchema) IsOptional() bool {
	return false
}

// IntSchema validates integer values
type IntSchema struct {
	min *int
	max *int
}

func Int() *IntSchema {
	return &IntSchema{}
}

func (s *IntSchema) Default(def int) *DefaultSchema {
	return Default(s, def)
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
		return nil, fmt.Errorf("expected int, got nil")
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

func (s *IntSchema) IsOptional() bool {
	return false
}

// StringSchema validates string values
type StringSchema struct{}

func String() *StringSchema {
	return &StringSchema{}
}

func (s *StringSchema) Default(def string) *DefaultSchema {
	return Default(s, def)
}

func (s *StringSchema) Validate(raw any) (any, error) {
	if raw == nil {
		return nil, fmt.Errorf("expected string, got nil")
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

func (s *StringSchema) IsOptional() bool {
	return false
}

// EnumSchema validates string values matching a set of options
type EnumSchema struct {
	allowed []string
}

func Enum(allowed ...string) *EnumSchema {
	return &EnumSchema{allowed: allowed}
}

func (s *EnumSchema) Default(def string) *DefaultSchema {
	return Default(s, def)
}

func (s *EnumSchema) Validate(raw any) (any, error) {
	if raw == nil {
		return nil, fmt.Errorf("expected string, got nil")
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
	return "(" + strings.Join(parts, " | ") + ")"
}

func (s *EnumSchema) IsOptional() bool {
	return false
}

// ArraySchema validates slices of values
type ArraySchema struct {
	item   Schema
	minLen *int
	maxLen *int
}

func Array(item Schema) *ArraySchema {
	return &ArraySchema{item: item}
}

func (s *ArraySchema) Default(def []any) *DefaultSchema {
	return Default(s, def)
}

func (s *ArraySchema) MinLen(n int) *ArraySchema {
	s.minLen = &n
	return s
}

func (s *ArraySchema) MaxLen(n int) *ArraySchema {
	s.maxLen = &n
	return s
}

func (s *ArraySchema) Len(n int) *ArraySchema {
	s.minLen = &n
	s.maxLen = &n
	return s
}

func (s *ArraySchema) Validate(raw any) (any, error) {
	if raw == nil {
		if s.IsOptional() {
			return []any{}, nil
		}
		return nil, fmt.Errorf("expected slice, got nil")
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

	if s.minLen != nil && len(arr) < *s.minLen {
		return nil, fmt.Errorf("array length %d is less than minimum %d", len(arr), *s.minLen)
	}
	if s.maxLen != nil && len(arr) > *s.maxLen {
		return nil, fmt.Errorf("array length %d is greater than maximum %d", len(arr), *s.maxLen)
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

func (s *ArraySchema) IsOptional() bool {
	return false
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
		suffix := ""
		if v.IsOptional() {
			suffix = "?"
		}
		parts = append(parts, fmt.Sprintf("%s%s: %s", k, suffix, v.TSType()))
	}
	sort.Strings(parts)
	if len(parts) == 0 {
		return "Record<string, any>"
	}
	return fmt.Sprintf("{ %s }", strings.Join(parts, "; "))
}

func (s *ObjectSchema) IsOptional() bool {
	for _, propSchema := range s.properties {
		if !propSchema.IsOptional() {
			return false
		}
	}
	return true
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
		return nil, fmt.Errorf("expected slice, got %T", raw)
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
		suffix := ""
		if item.IsOptional() {
			suffix = "?"
		}
		parts = append(parts, item.TSType()+suffix)
	}
	return fmt.Sprintf("[%s]", strings.Join(parts, ", "))
}

func (s *TupleSchema) IsOptional() bool {
	if len(s.items) == 0 {
		return true
	}
	for _, item := range s.items {
		if !item.IsOptional() {
			return false
		}
	}
	return true
}

// UnionSchema validates raw input against multiple alternatives
type UnionSchema struct {
	schemas []Schema
}

func Union(schemas ...Schema) *UnionSchema {
	return &UnionSchema{schemas: schemas}
}

func (s *UnionSchema) Default(def any) *DefaultSchema {
	return Default(s, def)
}

func (s *UnionSchema) Validate(raw any) (any, error) {
	if raw == nil {
		return nil, fmt.Errorf("expected non-nil value, got nil")
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
	return "(" + strings.Join(parts, " | ") + ")"
}

func (s *UnionSchema) IsOptional() bool {
	return false
}

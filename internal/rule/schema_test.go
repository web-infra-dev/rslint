package rule

import (
	"reflect"
	"strings"
	"testing"
)

func TestAnySchema(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		s := Any().Default("fallback")
		val, err := s.Validate(nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if val != "fallback" {
			t.Errorf("expected 'fallback', got %v", val)
		}
	})

	t.Run("Validate", func(t *testing.T) {
		s := Any()
		val, err := s.Validate("hello")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if val != "hello" {
			t.Errorf("expected 'hello', got %v", val)
		}

		val, err = s.Validate(42)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if val != 42 {
			t.Errorf("expected 42, got %v", val)
		}
	})

	t.Run("TSType", func(t *testing.T) {
		s := Any()
		if s.TSType() != "any" {
			t.Errorf("expected 'any', got %q", s.TSType())
		}
	})
}

func TestBoolSchema(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		s := Bool()
		_, err := s.Validate(nil)
		if err == nil {
			t.Error("expected error for nil when not wrapped")
		}

		sDef := Bool().Default(true)
		val, err := sDef.Validate(nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if val != true {
			t.Errorf("expected true, got %v", val)
		}
	})

	t.Run("Validate", func(t *testing.T) {
		s := Bool()
		val, err := s.Validate(true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if val != true {
			t.Errorf("expected true, got %v", val)
		}

		_, err = s.Validate("not a bool")
		if err == nil {
			t.Error("expected error for non-bool input")
		}
	})

	t.Run("TSType", func(t *testing.T) {
		s := Bool()
		if s.TSType() != "boolean" {
			t.Errorf("expected 'boolean', got %q", s.TSType())
		}
	})
}

func TestIntSchema(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		s := Int()
		_, err := s.Validate(nil)
		if err == nil {
			t.Error("expected error for nil when not wrapped")
		}

		sDef := Int().Default(42)
		val, err := sDef.Validate(nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if val != 42 {
			t.Errorf("expected 42, got %v", val)
		}
	})

	t.Run("Validate", func(t *testing.T) {
		s := Int()
		val, err := s.Validate(10)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if val != 10 {
			t.Errorf("expected 10, got %v", val)
		}

		// float64 conversion
		val, err = s.Validate(12.34)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if val != 12 {
			t.Errorf("expected 12, got %v", val)
		}

		_, err = s.Validate("not an int")
		if err == nil {
			t.Error("expected error for non-int input")
		}
	})

	t.Run("MinMaxConstraints", func(t *testing.T) {
		s := Int().Min(5).Max(15)

		_, err := s.Validate(4)
		if err == nil || !strings.Contains(err.Error(), "less than min") {
			t.Errorf("expected less than min error, got %v", err)
		}

		val, err := s.Validate(5)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if val != 5 {
			t.Errorf("expected 5, got %v", val)
		}

		val, err = s.Validate(15)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if val != 15 {
			t.Errorf("expected 15, got %v", val)
		}

		_, err = s.Validate(16)
		if err == nil || !strings.Contains(err.Error(), "greater than max") {
			t.Errorf("expected greater than max error, got %v", err)
		}
	})

	t.Run("MinOnly", func(t *testing.T) {
		s := Int().Min(0)

		_, err := s.Validate(-1)
		if err == nil || !strings.Contains(err.Error(), "less than min") {
			t.Errorf("expected less than min error, got %v", err)
		}

		val, err := s.Validate(0)
		if err != nil {
			t.Fatalf("unexpected error at boundary: %v", err)
		}
		if val != 0 {
			t.Errorf("expected 0, got %v", val)
		}

		val, err = s.Validate(9999)
		if err != nil {
			t.Fatalf("unexpected error above min: %v", err)
		}
		if val != 9999 {
			t.Errorf("expected 9999, got %v", val)
		}
	})

	t.Run("MaxOnly", func(t *testing.T) {
		s := Int().Max(100)

		val, err := s.Validate(100)
		if err != nil {
			t.Fatalf("unexpected error at boundary: %v", err)
		}
		if val != 100 {
			t.Errorf("expected 100, got %v", val)
		}

		_, err = s.Validate(101)
		if err == nil || !strings.Contains(err.Error(), "greater than max") {
			t.Errorf("expected greater than max error, got %v", err)
		}
	})

	t.Run("TSType", func(t *testing.T) {
		s := Int()
		if s.TSType() != "number" {
			t.Errorf("expected 'number', got %q", s.TSType())
		}
	})
}

func TestStringSchema(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		s := String()
		_, err := s.Validate(nil)
		if err == nil {
			t.Error("expected error for nil when not wrapped")
		}

		sDef := String().Default("foo")
		val, err := sDef.Validate(nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if val != "foo" {
			t.Errorf("expected 'foo', got %q", val)
		}
	})

	t.Run("Validate", func(t *testing.T) {
		s := String()
		val, err := s.Validate("bar")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if val != "bar" {
			t.Errorf("expected 'bar', got %q", val)
		}

		_, err = s.Validate(123)
		if err == nil {
			t.Error("expected error for non-string input")
		}
	})

	t.Run("TSType", func(t *testing.T) {
		s := String()
		if s.TSType() != "string" {
			t.Errorf("expected 'string', got %q", s.TSType())
		}
	})
}

func TestEnumSchema(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		s := Enum("a", "b")
		_, err := s.Validate(nil)
		if err == nil {
			t.Error("expected error for nil when not wrapped")
		}

		sDef := Enum("a", "b").Default("b")
		val, err := sDef.Validate(nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if val != "b" {
			t.Errorf("expected 'b', got %q", val)
		}
	})

	t.Run("Validate", func(t *testing.T) {
		s := Enum("a", "b")
		val, err := s.Validate("a")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if val != "a" {
			t.Errorf("expected 'a', got %q", val)
		}

		_, err = s.Validate("c")
		if err == nil || !strings.Contains(err.Error(), "expected one of") {
			t.Errorf("expected one of error, got %v", err)
		}

		_, err = s.Validate(true)
		if err == nil {
			t.Error("expected error for non-string enum value")
		}
	})

	t.Run("TSType", func(t *testing.T) {
		s := Enum("a", "b")
		expected := `("a" | "b")`
		if s.TSType() != expected {
			t.Errorf("expected %q, got %q", expected, s.TSType())
		}
	})
}

func TestArraySchema(t *testing.T) {
	t.Run("NilReturnsEmpty", func(t *testing.T) {
		s := Array(String())
		val, err := s.Validate(nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := []any{}
		if !reflect.DeepEqual(val, expected) {
			t.Errorf("expected %v, got %v", expected, val)
		}
	})

	t.Run("Default", func(t *testing.T) {
		// Custom default returned on nil input
		s := Array(String()).Default([]any{"a", "b"})
		val, err := s.Validate(nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !reflect.DeepEqual(val, []any{"a", "b"}) {
			t.Errorf("expected [a b], got %v", val)
		}

		// Explicit value overrides default
		val, err = s.Validate([]any{"c"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !reflect.DeepEqual(val, []any{"c"}) {
			t.Errorf("expected [c], got %v", val)
		}
	})

	t.Run("Validate", func(t *testing.T) {
		s := Array(Int())
		input := []any{1, 2.0, 3}
		val, err := s.Validate(input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := []any{1, 2, 3}
		if !reflect.DeepEqual(val, expected) {
			t.Errorf("expected %v, got %v", expected, val)
		}

		// typed slice input
		typedInput := []int{4, 5, 6}
		val, err = s.Validate(typedInput)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expectedTyped := []any{4, 5, 6}
		if !reflect.DeepEqual(val, expectedTyped) {
			t.Errorf("expected %v, got %v", expectedTyped, val)
		}

		// item failure
		badInput := []any{1, "not-an-int", 3}
		_, err = s.Validate(badInput)
		if err == nil || !strings.Contains(err.Error(), "at index 1") {
			t.Errorf("expected at index 1 validation error, got %v", err)
		}

		// non-slice input
		_, err = s.Validate("not-a-slice")
		if err == nil || !strings.Contains(err.Error(), "expected slice") {
			t.Errorf("expected slice error, got %v", err)
		}
	})

	t.Run("MinLen", func(t *testing.T) {
		s := Array(Int()).MinLen(2)

		// nil fails on slice check when not wrapped
		_, err := s.Validate(nil)
		if err == nil || !strings.Contains(err.Error(), "expected slice") {
			t.Errorf("expected expected slice error, got %v", err)
		}

		// too short
		_, err = s.Validate([]any{1})
		if err == nil || !strings.Contains(err.Error(), "less than minimum") {
			t.Errorf("expected less than minimum error, got %v", err)
		}

		// exactly at min
		val, err := s.Validate([]any{1, 2})
		if err != nil {
			t.Fatalf("unexpected error at min boundary: %v", err)
		}
		if !reflect.DeepEqual(val, []any{1, 2}) {
			t.Errorf("expected [1, 2], got %v", val)
		}

		// above min
		val, err = s.Validate([]any{1, 2, 3})
		if err != nil {
			t.Fatalf("unexpected error above min: %v", err)
		}
		if !reflect.DeepEqual(val, []any{1, 2, 3}) {
			t.Errorf("expected [1, 2, 3], got %v", val)
		}
	})

	t.Run("MaxLen", func(t *testing.T) {
		s := Array(Int()).MaxLen(2)

		// nil should succeed and return empty array since minLen is nil (0 is allowed)
		val, err := s.Validate(nil)
		if err != nil {
			t.Fatalf("unexpected error for nil: %v", err)
		}
		if !reflect.DeepEqual(val, []any{}) {
			t.Errorf("expected [], got %v", val)
		}

		// exactly at max
		val, err = s.Validate([]any{1, 2})
		if err != nil {
			t.Fatalf("unexpected error at max boundary: %v", err)
		}
		if !reflect.DeepEqual(val, []any{1, 2}) {
			t.Errorf("expected [1, 2], got %v", val)
		}

		// exceeds max
		_, err = s.Validate([]any{1, 2, 3})
		if err == nil || !strings.Contains(err.Error(), "greater than maximum") {
			t.Errorf("expected greater than maximum error, got %v", err)
		}
	})

	t.Run("Len", func(t *testing.T) {
		s := Array(Int()).Len(2)

		// too short
		_, err := s.Validate([]any{1})
		if err == nil || !strings.Contains(err.Error(), "less than minimum") {
			t.Errorf("expected less than minimum error, got %v", err)
		}

		// exact length
		val, err := s.Validate([]any{1, 2})
		if err != nil {
			t.Fatalf("unexpected error at exact length: %v", err)
		}
		if !reflect.DeepEqual(val, []any{1, 2}) {
			t.Errorf("expected [1, 2], got %v", val)
		}

		// too long
		_, err = s.Validate([]any{1, 2, 3})
		if err == nil || !strings.Contains(err.Error(), "greater than maximum") {
			t.Errorf("expected greater than maximum error, got %v", err)
		}
	})

	t.Run("Len(0)", func(t *testing.T) {
		s := Array(Int()).Len(0)

		// nil should succeed and return empty array since minLen is 0 (0 is allowed)
		val, err := s.Validate(nil)
		if err != nil {
			t.Fatalf("unexpected error for nil: %v", err)
		}
		if !reflect.DeepEqual(val, []any{}) {
			t.Errorf("expected [], got %v", val)
		}

		// empty slice — should pass
		val, err = s.Validate([]any{})
		if err != nil {
			t.Fatalf("unexpected error for empty slice with Len(0): %v", err)
		}
		if !reflect.DeepEqual(val, []any{}) {
			t.Errorf("expected [], got %v", val)
		}

		// non-empty — should fail
		_, err = s.Validate([]any{1})
		if err == nil || !strings.Contains(err.Error(), "greater than maximum") {
			t.Errorf("expected greater than maximum error for Len(0), got %v", err)
		}
	})

	t.Run("TSType", func(t *testing.T) {
		s := Array(String())
		expected := "Array<string>"
		if s.TSType() != expected {
			t.Errorf("expected %q, got %q", expected, s.TSType())
		}
	})
}

func TestObjectSchema(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		s := Object(map[string]Schema{
			"foo": String().Default("bar"),
		})
		// Since all fields are optional, ObjectSchema.Validate(nil) should succeed!
		val, err := s.Validate(nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := map[string]any{"foo": "bar"}
		if !reflect.DeepEqual(val, expected) {
			t.Errorf("expected %v, got %v", expected, val)
		}

		// If a field is required, it should fail on nil
		sRequired := Object(map[string]Schema{
			"foo": String(),
		})
		_, err = sRequired.Validate(nil)
		if err == nil {
			t.Error("expected error for nil when a field is required")
		}
	})

	t.Run("Validate", func(t *testing.T) {
		s := Object(map[string]Schema{
			"foo": String(),
			"bar": Int().Default(42),
		})

		input := map[string]any{
			"foo": "hello",
			"baz": "ignored",
		}
		val, err := s.Validate(input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expected := map[string]any{
			"foo": "hello",
			"bar": 42,
		}
		if !reflect.DeepEqual(val, expected) {
			t.Errorf("expected %v, got %v", expected, val)
		}

		// typed map input
		typedInput := map[string]string{
			"foo": "world",
		}
		val, err = s.Validate(typedInput)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expectedTyped := map[string]any{
			"foo": "world",
			"bar": 42,
		}
		if !reflect.DeepEqual(val, expectedTyped) {
			t.Errorf("expected %v, got %v", expectedTyped, val)
		}

		// property validation error
		badInput := map[string]any{
			"foo": 123,
		}
		_, err = s.Validate(badInput)
		if err == nil || !strings.Contains(err.Error(), `at property "foo"`) {
			t.Errorf("expected property foo error, got %v", err)
		}

		// non-map input
		_, err = s.Validate("not-a-map")
		if err == nil || !strings.Contains(err.Error(), "expected map") {
			t.Errorf("expected map error, got %v", err)
		}
	})

	t.Run("TSType", func(t *testing.T) {
		s := Object(map[string]Schema{
			"foo": String(),
			"bar": Int(),
		})
		// should be sorted alphabetically, no question marks since not optional
		expected := "{ bar: number; foo: string }"
		if s.TSType() != expected {
			t.Errorf("expected %q, got %q", expected, s.TSType())
		}

		sMixed := Object(map[string]Schema{
			"foo": String(),
			"bar": Int().Default(42),
		})
		expectedMixed := "{ bar?: number; foo: string }"
		if sMixed.TSType() != expectedMixed {
			t.Errorf("expected %q, got %q", expectedMixed, sMixed.TSType())
		}

		empty := Object(nil)
		if empty.TSType() != "Record<string, any>" {
			t.Errorf("expected Record<string, any>, got %q", empty.TSType())
		}
	})
}

func TestTupleSchema(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		s := Tuple(String().Default("a"), Int().Default(2))
		// Since all elements are optional, TupleSchema.Validate(nil) should succeed!
		val, err := s.Validate(nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := []any{"a", 2}
		if !reflect.DeepEqual(val, expected) {
			t.Errorf("expected %v, got %v", expected, val)
		}

		// If an element is required, it should fail on nil
		sRequired := Tuple(String(), Int().Default(2))
		_, err = sRequired.Validate(nil)
		if err == nil {
			t.Error("expected error for nil when an element is required")
		}
	})

	t.Run("Validate", func(t *testing.T) {
		s := Tuple(String(), Int().Default(0))

		input := []any{"hello", 10.0}
		val, err := s.Validate(input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := []any{"hello", 10}
		if !reflect.DeepEqual(val, expected) {
			t.Errorf("expected %v, got %v", expected, val)
		}

		// single element input is not a slice, should fail
		_, err = s.Validate("world")
		if err == nil {
			t.Error("expected error for non-slice input")
		}

		// element error
		badInput := []any{"hello", "not-an-int"}
		_, err = s.Validate(badInput)
		if err == nil || !strings.Contains(err.Error(), "at element 1") {
			t.Errorf("expected at element 1 error, got %v", err)
		}
	})

	t.Run("TSType", func(t *testing.T) {
		s := Tuple(String(), Int(), Bool())
		expected := "[string, number, boolean]"
		if s.TSType() != expected {
			t.Errorf("expected %q, got %q", expected, s.TSType())
		}

		sMixed := Tuple(String(), Int().Default(42), Bool().Default(true))
		expectedMixed := "[string, number?, boolean?]"
		if sMixed.TSType() != expectedMixed {
			t.Errorf("expected %q, got %q", expectedMixed, sMixed.TSType())
		}
	})
}

func TestUnionSchema(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		s := Union(String(), Bool()).Default("hello")
		val, err := s.Validate(nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if val != "hello" {
			t.Errorf("expected 'hello', got %v", val)
		}
	})

	t.Run("Validate", func(t *testing.T) {
		s := Union(String(), Int())

		val, err := s.Validate("test")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if val != "test" {
			t.Errorf("expected 'test', got %v", val)
		}

		val, err = s.Validate(100.0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if val != 100 {
			t.Errorf("expected 100, got %v", val)
		}

		_, err = s.Validate(true)
		if err == nil || !strings.Contains(err.Error(), "union failed") {
			t.Errorf("expected union failed error, got %v", err)
		}
	})

	t.Run("TSType", func(t *testing.T) {
		s := Union(String(), Int())
		expected := "(string | number)"
		if s.TSType() != expected {
			t.Errorf("expected %q, got %q", expected, s.TSType())
		}
	})
}

func TestHasDefault(t *testing.T) {
	// 1. Basic schemas do not have defaults (except AnySchema)
	if Bool().HasDefault() {
		t.Error("BoolSchema should not have a default")
	}
	if Int().HasDefault() {
		t.Error("IntSchema should not have a default")
	}
	if String().HasDefault() {
		t.Error("StringSchema should not have a default")
	}
	if !Any().HasDefault() {
		t.Error("AnySchema should have a default")
	}

	// 2. DefaultSchema has a default
	if !Bool().Default(true).HasDefault() {
		t.Error("DefaultSchema should have a default")
	}

	// 3. UnionSchema never has a default directly
	if Union(Bool(), String()).HasDefault() {
		t.Error("UnionSchema should not have a default")
	}
	if Union(Bool().Default(true), String().Default("")).HasDefault() {
		t.Error("UnionSchema should not have a default even if members do")
	}

	// 4. TupleSchema has a default if all options have defaults
	if Tuple(Bool(), String().Default("")).HasDefault() {
		t.Error("TupleSchema with non-default members should not have a default")
	}
	if !Tuple(Bool().Default(true), String().Default("")).HasDefault() {
		t.Error("TupleSchema with all default members should have a default")
	}

	// 5. ArraySchema has a default if length of 0 is allowed
	if !Array(Int()).HasDefault() {
		t.Error("Array(Int()) should have a default since length 0 is allowed")
	}
	if !Array(Int()).MinLen(0).HasDefault() {
		t.Error("Array(Int()).MinLen(0) should have a default since length 0 is allowed")
	}
	if Array(Int()).MinLen(1).HasDefault() {
		t.Error("Array(Int()).MinLen(1) should not have a default since length 0 is not allowed")
	}
}

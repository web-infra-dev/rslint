# Schema Validation Reference Manual (`internal/rule/schema.go`)

`rslint` uses a declarative, type-safe options validation framework inspired by libraries like Zod. When migrating a rule to the new framework, you define a `Schema` matching the rule's expected option payload, and the engine automatically handles validation, default injection, and type mapping.

---

## 1. Schema Combinators

All combinators are prefixed with `rule.` and are defined in `internal/rule/schema.go`.

### Basic Value Schemas

- **`rule.Bool()`**: Validates booleans. Maps to Go type `bool` / TypeScript `boolean`.
- **`rule.Int()`**: Validates integer values. Maps to Go type `int` / TypeScript `number`.
  - Supporting fluent helpers: `.Min(int)` and `.Max(int)`.
- **`rule.String()`**: Validates strings. Maps to Go type `string` / TypeScript `string`.
- **`rule.Enum(allowed ...string)`**: Validates string values matching a fixed set of allowed options. Maps to Go type `string` / TypeScript string union (e.g., `"always" | "never"`).
- **`rule.Any()`**: Bypasses validation and accepts any raw input. Maps to Go type `any` / TypeScript `any`.

### Structural / Composite Schemas

- **`rule.Object(map[string]rule.Schema)`**: Validates key-value properties. Maps to Go type `map[string]any` / TypeScript `{ key: valueType }`.
- **`rule.Array(itemSchema)`**: Validates a slice of homogeneous values. Maps to Go type `[]any` / TypeScript `Array<itemType>`.
  - Supporting fluent helpers: `.MinLen(int)`, `.MaxLen(int)`, and `.Len(int)`.
- **`rule.Tuple(itemSchemas ...Schema)`**: Validates an ordered slice of heterogeneous values (positional parameters). Maps to Go type `[]any` / TypeScript tuple type.
- **`rule.Union(itemSchemas ...Schema)`**: Validates if the input matches _at least one_ of the specified schemas. Maps to Go type `any` / TypeScript union type.

### Empty Schema

- **`rule.EmptyArray()`**: Enforces that no configuration options are provided (empty array or nil). If a rule takes no options, **you MUST specify `schema: rule.EmptyArray()`**.

---

## 2. Default Value Injection

By default, any basic schema (like `Bool`, `Int`, `String`, `Enum`, `Array`, etc.) requires a value unless it is wrapped with a default value:

- **`.Default(defaultValue)`**: Specifies a fallback value to inject when the option is omitted (i.e. `nil`).
- Calling `.Default(...)` wraps the schema in a `DefaultSchema`, making it optional.

Example:

```go
rule.Bool().Default(false)
rule.Enum("always", "never").Default("always")
```

---

## 3. Top-Level Options vs. Go Type Mapping

The top-level option container configured in `RslintConfig` (the `rules` block) is always processed as a slice. Therefore, **the top-level `Schema` defined on a `rule.Rule` MUST be one of: `rule.Tuple(...)`, `rule.Array(...)`, or `rule.EmptyArray()`**.

The framework parses and passes the validated options as `[]any` to the rule's `RunWithOptions` callback. Here is the mapping of schemas to Go types in `options []any`:

| Schema Combinator                             | Validated Go Type                                            | Example Usage in Go                                                                           |
| --------------------------------------------- | ------------------------------------------------------------ | --------------------------------------------------------------------------------------------- |
| `rule.EmptyArray()`                           | `[]any` (empty)                                              | _None (no options expected)_                                                                  |
| `rule.Tuple(rule.Object(...))`                | `options[0]` is `map[string]any`                             | `optsMap := rule.Must[map[string]any](options[0])`                                            |
| `rule.Tuple(rule.String(), rule.Object(...))` | `options[0]` is `string`<br>`options[1]` is `map[string]any` | `mode := rule.Must[string](options[0])`<br>`optsMap := rule.Must[map[string]any](options[1])` |
| `rule.Array(rule.String())`                   | `[]any` (homogeneous items)                                  | `for _, item := range options { str := rule.Must[string](item) }`                             |
| `rule.Union(rule.String(), ...)`              | Go type matching the matched union variant                   | `switch val := item.(type) { case string: ... }`                                              |

---

## 4. Retrieving Typed Options (`rule.Must[T]`)

To retrieve values from schema-validated `any` fields, use **`rule.Must[T](val)`** instead of raw Go type assertions (e.g., `val.(T)`).

```go
func Must[T any](v any) T
```

### Why use `rule.Must[T]`?

1. **Defensive Programming**: Since options are guaranteed to match the schema by the validation engine at runtime, a mismatch indicates a schema definition bug. `rule.Must[T]` panics with a descriptive error detailing the mismatch.
2. **Linter Compliance**: Avoids triggering `forcetypeassert` errors in `golangci-lint` without requiring `//nolint:forcetypeassert` bypass comments at every call site.

### Example

```go
optsMap := rule.Must[map[string]any](options[0])
allowFoo := rule.Must[bool](optsMap["allowFoo"])
excludeArr := rule.Must[[]any](optsMap["exclude"])
```

---
name: port-rule-to-schema
description: Port a legacy rslint rule to the new schema-driven options validation framework. Use this skill when you need to refactor a rule implementation to define a declarative Schema (a Tuple or Array) and consume typed options in RunWithOptions.
---

# Porting Rules to the Schema-Driven Options Framework

This skill guides you through the process of migrating a legacy `rslint` rule (which parses options manually in its `Run` function) to the new declarative, Zod-like schema-driven option validation framework.

---

## Migration Steps

### 1. Locate the Rule Implementation

Find the rule definition, implementation, documentation, and tests:

- Core Go Implementation: `internal/rules/<rule_name>/<rule_name>.go` (Tests: `internal/rules/<rule_name>/<rule_name>_test.go`)
- Plugin Go Implementation: `internal/plugins/<plugin_name>/rules/<rule_name>/<rule_name>.go` (Tests: `internal/plugins/<plugin_name>/rules/<rule_name>/<rule_name>_test.go`)
- Rule Documentation: `<rule_name>.md` co-located in the rule folder.

### 2. Identify the Existing Option Parsing Logic and Documentation

- **Read the Documentation**: Open and read the `<rule_name>.md` file in the rule directory. It contains critical user-facing descriptions of the option fields, their names, enums, and defaults.
- **Inspect legacy Run function**: Locate the legacy `Run` function:

```go
Run: func(ctx rule.RuleContext) rule.RuleListeners { ... }
```

Inspect how `ctx.Rule().Options` is parsed, checked, and converted. Cross-reference this with the documentation to determine:

- The types of the options (e.g., boolean flags, string enums, object properties).
- The default values used when option fields are missing.
- Positional structures (e.g., option index `0` vs. index `1`).

### 3. Define the Declarative Schema

Add a `Schema` field to the `rule.Rule` declaration. **The top-level schema must be one of `rule.Tuple(...)`, `rule.Array(...)`, or `rule.EmptyArray()`**:

- If the rule does not accept any configuration options, **you MUST provide `schema: rule.EmptyArray()`** to enforce that no options are passed.
- Use **`rule.Tuple(...)`** when the rule has a fixed number of positional arguments with potentially different types (e.g., a severity string followed by an options object).
- Use **`rule.Array(...)`** when the rule takes a variable-length list of homogeneous values (e.g., a list of allowed patterns or keywords).

All schema combinators are defined in [internal/rule/schema.go](file:///home/swwind/rslint/internal/rule/schema.go).

_Always specify `.Default(value)` on your schemas to let the framework handle missing values._

### 4. Implement `RunWithOptions`

Replace or supplement `Run` with `RunWithOptions`. The `options` parameter is always `[]any`:

```go
RunWithOptions: func(ctx rule.RuleContext, options []any) rule.RuleListeners { ... }
```

**When the top-level schema is `Tuple`**, access elements by index:

- `options[0]` → first positional arg (guaranteed by `Tuple` element 0)
- `options[1]` → second positional arg (guaranteed by `Tuple` element 1)
- etc.

**When the top-level schema is `Array`**, iterate over all elements — the length is variable:

```go
for _, item := range options {
    s := rule.Must[string](item) // or whatever the element type is
}
```

Use `rule.Must[T]` to extract and assert elements to the Go type that corresponds to its schema (this avoids `forcetypeassert` linter warnings and panics descriptively on schema mismatch):

- `rule.Object` → `rule.Must[map[string]any](val)`
- `rule.Enum` / `rule.String` → `rule.Must[string](val)`
- `rule.Bool` → `rule.Must[bool](val)`
- `rule.Array(...)` / `rule.Tuple(...)` → `rule.Must[[]any](val)`

### 5. Update Tests and Custom Call Sites

- In `<rule_name>_test.go`, make sure the tests pass. The rule tester runner automatically leverages the schemas to validate test case options before running rules.
- **Wrap Options in Slices**: In all test cases (both valid and invalid) inside the rule's tests, if `Options` are defined, **you MUST wrap them in a slice `[]interface{}{...}`** (or `[]interface{}{}` if empty). Because the schema-driven framework expects a slice at the top-level (Tuple or Array), passing a raw map/string directly will fail options validation with `expected slice, got map[string]interface {}`.
- **Verify Execution API**: Double-check that the ported rule is running with the new `.RunWithOptions` API rather than the legacy `.Run` API.
- **Drop Legacy `.Run` Calls**: If there is rule-specific code (such as custom test cases/fixtures in `<rule_name>_test.go` or rule-specific helper files) that only runs for this specific rule, you should drop support for the old `.Run` API entirely there. Migrate those custom calls to use `.RunWithOptions` (validating and hydrating options via `rule.ValidateAndHydrateOptions` first).

---

## Code Examples

### A. Single-Object Rule (`accessor-pairs`)

For a rule taking a single configuration object, wrap it in `Tuple`:

```go
type Options struct {
	GetWithoutSet bool
	SetWithoutGet bool
}

var AccessorPairsRule = rule.Rule{
	Name: "accessor-pairs",
	Schema: rule.Tuple(rule.Object(map[string]rule.Schema{
		"getWithoutSet": rule.Bool().Default(false),
		"setWithoutGet": rule.Bool().Default(true),
	})),
	RunWithOptions: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		optsMap := rule.Must[map[string]any](options[0])
		opts := Options{
			GetWithoutSet: rule.Must[bool](optsMap["getWithoutSet"]),
			SetWithoutGet: rule.Must[bool](optsMap["setWithoutGet"]),
		}
		// ... rule logic using opts
	},
}
```

### B. Variable-Length Rule (Array top-level)

For a rule that accepts a variable list of homogeneous values (e.g., allowed identifiers):

```go
var NoRestrictedGlobalsRule = rule.Rule{
	Name:   "no-restricted-globals",
	Schema: rule.Array(rule.Union(
		rule.String(),
		rule.Object(map[string]rule.Schema{
			"name":    rule.String(),
			"message": rule.String().Default(""),
		}),
	)),
	RunWithOptions: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		// options is a []any — one entry per restricted global
		for _, item := range options {
			switch v := item.(type) {
			case string:
				// plain name
			case map[string]any:
				name := rule.Must[string](v["name"])
				msg := rule.Must[string](v["message"])
				_ = name
				_ = msg
			}
		}
		// ... rule logic
	},
}
```

### C. Multi-Positional Rule (`eqeqeq`)

For a rule taking multiple positional arguments, list them as Tuple elements:

```go
var EqeqeqRule = rule.Rule{
	Name: "eqeqeq",
	Schema: rule.Tuple(
		rule.Enum("always", "smart").Default("always"),
		rule.Object(map[string]rule.Schema{
			"null": rule.Enum("always", "never", "ignore").Default("always"),
		}),
	),
	RunWithOptions: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		mode := rule.Must[string](options[0])
		optsMap := rule.Must[map[string]any](options[1])
		nullOption := rule.Must[string](optsMap["null"])

		// ... rule logic using mode & nullOption
	},
}
```

---

## References

- [SCHEMA_MANUAL.md](references/SCHEMA_MANUAL.md) - Manual for declarative schema combinators (types, defaults, and Go mapping)

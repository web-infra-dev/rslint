---
name: port-rule-to-schema
description: Port a legacy rslint rule to the new schema-driven options validation framework. Use this skill when you need to refactor a rule implementation to define a declarative Schema (always a Tuple) and consume typed options in RunWithOptions.
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

Add a `Schema` field to the `rule.Rule` declaration. **The top-level schema must always be `rule.Tuple(...)`**, even for rules with a single option — this ensures `RunWithOptions` always receives a `[]any` slice. Use the schema combinators defined in [internal/rule/schema.go](file:///home/swwind/rslint/internal/rule/schema.go) as the Tuple elements:

- `rule.Bool()`
- `rule.Int()`
- `rule.String()`
- `rule.Enum("option1", "option2")`
- `rule.Object(map[string]rule.Schema{ ... })`
- `rule.Array(itemSchema)`
- `rule.Union(schemas...)`

_Always specify `.Default(value)` on your schemas to let the framework handle missing values._

### 4. Implement `RunWithOptions`

Replace or supplement `Run` with `RunWithOptions`. The `options` parameter is always `[]any` — one element per positional schema in the `Tuple`:

```go
RunWithOptions: func(ctx rule.RuleContext, options []any) rule.RuleListeners { ... }
```

Access elements by index:

- `options[0]` → first positional arg (guaranteed by `Tuple` element 0)
- `options[1]` → second positional arg (guaranteed by `Tuple` element 1)
- etc.

Type-assert each element to the Go type that corresponds to its schema:

- `rule.Object` → `map[string]any`
- `rule.Enum` / `rule.String` → `string`
- `rule.Bool` → `bool`
- `rule.Array(rule.String())` → `[]any` (elements are `string`)

### 5. Update Tests and Custom Call Sites

- In `<rule_name>_test.go`, make sure the tests pass. The rule tester runner automatically leverages the schemas to validate test case options before running rules.
- **Verify Execution API**: Double-check that the ported rule is running with the new `.RunWithOptions` API rather than the legacy `.Run` API.
- **Drop Legacy `.Run` Calls**: If there is rule-specific code (such as custom test cases/fixtures in `<rule_name>_test.go` or rule-specific helper files) that only runs for this specific rule, you should drop support for the old `.Run` API entirely there. Migrate those custom calls to use `.RunWithOptions` (validating and hydrating options via `rule.ValidateAndHydrateOptions` first).

---

## Code Examples

### A. Single Option Rule (`accessor-pairs`)

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
		optsMap, _ := options[0].(map[string]any)
		opts := Options{
			GetWithoutSet: optsMap["getWithoutSet"].(bool),
			SetWithoutGet: optsMap["setWithoutGet"].(bool),
		}
		// ... rule logic using opts
	},
}
```

### B. Multi-Positional Rule (`eqeqeq`)

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
		mode, _ := options[0].(string)
		optsMap, _ := options[1].(map[string]any)
		nullOption, _ := optsMap["null"].(string)

		// ... rule logic using mode & nullOption
	},
}
```

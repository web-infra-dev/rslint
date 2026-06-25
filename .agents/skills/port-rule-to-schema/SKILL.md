---
name: port-rule-to-schema
description: Port a legacy rslint rule to the new schema-driven options validation framework. Use this skill when you need to refactor a rule implementation to define declarative options schemas (Schema0, Schema1) and consume typed options in RunWithOptions.
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

### 3. Define the Declarative Schemas

Add `Schema0` (and optionally `Schema1` if the rule takes a second positional option) to the `rule.Rule` declaration using the schema combinators defined in [internal/rule/schema.go](file:///home/swwind/rslint/internal/rule/schema.go):

- `rule.Bool()`
- `rule.Int()`
- `rule.String()`
- `rule.Enum("option1", "option2")`
- `rule.Object(map[string]rule.Schema{ ... })`
- `rule.Array(itemSchema)`
- `rule.Tuple(schemas...)`
- `rule.Union(schemas...)`

_Always specify `.Default(value)` on your schemas to let the framework handle missing values._

### 4. Implement `RunWithOptions`

Replace or supplement `Run` with `RunWithOptions`:

```go
RunWithOptions: func(ctx rule.RuleContext, options any) rule.RuleListeners { ... }
```

Within `RunWithOptions`, cast the `options` parameter to the expected Go types guaranteed by your schema:

- **Single Option Rules** (`Schema1 == nil`): `options` corresponds directly to `Schema0`.
  - If `Schema0` is a `rule.Object`, cast `options` to `map[string]any`.
  - If `Schema0` is a `rule.Enum`, cast `options` to `string`.
- **Double Option Rules** (`Schema1 != nil`): `options` is guaranteed to be a flat slice of exactly 2 elements (`[]any`).
  - Cast `options` to `[]any`.
  - Access the elements via `options.([]any)[0]` (validated by `Schema0`) and `options.([]any)[1]` (validated by `Schema1`).

### 5. Update Tests and Custom Call Sites

- In `<rule_name>_test.go`, make sure the tests pass. The rule tester runner automatically leverages the schemas to validate test case options before running rules.
- **Verify Execution API**: Double-check that the ported rule is running with the new `.RunWithOptions` API rather than the legacy `.Run` API.
- **Drop Legacy `.Run` Calls**: If there is rule-specific code (such as custom test cases/fixtures in `<rule_name>_test.go` or rule-specific helper files) that only runs for this specific rule, you should drop support for the old `.Run` API entirely there. Migrate those custom calls to use `.RunWithOptions` (validating and hydrating options via `rule.ValidateAndHydrateOptions` first).

---

## Code Examples

### A. Single Option Rule (`accessor-pairs`)

For a rule taking a single configuration object:

```go
type Options struct {
	GetWithoutSet bool
	SetWithoutGet bool
}

var AccessorPairsRule = rule.Rule{
	Name: "accessor-pairs",
	Schema0: rule.Object(map[string]rule.Schema{
		"getWithoutSet": rule.Bool().Default(false),
		"setWithoutGet": rule.Bool().Default(true),
	}),
	RunWithOptions: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		optsMap := options.(map[string]any)
		opts := Options{
			GetWithoutSet: optsMap["getWithoutSet"].(bool),
			SetWithoutGet: optsMap["setWithoutGet"].(bool),
		}
		// ... rule logic using opts
	},
}
```

### B. Double Option Rule (`eqeqeq`)

For a rule taking a string enum as the first option and a configuration object as the second option:

```go
var EqeqeqRule = rule.Rule{
	Name: "eqeqeq",
	Schema0: rule.Enum("always", "smart").Default("always"),
	Schema1: rule.Object(map[string]rule.Schema{
		"null": rule.Enum("always", "never", "ignore").Default("always"),
	}),
	RunWithOptions: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := options.([]any)
		mode := opts[0].(string)
		optsMap := opts[1].(map[string]any)
		nullOption := optsMap["null"].(string)

		// ... rule logic using mode & nullOption
	},
}
```

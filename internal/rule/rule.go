package rule

import "github.com/microsoft/typescript-go/shim/ast"

// DiagnosticSeverity represents the severity level of a diagnostic
type DiagnosticSeverity int

const (
	SeverityError DiagnosticSeverity = iota
	SeverityWarning
	SeverityOff
)

// String returns the string representation of the severity
func (s DiagnosticSeverity) String() string {
	switch s {
	case SeverityError:
		return "error"
	case SeverityWarning:
		return "warn"
	case SeverityOff:
		return "off"
	default:
		return "error"
	}
}

// Int returns the numeric representation of the severity.
func (s DiagnosticSeverity) Int() int {
	switch s {
	case SeverityError:
		return 1
	case SeverityWarning:
		return 2
	case SeverityOff:
		return 0
	default:
		return 0
	}
}

// ParseSeverity converts a string to DiagnosticSeverity
func ParseSeverity(level string) DiagnosticSeverity {
	switch level {
	case "error":
		return SeverityError
	case "warn", "warning":
		return SeverityWarning
	case "off":
		return SeverityOff
	default:
		return SeverityError // Default to error for unknown values
	}
}

// NormalizeOptions returns a rule's options in ESLint context.options form
// ([]any). config.parseArrayRuleConfig never collapses RuleConfig.Options, but
// LegacyUnwrapOptions (the compatibility shim for pre-migration `parseOptions
// any` bodies) does — a rule that round-trips through it and then wants the
// eslint-format array back (e.g. reads optArray[0]) would silently miss a
// single-option config otherwise. Re-wrapping a bare value lets every caller
// read options[0] uniformly, whether the option arrived wrapped (multi-option
// or an explicit array) or unwrapped (a single option).
//
// It returns an empty (non-nil) slice when no options were configured, so both
// native rules (which key on `len == 0 → defaults`) and the eslint-plugin host
// (which serializes context.options to JSON and needs `[]`, not `null`) share a
// single normalization path.
func NormalizeOptions(raw any) []any {
	if raw == nil {
		return []any{}
	}
	if arr, ok := raw.([]interface{}); ok {
		return arr
	}
	return []any{raw}
}

// LegacyUnwrapOptions is NormalizeOptions' inverse: it collapses a rule's options
// array (Run's []any parameter — ESLint's context.options, the config array
// after the severity level) back to the single bare value most existing rule
// implementations parse. Empty → nil; a single element → that element;
// otherwise the slice itself. This is the compatibility shim old
// `parseOptions(options any)` bodies call so they don't need to change beyond
// their Run signature.
func LegacyUnwrapOptions(options []any) any {
	switch len(options) {
	case 0:
		return nil
	case 1:
		return options[0]
	default:
		return options
	}
}

const (
	lastTokenKind                        ast.Kind = 1000
	lastOnExitTokenKind                  ast.Kind = 2000
	lastOnAllowPatternTokenKind          ast.Kind = 3000
	lastOnAllowPatternOnExitTokenKind    ast.Kind = 4000
	lastOnNotAllowPatternTokenKind       ast.Kind = 5000
	lastOnNotAllowPatternOnExitTokenKind ast.Kind = 6000
)

func ListenerOnExit(kind ast.Kind) ast.Kind {
	return kind + 1000
}

// TODO(port): better name
func ListenerOnAllowPattern(kind ast.Kind) ast.Kind {
	return kind + lastOnExitTokenKind
}
func ListenerOnNotAllowPattern(kind ast.Kind) ast.Kind {
	return kind + lastOnAllowPatternOnExitTokenKind
}

type RuleListeners map[ast.Kind](func(node *ast.Node))

type Rule struct {
	Name             string
	RequiresTypeInfo bool
	// IsEslintPluginRule marks a placeholder rule whose actual execution
	// happens in a Node worker — an ESLint-plugin rule mounted via the
	// config's object-form `plugins`. Its Run is a no-op in Go; the linter
	// splits these out and dispatches them to the plugin-lint host.
	IsEslintPluginRule bool
	// Schema validates the rule's resolved options array (ESLint's
	// context.options — the config array after the severity level) before
	// linting starts, filling schema-declared `default` values into the
	// options in place the way ajv's `useDefaults` does for ESLint (see
	// [Schema.Validate]). Rules that take no options should set it to the
	// shared [EmptyArraySchema]. nil means the rule has not declared a schema
	// yet (most rules, until migrated one-by-one): its options pass through
	// unvalidated, exactly as before the schema framework existed.
	Schema *Schema
	Run    func(ctx RuleContext, options []any) RuleListeners
}

func CreateRule(r Rule) Rule {
	return Rule{
		Name:             "@typescript-eslint/" + r.Name,
		RequiresTypeInfo: r.RequiresTypeInfo,
		Schema:           r.Schema,
		Run:              r.Run,
	}
}

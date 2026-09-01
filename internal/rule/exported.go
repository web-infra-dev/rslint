package rule

// Exported is the file's complete `/* exported */` view: the names an inline
// directive marks as intentionally global, plus the declaration metadata behind
// them. Its zero value declares no names.
//
// The directive is ESLint's script-only escape hatch for code that shares
// globals across separately loaded files: a name listed here is declared in one
// file and consumed by another, so the rules that would otherwise call it
// unused, or call declaring it a global-scope pollution, leave it alone.
//
// ESLint applies it by looking each name up in the global scope and setting
// `eslintUsed` and `eslintExported` on the variable it finds, which silently
// drops names bound anywhere else — a module's top-level bindings live in the
// module scope, so the directive does nothing there. rslint has no such
// per-variable flag, so this view answers for the raw name and each rule
// applies its own scope gating.
type Exported struct {
	names        map[string]bool
	declarations []InlineExported
}

// NewExported constructs the per-file exported view. The map and slice are
// treated as read-only and may be shared with the comment store that produced
// them.
func NewExported(names map[string]bool, declarations []InlineExported) Exported {
	return Exported{names: names, declarations: declarations}
}

// Has reports whether an `/* exported */` comment in this file lists name.
func (e Exported) Has(name string) bool {
	return e.names[name]
}

// Declarations returns every parsed inline exported declaration in source
// order, including repeated names and their source ranges. Treat it as
// read-only.
func (e Exported) Declarations() []InlineExported {
	return e.declarations
}

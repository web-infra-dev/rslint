package test_framework

import "github.com/microsoft/typescript-go/shim/ast"

// FnKind classifies a parsed test-framework call without encoding one
// framework's root names or aliases. Framework parsers may define additional
// values for assertion and namespace APIs.
type FnKind string

const (
	FnKindDescribe FnKind = "describe"
	FnKindHook     FnKind = "hook"
	FnKindTest     FnKind = "test"
	FnKindUnknown  FnKind = "unknown"
)

// ReferenceMode records whether a test API was resolved as a global or as an
// import/require binding from a framework module.
type ReferenceMode string

const (
	ReferenceModeGlobal ReferenceMode = "global"
	ReferenceModeImport ReferenceMode = "import"
)

// MemberEntry describes one segment in a test API member chain.
type MemberEntry struct {
	Name string
	Node *ast.Node
	Call *ast.Node
}

// ParsedCall is the framework-neutral portion of a parsed test API call.
// Framework parsers own root/member validation and may embed this type to add
// matcher or framework-specific metadata.
type ParsedCall struct {
	Name          string
	LocalName     string
	Kind          FnKind
	Members       []string
	MemberEntries []MemberEntry
	Head          ParsedCallHead
}

type ParsedCallHead struct {
	Type     ReferenceMode
	Local    ParsedCallHeadEntry
	Original ParsedCallHeadEntry
}

type ParsedCallHeadEntry struct {
	Value string
	Node  *ast.Node
}

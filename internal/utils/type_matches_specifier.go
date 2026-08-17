package utils

import (
	"fmt"
	"slices"
	"strings"
	"sync"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/module"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/program"
	esregexp "github.com/web-infra-dev/rslint/internal/utils/ecmascript/regexp"
)

type TypeOrValueSpecifierFrom uint8

const (
	TypeOrValueSpecifierFromFile TypeOrValueSpecifierFrom = iota
	TypeOrValueSpecifierFromLib
	TypeOrValueSpecifierFromPackage
	// TypeOrValueSpecifierFromString represents the string shorthand form.
	// It matches any type with the given name regardless of origin.
	TypeOrValueSpecifierFromString
)

type NameList []string

type TypeOrValueSpecifier struct {
	From TypeOrValueSpecifierFrom
	Name NameList
	// Can be used when From == TypeOrValueSpecifierFromFile
	Path string
	// Can be used when From == TypeOrValueSpecifierFromPackage
	Package string
	// pathProvided distinguishes an omitted path from an explicitly empty one.
	// Both decode to Path == "", but upstream treats only the omitted form as
	// matching declarations anywhere in the current project.
	pathProvided bool
}

// ParseTypeOrValueSpecifier decodes one entry of a type specifier option from
// its configured value, handling both the string shorthand ("Promise") and the
// object form ({"from": "lib", "name": "Promise"}), matching the original
// TypeScript-ESLint TypeOrValueSpecifier union type. It reports false for a
// value that matches neither shape.
func ParseTypeOrValueSpecifier(raw any) (TypeOrValueSpecifier, bool) {
	if str, ok := raw.(string); ok {
		return TypeOrValueSpecifier{
			From: TypeOrValueSpecifierFromString,
			Name: NameList{str},
		}, true
	}

	fields, ok := raw.(map[string]any)
	if !ok {
		return TypeOrValueSpecifier{}, false
	}

	var specifier TypeOrValueSpecifier
	if raw, ok := fields["from"]; ok {
		str, ok := raw.(string)
		if !ok {
			return TypeOrValueSpecifier{}, false
		}
		switch str {
		case "file":
			specifier.From = TypeOrValueSpecifierFromFile
		case "lib":
			specifier.From = TypeOrValueSpecifierFromLib
		case "package":
			specifier.From = TypeOrValueSpecifierFromPackage
		default:
			return TypeOrValueSpecifier{}, false
		}
	}
	if raw, ok := fields["name"]; ok {
		names, ok := parseNameList(raw)
		if !ok {
			return TypeOrValueSpecifier{}, false
		}
		specifier.Name = names
	}
	if raw, ok := fields["path"]; ok {
		str, ok := raw.(string)
		if !ok {
			return TypeOrValueSpecifier{}, false
		}
		specifier.Path = str
		specifier.pathProvided = true
	}
	if raw, ok := fields["package"]; ok {
		str, ok := raw.(string)
		if !ok {
			return TypeOrValueSpecifier{}, false
		}
		specifier.Package = str
	}
	return specifier, true
}

// ParseTypeOrValueSpecifiers decodes a list of type specifiers. It returns nil
// when the value is not a list or when any entry matches neither specifier
// shape, so a malformed option falls back to the rule's default.
func ParseTypeOrValueSpecifiers(raw any) []TypeOrValueSpecifier {
	items, ok := raw.([]any)
	if !ok {
		return nil
	}
	specifiers := make([]TypeOrValueSpecifier, 0, len(items))
	for _, item := range items {
		specifier, ok := ParseTypeOrValueSpecifier(item)
		if !ok {
			return nil
		}
		specifiers = append(specifiers, specifier)
	}
	return specifiers
}

// parseNameList decodes a specifier's `name`, written either as a single name
// or as a list of them.
func parseNameList(raw any) (NameList, bool) {
	if str, ok := raw.(string); ok {
		return NameList{str}, true
	}
	items, ok := raw.([]any)
	if !ok {
		return nil, false
	}
	names := make(NameList, 0, len(items))
	for _, item := range items {
		str, ok := item.(string)
		if !ok {
			return nil, false
		}
		names = append(names, str)
	}
	return names, true
}

func specifierNameMatchesWithCalleeNames(
	t *checker.Type,
	names []string,
	calleeNames []string,
) bool {
	alias := checker.Type_alias(t)
	var symbol *ast.Symbol
	if alias == nil {
		symbol = checker.Type_symbol(t)
	} else {
		symbol = alias.Symbol()
	}

	if symbol != nil && slices.Contains(names, symbol.Name) {
		return true
	}

	// Also check against callee names (handles export aliases like `export { test as it }`)
	// where the type's symbol name is "test" but the callee identifier is "it"
	for _, calleeName := range calleeNames {
		if slices.Contains(names, calleeName) {
			return true
		}
	}

	if IsIntrinsicType(t) && slices.Contains(names, t.AsIntrinsicType().IntrinsicName()) {
		return true
	}

	return false
}

func typeMatchesInlineSpecifierWithCalleeNames(
	t *checker.Type,
	names []string,
	calleeNames []string,
) bool {
	if parts := UnionTypeParts(t); len(parts) > 1 {
		return Every(parts, func(part *checker.Type) bool {
			return typeMatchesInlineSpecifierWithCalleeNames(part, names, calleeNames)
		})
	}
	if IsIntrinsicErrorType(t) {
		return false
	}
	if specifierNameMatchesWithCalleeNames(t, names, calleeNames) {
		return true
	}
	if parts := IntersectionTypeParts(t); len(parts) > 1 {
		return Some(parts, func(part *checker.Type) bool {
			return typeMatchesInlineSpecifierWithCalleeNames(part, names, calleeNames)
		})
	}
	return false
}

func typeDeclaredInFile(
	relativePath string,
	pathProvided bool,
	declarationFiles []*ast.SourceFile,
	program *program.Program,
) bool {
	cwd := program.CurrentDirectory()
	useCaseSensitiveFileNames := program.FS().UseCaseSensitiveFileNames()
	canonical := func(fileName string) string {
		return tspath.GetCanonicalFileName(fileName, useCaseSensitiveFileNames)
	}
	if !pathProvided {
		return Some(declarationFiles, func(f *ast.SourceFile) bool {
			return strings.HasPrefix(canonical(f.FileName()), canonical(cwd))
		})
	}
	absPath := canonical(tspath.GetNormalizedAbsolutePath(relativePath, cwd))
	return Some(declarationFiles, func(f *ast.SourceFile) bool {
		return canonical(f.FileName()) == absPath
	})
}

func typeDeclaredInLib(
	declarationFiles []*ast.SourceFile,
	program *program.Program,
) bool {
	// Assertion: The type is not an error type.

	// Intrinsic type (i.e. string, number, boolean, etc) - Treat it as if it's from lib.
	if len(declarationFiles) == 0 {
		return true
	}
	return Some(declarationFiles, func(d *ast.SourceFile) bool {
		return program.IsSourceFileDefaultLibrary(d)
	})
}

func findParentModuleDeclaration(
	node *ast.Node,
) *ast.ModuleDeclaration {
	switch node.Kind {
	case ast.KindModuleDeclaration:
		decl := node.AsModuleDeclaration()
		// A namespace is transparent here: what it wraps still belongs to
		// whichever `declare module "pkg"` encloses the namespace itself.
		if decl.Keyword != ast.KindNamespaceKeyword {
			if ast.IsStringLiteral(decl.Name()) {
				return decl
			}
			return nil
		}
	case ast.KindSourceFile:
		return nil
	}
	return findParentModuleDeclaration(node.Parent)
}

func typeDeclaredInDeclareModule(
	packageName string,
	declarations []*ast.Node,
) bool {
	return Some(declarations, func(d *ast.Node) bool {
		parentModule := findParentModuleDeclaration(d)
		return parentModule != nil && parentModule.Name().Text() == packageName
	})
}

// packageMatchers caches the compiled matcher per `package` specifier, which
// comes from the rule options and so takes only a handful of distinct values.
var packageMatchers sync.Map // package name -> *esregexp.RegExp, nil when the pattern is invalid

// packageMatcher builds `new RegExp(`${packageName}|${typesPackageName}`)`. The
// pattern carries no anchors, so "demo" matches "demo-pkg" as well.
func packageMatcher(packageName string) *esregexp.RegExp {
	if cached, ok := packageMatchers.Load(packageName); ok {
		matcher, _ := cached.(*esregexp.RegExp)
		return matcher
	}
	matcher, err := esregexp.Compile(packageName+"|"+module.MangleScopedPackageName(packageName), "")
	if err != nil {
		matcher = nil
	}
	packageMatchers.Store(packageName, matcher)
	return matcher
}

func typeDeclaredInDeclarationFile(
	packageName string,
	declarationFiles []*ast.SourceFile,
	program *program.Program,
) bool {
	matcher := packageMatcher(packageName)
	if matcher == nil {
		return false
	}
	for _, file := range declarationFiles {
		if file == nil || !program.IsSourceFileFromExternalLibrary(file) {
			continue
		}
		for _, name := range program.PackageNamesForSourceFile(file) {
			if matcher.Test(name) {
				return true
			}
		}
	}
	return false
}

func typeDeclaredInPackageDeclarationFile(
	packageName string,
	declarations []*ast.Node,
	declarationFiles []*ast.SourceFile,
	program *program.Program,
) bool {
	return typeDeclaredInDeclareModule(packageName, declarations) ||
		typeDeclaredInDeclarationFile(packageName, declarationFiles, program)
}

func typeMatchesSpecifier(
	t *checker.Type,
	specifier TypeOrValueSpecifier,
	program *program.Program,
	calleeNames []string,
) bool {
	// Handle union types: all constituents must match the specifier
	if parts := UnionTypeParts(t); len(parts) > 1 {
		return Every(parts, func(part *checker.Type) bool {
			return typeMatchesSpecifier(part, specifier, program, calleeNames)
		})
	}
	if IsIntrinsicErrorType(t) {
		return false
	}

	wholeTypeMatches := false
	if specifierNameMatchesWithCalleeNames(t, specifier.Name, calleeNames) {
		symbol := checker.Type_symbol(t)
		if symbol == nil {
			alias := checker.Type_alias(t)
			if alias != nil {
				symbol = alias.Symbol()
			}
		}
		var declarations []*ast.Node
		if symbol != nil {
			declarations = symbol.Declarations
		}
		declarationFiles := Map(declarations, func(d *ast.Node) *ast.SourceFile {
			return ast.GetSourceFileOfNode(d)
		})

		switch specifier.From {
		case TypeOrValueSpecifierFromString:
			wholeTypeMatches = true
		case TypeOrValueSpecifierFromFile:
			wholeTypeMatches = typeDeclaredInFile(
				specifier.Path,
				specifier.pathProvided || specifier.Path != "",
				declarationFiles,
				program,
			)
		case TypeOrValueSpecifierFromLib:
			wholeTypeMatches = typeDeclaredInLib(declarationFiles, program)
		case TypeOrValueSpecifierFromPackage:
			wholeTypeMatches = typeDeclaredInPackageDeclarationFile(specifier.Package, declarations, declarationFiles, program)
		default:
			panic(fmt.Sprintf("unknown type specifier from: %v", specifier.From))
		}
	}
	if wholeTypeMatches {
		return true
	}
	if parts := IntersectionTypeParts(t); len(parts) > 1 {
		return Some(parts, func(part *checker.Type) bool {
			return typeMatchesSpecifier(part, specifier, program, calleeNames)
		})
	}
	return false
}

func TypeMatchesSomeSpecifier(
	t *checker.Type,
	specifiers []TypeOrValueSpecifier,
	inlineSpecifiers []string,
	program *program.Program,
) bool {
	return TypeMatchesSomeSpecifierWithCalleeNames(t, specifiers, inlineSpecifiers, program, nil)
}

// TypeMatchesSomeSpecifierWithCalleeNames is like TypeMatchesSomeSpecifier but also accepts
// callee names for matching export aliases (e.g., `export { test as it }` where the type's
// symbol name is "test" but the callee identifier is "it")
func TypeMatchesSomeSpecifierWithCalleeNames(
	t *checker.Type,
	specifiers []TypeOrValueSpecifier,
	inlineSpecifiers []string,
	program *program.Program,
	calleeNames []string,
) bool {
	return Some(specifiers, func(s TypeOrValueSpecifier) bool {
		return typeMatchesSpecifier(t, s, program, calleeNames)
	}) || typeMatchesInlineSpecifierWithCalleeNames(t, inlineSpecifiers, calleeNames)
}

// specifierStaticName is the name a specifier compares against when it matches
// the referenced value rather than its type.
func specifierStaticName(node *ast.Node) string {
	if node == nil {
		return ""
	}
	switch node.Kind {
	case ast.KindIdentifier, ast.KindStringLiteral:
		return node.Text()
	case ast.KindPrivateIdentifier:
		// A specifier names the member, so `#value` is spelled `value`.
		return strings.TrimPrefix(node.Text(), "#")
	}
	return ""
}

func valueMatchesSpecifier(
	node *ast.Node,
	specifier TypeOrValueSpecifier,
	program *program.Program,
	t *checker.Type,
) bool {
	staticName := specifierStaticName(node)
	if staticName == "" || !slices.Contains(specifier.Name, staticName) {
		return false
	}
	if specifier.From != TypeOrValueSpecifierFromPackage {
		return true
	}

	symbol := checker.Type_symbol(t)
	if symbol == nil {
		if alias := checker.Type_alias(t); alias != nil {
			symbol = alias.Symbol()
		}
	}
	var declarations []*ast.Node
	if symbol != nil {
		declarations = symbol.Declarations
	}
	declarationFiles := Map(declarations, func(d *ast.Node) *ast.SourceFile {
		return ast.GetSourceFileOfNode(d)
	})
	return typeDeclaredInPackageDeclarationFile(specifier.Package, declarations, declarationFiles, program)
}

// ValueMatchesSomeSpecifier matches the value a node references instead of its
// type, so a specifier can name an export whose type carries a different name.
// Only `from: "package"` narrows further than the name; `file` and `lib` value
// specifiers match on the name alone.
func ValueMatchesSomeSpecifier(
	node *ast.Node,
	specifiers []TypeOrValueSpecifier,
	program *program.Program,
	t *checker.Type,
) bool {
	return Some(specifiers, func(s TypeOrValueSpecifier) bool {
		return valueMatchesSpecifier(node, s, program, t)
	})
}

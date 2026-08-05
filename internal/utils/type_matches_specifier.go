package utils

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"sync"

	"github.com/dlclark/regexp2"
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/module"
	"github.com/microsoft/typescript-go/shim/tspath"
)

type TypeOrValueSpecifierFrom uint8

// unmarshal TypeOrValueSpecifierFrom from JSON string
func (s *TypeOrValueSpecifierFrom) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return fmt.Errorf("failed to unmarshal TypeOrValueSpecifierFrom: %w", err)
	}
	switch str {
	case "file":
		*s = TypeOrValueSpecifierFromFile
	case "lib":
		*s = TypeOrValueSpecifierFromLib
	case "package":
		*s = TypeOrValueSpecifierFromPackage
	default:
		return fmt.Errorf("unknown TypeOrValueSpecifierFrom value: %s", str)
	}
	return nil
}

const (
	TypeOrValueSpecifierFromFile TypeOrValueSpecifierFrom = iota
	TypeOrValueSpecifierFromLib
	TypeOrValueSpecifierFromPackage
	// TypeOrValueSpecifierFromString represents the string shorthand form.
	// It matches any type with the given name regardless of origin.
	TypeOrValueSpecifierFromString
)

type NameList []string

// unmarshal a string or a list of strings to NameList
func (s *NameList) UnmarshalJSON(data []byte) error {
	var singleName string
	if err := json.Unmarshal(data, &singleName); err == nil {
		*s = NameList{singleName}
		return nil
	}

	var names []string
	if err := json.Unmarshal(data, &names); err != nil {
		return fmt.Errorf("failed to unmarshal NameList: %w", err)
	}
	*s = names
	return nil
}

type TypeOrValueSpecifier struct {
	From TypeOrValueSpecifierFrom `json:"from"`
	Name NameList                 `json:"name"`
	// Can be used when From == TypeOrValueSpecifierFromFile
	Path string `json:"path"`
	// Can be used when From == TypeOrValueSpecifierFromPackage
	Package string `json:"package"`
}

// UnmarshalJSON handles both string shorthand ("Promise") and object form
// ({"from": "lib", "name": "Promise"}) for TypeOrValueSpecifier, matching
// the original TypeScript-ESLint TypeOrValueSpecifier union type.
func (s *TypeOrValueSpecifier) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		s.From = TypeOrValueSpecifierFromString
		s.Name = NameList{str}
		return nil
	}
	type Alias TypeOrValueSpecifier
	var alias Alias
	if err := json.Unmarshal(data, &alias); err != nil {
		return err
	}
	*s = TypeOrValueSpecifier(alias)
	return nil
}

func typeMatchesStringSpecifierWithCalleeNames(
	t *checker.Type,
	names []string,
	calleeNames []string,
) bool {
	// Handle union types: all constituents must match the specifier
	if parts := UnionTypeParts(t); len(parts) > 1 {
		return Every(parts, func(part *checker.Type) bool {
			return typeMatchesStringSpecifierWithCalleeNames(part, names, calleeNames)
		})
	}

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

func typeDeclaredInFile(
	relativePath string,
	declarationFiles []*ast.SourceFile,
	program *compiler.Program,
) bool {
	cwd := program.Host().GetCurrentDirectory()
	useCaseSensitiveFileNames := program.Host().FS().UseCaseSensitiveFileNames()
	canonical := func(fileName string) string {
		return tspath.GetCanonicalFileName(fileName, useCaseSensitiveFileNames)
	}
	if relativePath == "" {
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
	program *compiler.Program,
) bool {
	// Assertion: The type is not an error type.

	// Intrinsic type (i.e. string, number, boolean, etc) - Treat it as if it's from lib.
	if len(declarationFiles) == 0 {
		return true
	}
	return Some(declarationFiles, func(d *ast.SourceFile) bool {
		return IsSourceFileDefaultLibrary(program, d)
	})
}

func findParentModuleDeclaration(
	node *ast.Node,
) *ast.ModuleDeclaration {
	switch node.Kind {
	case ast.KindModuleDeclaration:
		decl := node.AsModuleDeclaration()
		if ast.IsStringLiteral(decl.Name()) {
			return decl
		}
		return nil
	case ast.KindSourceFile:
		return nil
	default:
		return findParentModuleDeclaration(node.Parent)
	}
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

// resolvedPackageName returns the package a declaration file belongs to, such as
// "demo-pkg" or "@scope/pkg" for a file under node_modules.
func resolvedPackageName(fileName string) string {
	packageDirectory := module.ParseNodeModuleFromPath(fileName, false)
	idx := strings.LastIndex(packageDirectory, "/node_modules/")
	if idx == -1 {
		return ""
	}
	return packageDirectory[idx+len("/node_modules/"):]
}

// packageMatchers caches the compiled matcher per `package` specifier, which
// comes from the rule options and so takes only a handful of distinct values.
var packageMatchers sync.Map // package name -> *regexp2.Regexp, nil when the pattern is invalid

// packageMatcher builds `new RegExp(`${packageName}|${typesPackageName}`)`. The
// pattern carries no anchors, so "demo" matches "demo-pkg" as well.
func packageMatcher(packageName string) *regexp2.Regexp {
	if cached, ok := packageMatchers.Load(packageName); ok {
		matcher, _ := cached.(*regexp2.Regexp)
		return matcher
	}
	matcher, err := CompileRegexp2(packageName+"|"+module.MangleScopedPackageName(packageName), JSRegexOptions)
	if err != nil {
		matcher = nil
	}
	packageMatchers.Store(packageName, matcher)
	return matcher
}

func typeDeclaredInDeclarationFile(
	packageName string,
	declarationFiles []*ast.SourceFile,
	program *compiler.Program,
) bool {
	matcher := packageMatcher(packageName)
	if matcher == nil {
		return false
	}
	for _, file := range declarationFiles {
		if file == nil || !program.IsSourceFileFromExternalLibrary(file) {
			continue
		}
		name := resolvedPackageName(file.FileName())
		if name != "" && Regexp2MatchString(matcher, name) {
			return true
		}
	}
	return false
}

// getImportModuleSpecifier traverses up from a declaration to find the import module specifier
func getImportModuleSpecifier(declaration *ast.Node) string {
	if declaration == nil {
		return ""
	}

	// Walk up to find ImportDeclaration
	current := declaration
	for current != nil {
		if ast.IsImportDeclaration(current) {
			moduleSpec := current.AsImportDeclaration().ModuleSpecifier
			if moduleSpec != nil && ast.IsStringLiteral(moduleSpec) {
				return moduleSpec.Text()
			}
			return ""
		}
		current = current.Parent
	}
	return ""
}

// typeDeclaredFromImport checks if any declaration comes from an import with the specified package name
func typeDeclaredFromImport(
	packageName string,
	declarations []*ast.Node,
) bool {
	for _, decl := range declarations {
		if decl == nil {
			continue
		}
		moduleSpec := getImportModuleSpecifier(decl)
		if moduleSpec == packageName {
			return true
		}
	}
	return false
}

func typeDeclaredInPackageDeclarationFile(
	packageName string,
	declarations []*ast.Node,
	declarationFiles []*ast.SourceFile,
	program *compiler.Program,
) bool {
	return typeDeclaredInDeclareModule(packageName, declarations) ||
		typeDeclaredFromImport(packageName, declarations) ||
		typeDeclaredInDeclarationFile(packageName, declarationFiles, program)
}

func typeMatchesSpecifier(
	t *checker.Type,
	specifier TypeOrValueSpecifier,
	program *compiler.Program,
	calleeNames []string,
) bool {
	// Handle union types: all constituents must match the specifier
	if parts := UnionTypeParts(t); len(parts) > 1 {
		return Every(parts, func(part *checker.Type) bool {
			return typeMatchesSpecifier(part, specifier, program, calleeNames)
		})
	}

	if !typeMatchesStringSpecifierWithCalleeNames(t, specifier.Name, calleeNames) {
		return false
	}

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
		return true // string shorthand matches any origin; name already matched above
	case TypeOrValueSpecifierFromFile:
		return typeDeclaredInFile(specifier.Path, declarationFiles, program)
	case TypeOrValueSpecifierFromLib:
		return typeDeclaredInLib(declarationFiles, program)
	case TypeOrValueSpecifierFromPackage:
		return typeDeclaredInPackageDeclarationFile(specifier.Package, declarations, declarationFiles, program)
	default:
		panic(fmt.Sprintf("unknown type specifier from: %v", specifier.From))
	}
}

func TypeMatchesSomeSpecifier(
	t *checker.Type,
	specifiers []TypeOrValueSpecifier,
	inlineSpecifiers []string,
	program *compiler.Program,
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
	program *compiler.Program,
	calleeNames []string,
) bool {
	for _, typePart := range IntersectionTypeParts(t) {
		if IsIntrinsicErrorType(typePart) {
			continue
		}
		if Some(specifiers, func(s TypeOrValueSpecifier) bool {
			return typeMatchesSpecifier(t, s, program, calleeNames)
		}) || typeMatchesStringSpecifierWithCalleeNames(t, inlineSpecifiers, calleeNames) {
			return true
		}
	}
	return false
}

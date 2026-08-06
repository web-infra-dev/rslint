package utils

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"sync"
	"weak"

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
	// pathProvided distinguishes an omitted path from an explicitly empty one.
	// Both decode to Path == "", but upstream treats only the omitted form as
	// matching declarations anywhere in the current project.
	pathProvided bool
}

// UnmarshalJSON handles both string shorthand ("Promise") and object form
// ({"from": "lib", "name": "Promise"}) for TypeOrValueSpecifier, matching
// the original TypeScript-ESLint TypeOrValueSpecifier union type.
func (s *TypeOrValueSpecifier) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		*s = TypeOrValueSpecifier{
			From: TypeOrValueSpecifierFromString,
			Name: NameList{str},
		}
		return nil
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	type Alias TypeOrValueSpecifier
	var alias Alias
	if err := json.Unmarshal(data, &alias); err != nil {
		return err
	}
	*s = TypeOrValueSpecifier(alias)
	_, s.pathProvided = fields["path"]
	return nil
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
	program *compiler.Program,
) bool {
	cwd := program.Host().GetCurrentDirectory()
	useCaseSensitiveFileNames := program.Host().FS().UseCaseSensitiveFileNames()
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

type sourceFilePackageNamesCacheEntry struct {
	once  sync.Once
	names map[tspath.Path][]string
}

// sourceFilePackageNamesCache reconstructs TypeScript's sourceFileToPackageName
// data from TypeScript-Go's resolution caches. Its weak Program keys keep
// repeated rule matches cheap without retaining old Programs after an LSP
// rebuild.
var sourceFilePackageNamesCache = struct {
	sync.Mutex
	entries map[weak.Pointer[compiler.Program]]*sourceFilePackageNamesCacheEntry
}{
	entries: make(map[weak.Pointer[compiler.Program]]*sourceFilePackageNamesCacheEntry),
}

func sourceFilePackageNames(program *compiler.Program) map[tspath.Path][]string {
	key := weak.Make(program)
	sourceFilePackageNamesCache.Lock()
	entry := sourceFilePackageNamesCache.entries[key]
	if entry == nil {
		// Weak keys do not remove themselves from a map. Opportunistically prune
		// entries whose Programs have been collected whenever a new one appears.
		for candidate := range sourceFilePackageNamesCache.entries {
			if candidate.Value() == nil {
				delete(sourceFilePackageNamesCache.entries, candidate)
			}
		}
		entry = &sourceFilePackageNamesCacheEntry{}
		sourceFilePackageNamesCache.entries[key] = entry
	}
	sourceFilePackageNamesCache.Unlock()

	entry.once.Do(func() {
		entry.names = make(map[tspath.Path][]string)
		addResolution := func(fileName string, packageID module.PackageId) {
			if packageID.Name == "" {
				return
			}
			file := program.GetSourceFileForResolvedModule(fileName)
			if file == nil {
				return
			}
			packageName := packageID.PackageName()
			if !slices.Contains(entry.names[file.Path()], packageName) {
				entry.names[file.Path()] = append(entry.names[file.Path()], packageName)
			}
		}
		for _, resolutions := range program.GetResolvedModules() {
			for _, resolution := range resolutions {
				if resolution != nil {
					addResolution(resolution.ResolvedFileName, resolution.PackageId)
				}
			}
		}
		for _, resolutions := range program.GetResolvedTypeReferenceDirectives() {
			for _, resolution := range resolutions {
				if resolution != nil {
					addResolution(resolution.ResolvedFileName, resolution.PackageId)
				}
			}
		}
	})
	return entry.names
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
		for _, name := range sourceFilePackageNames(program)[file.Path()] {
			if Regexp2MatchString(matcher, name) {
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
	program *compiler.Program,
) bool {
	return typeDeclaredInDeclareModule(packageName, declarations) ||
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
	program *compiler.Program,
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
	program *compiler.Program,
	t *checker.Type,
) bool {
	return Some(specifiers, func(s TypeOrValueSpecifier) bool {
		return valueMatchesSpecifier(node, s, program, t)
	})
}

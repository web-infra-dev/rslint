package utils

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/compiler"
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

func typeMatchesStringSpecifier(
	t *checker.Type,
	names []string,
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
	if relativePath == "" {
		return Some(declarationFiles, func(f *ast.SourceFile) bool {
			return strings.HasPrefix(f.FileName(), cwd)
		})
	}
	absPath := tspath.GetNormalizedAbsolutePath(relativePath, cwd)
	return Some(declarationFiles, func(f *ast.SourceFile) bool {
		return f.FileName() == absPath
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

func typeDeclaredInDeclarationFile(
	packageName string,
	declarationFiles []*ast.SourceFile,
	program *compiler.Program,
) bool {
	// typesPackageName := ""
	//  // Handle scoped packages: if the name starts with @, remove it and replace / with __
	// slashIndex := strings.Index(packageName, "/")
	// if packageName[0] == '@' && slashIndex >= 0 {
	// 	typesPackageName = packageName[1:slashIndex] + "__" + packageName[slashIndex+1:]
	// }

	// TODO(port): there is no sourceFileToPackageName anymore
	// it looks like there is no other way to know sourceFile2PackageName,
	// other than set package name for ast.SourceFile in resolver

	return false

	// const matcher = new RegExp(`${packageName}|${typesPackageName}`);
	// return declarationFiles.some(declaration => {
	//   const packageIdName = program.sourceFileToPackageName.get(declaration.path);
	//   return (
	//     packageIdName != null &&
	//     matcher.test(packageIdName) &&
	//     program.isSourceFileFromExternalLibrary(declaration)
	//   );
	// });
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
) bool {
	if !typeMatchesStringSpecifier(t, specifier.Name) {
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
	for _, typePart := range IntersectionTypeParts(t) {
		if IsIntrinsicErrorType(typePart) {
			continue
		}
		if Some(specifiers, func(s TypeOrValueSpecifier) bool {
			return typeMatchesSpecifier(t, s, program)
		}) || typeMatchesStringSpecifier(t, inlineSpecifiers) {
			return true
		}
	}
	return false
}

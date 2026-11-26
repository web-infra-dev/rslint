package typeservice

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
)

func TestJSDocDocumentationAndInheritanceWithAndWithoutChecker(t *testing.T) {
	p, _, root := buildTestProgram(t, map[string]string{
		"docs.ts": `
			export function replacement() {}
			export interface Base {
				/** Base docs linking to {@link replacement}.
				 * @deprecated Use replacement instead.
				 */
				run(): void;
			}
			export class Derived implements Base {
				/** @inheritDoc Extra detail. */
				run() {}
			}
		`,
	}, []string{"docs.ts"})
	file := p.GetSourceFile(tspath.ResolvePath(root, "docs.ts"))
	if file == nil {
		t.Fatal("missing fixture file")
	}

	c, release := p.TypeCheckerForFileExclusive(core.WithCheckerLifetime(t.Context(), core.CheckerLifetimeAPI), file)
	defer release()

	var derivedMethod *ast.Node
	for _, stmt := range file.Statements.Nodes {
		if stmt.Kind == ast.KindClassDeclaration && stmt.Name().Text() == "Derived" {
			derivedMethod = stmt.Members()[0]
		}
	}
	if derivedMethod == nil {
		t.Fatal("Derived.run not found")
	}
	sym := c.GetSymbolAtLocation(derivedMethod.Name())
	if sym == nil {
		t.Fatal("symbol for Derived.run not found")
	}
	decls, _ := symbolDeclarations(sym, false)

	// Without checker: no @inheritDoc traversal, no {@link} resolution.
	docsNoChecker := documentation(nil, decls, map[*ast.Symbol]bool{sym: true})
	tagsNoChecker := jsDocTags(nil, decls, map[*ast.Symbol]bool{sym: true})
	for _, part := range docsNoChecker {
		if part.Kind == "linkName" {
			t.Fatalf("unexpected resolved linkName without checker: %+v", docsNoChecker)
		}
	}
	for _, tag := range tagsNoChecker {
		if tag.Name == "deprecated" {
			t.Fatalf("unexpected inherited @deprecated tag without checker: %+v", tagsNoChecker)
		}
	}

	// With checker: inherits Base docs + @deprecated tag + resolves {@link replacement}.
	docsWithChecker := documentation(c, decls, map[*ast.Symbol]bool{sym: true})
	tagsWithChecker := jsDocTags(c, decls, map[*ast.Symbol]bool{sym: true})
	hasLinkTarget := false
	for _, part := range docsWithChecker {
		if part.Kind == "linkName" && part.Target != nil && part.Target.File == file.FileName() {
			hasLinkTarget = true
		}
	}
	if !hasLinkTarget {
		t.Fatalf("expected resolved linkName target with checker, got %+v", docsWithChecker)
	}
	hasDeprecated := false
	for _, tag := range tagsWithChecker {
		if tag.Name == "deprecated" {
			hasDeprecated = true
		}
	}
	if !hasDeprecated {
		t.Fatalf("expected inherited @deprecated tag with checker, got %+v", tagsWithChecker)
	}
}

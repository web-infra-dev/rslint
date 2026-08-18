package program

import (
	"context"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/core"
)

// collectNoEmitDiagnostics mirrors compiler.GetDiagnosticsOfAnyProgram while
// enforcing tsc --noEmit semantics. It stays behind Program's private type
// adapter so linter and rules never unwrap a backend.
//
// Calling GetDiagnosticsOfAnyProgram directly is not equivalent. tsc injects
// NoEmit at its command-line layer, before compiler-option diagnostics are
// cached; rslint receives the user's original tsconfig. Invisible noEmit-gated
// option errors can therefore enter ProgramDiagnostics and short-circuit the
// semantic work even though rslint later drops those diagnostics. This helper
// removes diagnostics that cannot survive rslint's output boundary before the
// short-circuit, filters semantic diagnostics under an explicit NoEmit clone,
// and retains declaration diagnostics under the same policy.
func collectNoEmitDiagnostics(ctx context.Context, raw *compiler.Program) []*ast.Diagnostic {
	noEmitOptions := raw.Options().Clone()
	noEmitOptions.NoEmit = core.TSTrue
	configFilePath := raw.Options().ConfigFilePath

	keep := func(diagnostics []*ast.Diagnostic) []*ast.Diagnostic {
		return filterShortCircuitDiagnostics(diagnostics, configFilePath)
	}

	configDiagnostics := keep(raw.GetConfigFileParsingDiagnostics())
	baseline := len(configDiagnostics)
	all := append([]*ast.Diagnostic(nil), configDiagnostics...)
	all = append(all, keep(raw.GetSyntacticDiagnostics(ctx, nil))...)
	all = append(all, keep(raw.GetProgramDiagnostics())...)
	if len(all) != baseline {
		return all
	}

	raw.GetBindDiagnostics(ctx, nil)
	if raw.Options().ListFilesOnly.IsTrue() {
		return all
	}

	all = append(all, raw.GetGlobalDiagnostics(ctx)...)
	if len(all) == baseline {
		semantic := compiler.FilterNoEmitSemanticDiagnostics(raw.GetSemanticDiagnostics(ctx, nil), noEmitOptions)
		all = append(all, semantic...)
		all = append(all, raw.GetGlobalDiagnostics(ctx)...)
	}
	if noEmitOptions.GetEmitDeclarations() && len(all) == baseline {
		all = append(all, raw.GetDeclarationDiagnostics(ctx, nil)...)
	}
	return all
}

func filterShortCircuitDiagnostics(diagnostics []*ast.Diagnostic, configFilePath string) []*ast.Diagnostic {
	filtered := make([]*ast.Diagnostic, 0, len(diagnostics))
	for _, diagnostic := range diagnostics {
		file := diagnostic.File()
		if file == nil || configFilePath != "" && file.FileName() == configFilePath {
			continue
		}
		filtered = append(filtered, diagnostic)
	}
	return filtered
}

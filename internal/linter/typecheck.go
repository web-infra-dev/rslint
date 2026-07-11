package linter

import (
	"context"
	"fmt"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// typeCheckRequest bundles the inputs for runTypeCheckAcrossPrograms.
type typeCheckRequest struct {
	Programs       []*compiler.Program
	Skip           []bool // parallel to Programs; nil → check all
	SingleThreaded bool
	OnDiagnostic   DiagnosticHandler
}

// runTypeCheckAcrossPrograms runs program-scoped semantic diagnostics
// against every program except those explicitly opted out.
//
// Aggregates diagnostics via collectNoEmitDiagnostics, which mirrors
// compiler.GetDiagnosticsOfAnyProgram(file=nil) but enforces tsc --noEmit
// semantics regardless of whether the user's tsconfig sets noEmit.
// It covers config-parsing, syntactic, program, global, semantic, and
// (when GetEmitDeclarations() is set) declaration diagnostics.
// @ts-ignore / @ts-expect-error suppression and TS2578 unused-directive
// errors are handled inside typescript-go.
//
// Cross-program dedupe: shared declaration files (typical with project
// references and pulled-in node_modules .d.ts) appear in multiple programs.
// We dedupe on canonical (filePath, code, pos, end, final message), including
// message chains and related information.
//
// req.OnDiagnostic must be non-nil and safe to call from multiple
// goroutines concurrently — RunLinter is responsible for that contract.
func runTypeCheckAcrossPrograms(req typeCheckRequest) {
	collected := make([][]collectedTypeCheckDiagnostic, len(req.Programs))
	wg := core.NewWorkGroup(req.SingleThreaded)
	for i, prog := range req.Programs {
		if i < len(req.Skip) && req.Skip[i] {
			continue
		}
		programIndex := i
		program := prog
		wg.Queue(func() {
			collected[programIndex] = runTypeCheckForProgram(program)
		})
	}
	wg.RunAndWait()

	// Program diagnostics are computed in parallel above, but survivor choice
	// follows stable Program input order rather than goroutine completion order.
	emitDeduplicatedTypeCheckDiagnostics(collected, req.OnDiagnostic)
}

func emitDeduplicatedTypeCheckDiagnostics(collected [][]collectedTypeCheckDiagnostic, onDiagnostic DiagnosticHandler) {
	seen := make(map[typeCheckDedupeKey]struct{})
	for _, programDiagnostics := range collected {
		for _, diagnostic := range programDiagnostics {
			if _, duplicate := seen[diagnostic.key]; duplicate {
				continue
			}
			seen[diagnostic.key] = struct{}{}
			onDiagnostic(diagnostic.ruleDiagnostic)
		}
	}
}

type collectedTypeCheckDiagnostic struct {
	key            typeCheckDedupeKey
	ruleDiagnostic rule.RuleDiagnostic
}

func runTypeCheckForProgram(prog *compiler.Program) []collectedTypeCheckDiagnostic {
	ctx := context.Background()
	diags := collectNoEmitDiagnostics(ctx, prog)
	collected := make([]collectedTypeCheckDiagnostic, 0, len(diags))

	for _, d := range diags {
		file := d.File()
		// Diagnostics without an attached file have no source location to
		// render against. Examples: TS18003 ("No inputs were found in
		// config file"), TS5108 ("Option ... has been removed"), TS2318
		// ("Cannot find global type") — all surface from
		// GetDiagnosticsOfAnyProgram with d.File()==nil.
		//
		// rslint --type-check intentionally drops these — see
		// "Diagnostics without a source location" in
		// website/docs/en/guide/type-checking.md. The behaviour is
		// pinned by TestBoundary_NoSourceLocationDiagnosticsDropped.
		if file == nil {
			continue
		}

		// Suppression for @ts-ignore / @ts-expect-error / @ts-nocheck and
		// the corresponding TS2578 unused-directive errors is already
		// handled inside typescript-go (GetDiagnosticsOfAnyProgram). We do
		// not add an rslint-specific disable channel here on purpose:
		// type-check is meant to be a transparent passthrough of tsc.

		message := flattenDiagnosticMessage(d)
		loc := d.Loc()
		collected = append(collected, collectedTypeCheckDiagnostic{
			key: typeCheckDedupeKeyForDiagnostic(prog, d),
			ruleDiagnostic: rule.RuleDiagnostic{
				RuleName:     fmt.Sprintf("TypeScript(TS%d)", d.Code()),
				Range:        loc,
				Message:      rule.RuleMessage{Description: message},
				SourceFile:   file,
				FilePath:     file.FileName(),
				Severity:     rule.SeverityError,
				PreFormatted: true,
			},
		})
	}
	return collected
}

type typeCheckDedupeKey struct {
	path    string
	code    int32
	pos     int
	end     int
	message string
}

func typeCheckDedupeKeyForDiagnostic(prog *compiler.Program, d *ast.Diagnostic) typeCheckDedupeKey {
	loc := d.Loc()
	return typeCheckDedupeKey{
		path:    typeCheckFilesystemPathID(prog, d.File().FileName()),
		code:    d.Code(),
		pos:     loc.Pos(),
		end:     loc.End(),
		message: flattenDiagnosticMessageForIdentity(prog, d),
	}
}

func typeCheckFilesystemPathID(prog *compiler.Program, filePath string) string {
	filePath = tspath.NormalizePath(filePath)
	if prog == nil {
		return filePath
	}
	filePath = tspath.GetNormalizedAbsolutePath(filePath, prog.GetCurrentDirectory())
	fsys := prog.Host().FS()
	if fsys != nil {
		if realPath := fsys.Realpath(filePath); realPath != "" {
			filePath = tspath.NormalizePath(realPath)
		}
	}
	return string(tspath.ToPath(filePath, "", true))
}

// flattenDiagnosticMessage builds a human-readable message from a
// TypeScript diagnostic, including its MessageChain and
// RelatedInformation. Format mirrors tsc's output.
func flattenDiagnosticMessage(d *ast.Diagnostic) string {
	return flattenDiagnosticMessageWithRelatedPath(d, func(file *ast.SourceFile) string {
		return file.FileName()
	})
}

func flattenDiagnosticMessageForIdentity(prog *compiler.Program, d *ast.Diagnostic) string {
	return flattenDiagnosticMessageWithRelatedPath(d, func(file *ast.SourceFile) string {
		return typeCheckFilesystemPathID(prog, file.FileName())
	})
}

func flattenDiagnosticMessageWithRelatedPath(d *ast.Diagnostic, relatedPath func(*ast.SourceFile) string) string {
	var b strings.Builder
	b.WriteString(d.String())
	for _, chain := range d.MessageChain() {
		flattenMessageChain(&b, chain, 1)
	}
	for _, related := range d.RelatedInformation() {
		if related.File() != nil {
			line, _ := scanner.GetECMALineAndUTF16CharacterOfPosition(related.File(), related.Pos())
			fmt.Fprintf(&b, "\n  %s:%d: %s", relatedPath(related.File()), line+1, related.String())
		}
	}
	return b.String()
}

func flattenMessageChain(b *strings.Builder, chain *ast.Diagnostic, level int) {
	b.WriteByte('\n')
	for range level {
		b.WriteString("  ")
	}
	b.WriteString(chain.String())
	for _, child := range chain.MessageChain() {
		flattenMessageChain(b, child, level+1)
	}
}

// collectNoEmitDiagnostics aggregates program-scoped diagnostics in the
// same shape as compiler.GetDiagnosticsOfAnyProgram(file=nil) but adjusted
// to behave as if --noEmit had been on the command line.
//
// Why we don't just call GetDiagnosticsOfAnyProgram directly:
//
//   - tsc --noEmit injects NoEmit=true at the command-line layer, so when
//     verifyCompilerOptions runs during program construction it never
//     produces noEmit-gated option errors (e.g. TS5096 for
//     allowImportingTsExtensions, TS5055 for output overwriting input).
//   - rslint cannot inject anything at that layer: it just receives the
//     user's tsconfig as-is, so those option errors land in the program's
//     cached programDiagnostics. They are anchored to the tsconfig file
//     (or have no file at all) — either way, rslint's downstream filters
//     drop them, so the user never sees them.
//   - GetDiagnosticsOfAnyProgram short-circuits binding / global / semantic
//     diagnostic collection whenever config-parsing + syntactic + program
//     diagnostics exceed the config-parsing baseline. Because rslint
//     drops the option errors downstream, the user gets neither the
//     option error nor any of the real semantic errors — silent failure.
//
// This function reproduces the GetDiagnosticsOfAnyProgram contract but:
//
//  1. Strips diagnostics that would not survive runTypeCheckForProgram's
//     downstream filters (nil-file and tsconfig-anchored option diagnostics)
//     before applying the short-circuit. This keeps invisible option errors
//     from masking real semantic work. File-anchored syntactic / program
//     errors that do reach the user still short-circuit semantic collection,
//     matching tsc's behaviour for real pre-semantic errors.
//  2. Re-applies compiler.FilterNoEmitSemanticDiagnostics over the
//     semantic diagnostics with NoEmit=true, so emit-only checks
//     (SkippedOnNoEmit, e.g. __esModule reservation errors) drop out
//     even when the user's tsconfig leaves NoEmit unset.
//  3. Collects declaration diagnostics under the noEmit branch, matching
//     tsc --noEmit's behaviour when GetEmitDeclarations() is set.
func collectNoEmitDiagnostics(ctx context.Context, prog *compiler.Program) []*ast.Diagnostic {
	noEmitOpts := prog.Options().Clone()
	noEmitOpts.NoEmit = core.TSTrue
	configFilePath := prog.Options().ConfigFilePath

	keep := func(in []*ast.Diagnostic) []*ast.Diagnostic {
		return filterShortCircuitDiagnostics(in, configFilePath)
	}

	configDiags := keep(prog.GetConfigFileParsingDiagnostics())
	baseline := len(configDiags)
	all := append([]*ast.Diagnostic(nil), configDiags...)

	all = append(all, keep(prog.GetSyntacticDiagnostics(ctx, nil))...)
	all = append(all, keep(prog.GetProgramDiagnostics())...)

	if len(all) != baseline {
		return all
	}

	// Match GetDiagnosticsOfAnyProgram: bind early so its time is tracked
	// separately; do not aggregate bind diagnostics.
	prog.GetBindDiagnostics(ctx, nil)

	if prog.Options().ListFilesOnly.IsTrue() {
		return all
	}

	all = append(all, prog.GetGlobalDiagnostics(ctx)...)
	if len(all) == baseline {
		semantic := compiler.FilterNoEmitSemanticDiagnostics(prog.GetSemanticDiagnostics(ctx, nil), noEmitOpts)
		all = append(all, semantic...)
		// Globals can grow once the checker pulls in missing types — re-collect.
		all = append(all, prog.GetGlobalDiagnostics(ctx)...)
	}
	if noEmitOpts.GetEmitDeclarations() && len(all) == baseline {
		all = append(all, prog.GetDeclarationDiagnostics(ctx, nil)...)
	}
	return all
}

// filterShortCircuitDiagnostics returns the subset of diagnostics that should
// participate in GetDiagnosticsOfAnyProgram's short-circuit check. We exclude:
//
//   - diagnostics with no source file (rslint always drops these downstream),
//   - diagnostics anchored to the program's tsconfig file (the implicit form
//     of "fileless" — they record an option-level problem with the tsconfig
//     itself, not with the user's source code, and rslint --type-check is
//     defined to mirror tsc --noEmit, which silently suppresses noEmit-gated
//     option errors when --noEmit is on the command line),
//
// Diagnostics anchored to user source files (real syntactic / program errors)
// flow through unchanged, so they still trip the short-circuit — matching tsc.
func filterShortCircuitDiagnostics(in []*ast.Diagnostic, configFilePath string) []*ast.Diagnostic {
	out := make([]*ast.Diagnostic, 0, len(in))
	for _, d := range in {
		f := d.File()
		if f == nil {
			continue
		}
		if configFilePath != "" && f.FileName() == configFilePath {
			continue
		}
		out = append(out, d)
	}
	return out
}

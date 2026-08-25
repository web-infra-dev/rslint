package linter

import (
	"context"
	"fmt"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// typeCheckRequest bundles the inputs for runTypeCheckAcrossPrograms.
type typeCheckRequest struct {
	Programs       []*program.Program
	SingleThreaded bool
	OnDiagnostic   DiagnosticHandler
}

// runTypeCheckAcrossPrograms runs program-scoped semantic diagnostics
// against every Program that exposes that capability. The source construction
// strategy remains private; the linter schedules only the capability it needs.
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
		if !prog.CanProvideProgramDiagnostics() {
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

func runTypeCheckForProgram(prog *program.Program) []collectedTypeCheckDiagnostic {
	ctx := context.Background()
	diags := prog.NoEmitDiagnostics(ctx)
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
		collected = append(collected, collectedTypeCheckDiagnostic{
			key:            typeCheckDedupeKeyForDiagnostic(prog, d),
			ruleDiagnostic: newTypeScriptDiagnostic(file, d, message),
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

func typeCheckDedupeKeyForDiagnostic(prog *program.Program, d *ast.Diagnostic) typeCheckDedupeKey {
	loc := d.Loc()
	return typeCheckDedupeKey{
		path:    typeCheckFilesystemPathID(prog, d.File().FileName()),
		code:    d.Code(),
		pos:     loc.Pos(),
		end:     loc.End(),
		message: flattenDiagnosticMessageForIdentity(prog, d),
	}
}

func typeCheckFilesystemPathID(prog *program.Program, filePath string) string {
	filePath = tspath.NormalizePath(filePath)
	if prog == nil {
		return filePath
	}
	filePath = tspath.GetNormalizedAbsolutePath(filePath, prog.CurrentDirectory())
	fsys := prog.FS()
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

func flattenDiagnosticMessageForIdentity(prog *program.Program, d *ast.Diagnostic) string {
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

package linter

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// typeCheckRequest bundles the inputs for runTypeCheckAcrossPrograms.
type typeCheckRequest struct {
	Programs       []*compiler.Program
	Skip           []bool // parallel to Programs; nil → check all
	SingleThreaded bool
	// TypeInfoFiles, when non-nil, restricts diagnostic output to files in
	// the set. This honors the gap-file commitment in type-checking.md:
	// files matched by config `files` but not covered by any tsconfig
	// (e.g. root scripts, config files) do not receive semantic errors.
	// nil = no gap-file distinction.
	TypeInfoFiles map[string]struct{}
	OnDiagnostic  DiagnosticHandler
}

// runTypeCheckAcrossPrograms runs program-scoped semantic diagnostics
// against every program except those explicitly opted out.
//
// Uses compiler.GetDiagnosticsOfAnyProgram(file=nil) — tsc's own aggregate.
// It covers config-parsing, syntactic, program, global, and semantic
// diagnostics, with NoEmit filtering, @ts-ignore / @ts-expect-error
// suppression, and TS2578 unused-directive errors handled internally.
//
// Cross-program dedupe: shared declaration files (typical with project
// references and pulled-in node_modules .d.ts) appear in multiple programs.
// We dedupe on (filePath, code, pos, end, message).
//
// req.OnDiagnostic must be non-nil and safe to call from multiple
// goroutines concurrently — RunLinter is responsible for that contract.
func runTypeCheckAcrossPrograms(req typeCheckRequest) {
	var seen sync.Map
	wg := core.NewWorkGroup(req.SingleThreaded)
	for i, prog := range req.Programs {
		if i < len(req.Skip) && req.Skip[i] {
			continue
		}
		wg.Queue(func() {
			runTypeCheckForProgram(prog, &seen, req.TypeInfoFiles, req.OnDiagnostic)
		})
	}
	wg.RunAndWait()
}

func runTypeCheckForProgram(prog *compiler.Program, seen *sync.Map, typeInfoFiles map[string]struct{}, onDiagnostic DiagnosticHandler) {
	ctx := context.Background()
	diags := compiler.GetDiagnosticsOfAnyProgram(
		ctx, prog, nil, false,
		prog.GetBindDiagnostics,
		prog.GetSemanticDiagnostics,
	)

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

		// Gap-file gate: when TypeInfoFiles is supplied, drop diagnostics
		// for files outside it (matches the commitment in
		// type-checking.md).
		if typeInfoFiles != nil {
			if _, ok := typeInfoFiles[file.FileName()]; !ok {
				continue
			}
		}

		if !markSeen(seen, file, d) {
			continue
		}

		// Suppression for @ts-ignore / @ts-expect-error / @ts-nocheck and
		// the corresponding TS2578 unused-directive errors is already
		// handled inside typescript-go (GetDiagnosticsOfAnyProgram). We do
		// not add an rslint-specific disable channel here on purpose:
		// type-check is meant to be a transparent passthrough of tsc.

		onDiagnostic(rule.RuleDiagnostic{
			RuleName:     fmt.Sprintf("TypeScript(TS%d)", d.Code()),
			Range:        d.Loc(),
			Message:      rule.RuleMessage{Description: flattenDiagnosticMessage(d)},
			SourceFile:   file,
			Severity:     rule.SeverityError,
			PreFormatted: true,
		})
	}
}

type typeCheckDedupeKey struct {
	path    string
	code    int32
	pos     int
	end     int
	message string
}

// markSeen records (filePath, code, pos, end, message) and reports whether
// this exact tuple has been seen before. Callers must pass a non-nil
// `file` — runTypeCheckForProgram has already filtered out file=nil
// diagnostics before reaching here.
func markSeen(seen *sync.Map, file *ast.SourceFile, d *ast.Diagnostic) bool {
	loc := d.Loc()
	key := typeCheckDedupeKey{
		path:    file.FileName(),
		code:    d.Code(),
		pos:     loc.Pos(),
		end:     loc.End(),
		message: d.String(),
	}
	_, dup := seen.LoadOrStore(key, struct{}{})
	return !dup
}

// flattenDiagnosticMessage builds a human-readable message from a
// TypeScript diagnostic, including its MessageChain and
// RelatedInformation. Format mirrors tsc's output.
func flattenDiagnosticMessage(d *ast.Diagnostic) string {
	var b strings.Builder
	b.WriteString(d.String())
	for _, chain := range d.MessageChain() {
		flattenMessageChain(&b, chain, 1)
	}
	for _, related := range d.RelatedInformation() {
		if related.File() != nil {
			line, _ := scanner.GetECMALineAndUTF16CharacterOfPosition(related.File(), related.Pos())
			fmt.Fprintf(&b, "\n  %s:%d: %s", related.File().FileName(), line+1, related.String())
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


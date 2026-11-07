package linter

import (
	"sync"

	"github.com/web-infra-dev/rslint/internal/rule"
)

// fileDirectives belongs to one execution round. It never retains syntax or
// a generation lease; native and plugin completion can arrive in either order.
type fileDirectives struct {
	manager                                   *rule.DisableManager
	path                                      string
	severity                                  rule.DiagnosticSeverity
	demand                                    rule.EditDemand
	ignored                                   map[string]bool
	mu                                        sync.Mutex
	pendingPlugins                            int
	nativeDone, pluginDone, failed, finalized bool
}

func (f *fileDirectives) complete(native, failed bool, path string) []rule.RuleDiagnostic {
	if f == nil {
		return nil
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.failed = f.failed || failed
	if native {
		f.nativeDone = true
	} else {
		if f.pendingPlugins > 0 {
			f.pendingPlugins--
		}
		f.pluginDone = f.pendingPlugins == 0
	}
	if !f.nativeDone || !f.pluginDone || f.finalized {
		return nil
	}
	f.finalized = true
	if f.failed {
		return nil
	}
	return f.manager.UnusedDiagnostics(path, f.severity, f.demand, f.ignored)
}

func prepareDirectiveStates(plan *LintPlan, generation Generation, policy ObservationPolicy) map[string]*fileDirectives {
	if plan == nil {
		return nil
	}
	var states map[string]*fileDirectives
	for _, program := range plan.programs {
		for _, file := range program.files {
			source := file.file
			// Avoid scanning comments or retaining state on the ordinary no-directive path.
			if !file.hasDirectiveCandidate {
				continue
			}
			manager := rule.NewDisableManager(source, rule.NewCommentStore(source))
			if generation.Native.DirectiveRuleExists != nil {
				manager.ValidateRules(func(name string) bool { return generation.Native.DirectiveRuleExists(source.FileName(), name) })
			}
			manager.Freeze(diagnosticTextSource(source))
			severity := rule.SeverityOff
			if generation.Native.UnusedDirectiveSeverity != nil {
				severity = generation.Native.UnusedDirectiveSeverity(source.FileName())
			}
			f := &fileDirectives{manager: manager, path: projectTargetPath(generation.Target.Path, source.FileName()), severity: severity, demand: policy.Demand.Native | policy.Demand.Plugin, pluginDone: file.rules == nil || len(file.rules.plugins) == 0, ignored: file.skippedRules}
			if generation.Native.SkippedDirectiveRules != nil {
				for _, name := range generation.Native.SkippedDirectiveRules(source.FileName()) {
					if f.ignored == nil {
						f.ignored = map[string]bool{}
					}
					f.ignored[name] = true
				}
			}
			if states == nil {
				states = make(map[string]*fileDirectives)
			}
			states[source.FileName()] = f
		}
	}
	for _, group := range plan.syntacticDiagnosticGroups {
		for _, diagnostic := range group.diagnostics {
			if f := states[diagnostic.FilePath]; f != nil {
				f.failed = true
			}
		}
	}
	return states
}

type directiveStates map[string]*fileDirectives

// Package projectselection owns the entry-neutral policy that binds lint
// targets to configured TypeScript projects. Callers retain responsibility for
// target discovery, Program lifetime, caching, and concurrency.
package projectselection

import "fmt"

const NoProject = -1

type Tier uint8

const (
	TierNone Tier = iota
	TierDirect
	TierImport
)

// Target carries the two path identities established before project selection.
type Target struct {
	Path          string
	CanonicalPath string
}

// Plan is one config owner's target set and parserOptions.project declaration
// order. Project indexes are adapter-owned stable identities and may be shared
// by multiple independent owner plans.
type Plan struct {
	Targets  []Target
	Projects []int
}

type Binding struct {
	Project int
	Tier    Tier
}

// Metadata is the immutable project evidence needed before constructing a
// Program. DirectRoot must use exact lexical identity plus authoritative
// canonical identity; Supports applies the project's compiler options.
type Metadata interface {
	DirectRoot(target Target) bool
	Supports(target Target) bool
}

type LoadMetadata func(project int) (metadata Metadata, available bool, err error)

// ScheduleProject may start constructing a confirmed direct winner. It must
// not await or publish the construction result; LoadProject owns that logical
// error frontier.
type ScheduleProject func(project int)
type LoadProject func(project int) (available bool, err error)
type ContainsTarget func(project int, target Target) bool

// DirectProjectUnavailableError means metadata declared a target as a root but
// its Program provider could not supply that project generation.
type DirectProjectUnavailableError struct {
	Project int
	Target  Target
}

func (err *DirectProjectUnavailableError) Error() string {
	return fmt.Sprintf("direct project %d is unavailable for %q", err.Project, err.Target.Path)
}

// DirectRootAbsentError means the authoritative metadata declared a direct
// root that the constructed Program did not contain.
type DirectRootAbsentError struct {
	Project int
	Target  Target
}

func (err *DirectRootAbsentError) Error() string {
	return fmt.Sprintf("direct project %d did not contain root %q", err.Project, err.Target.Path)
}

type metadataEvidence struct {
	metadata  Metadata
	available bool
}

// Selection is the stable state between the metadata-only direct phase and
// Program-backed validation/import fallback. This split lets request loaders
// schedule confirmed direct builds without leaking scheduler policy into the
// ownership state machine.
type Selection struct {
	plan              Plan
	bindings          []Binding
	metadataByProject map[int]metadataEvidence
	loadedProjects    map[int]bool
	directValidated   []bool
}

func ResolveDirect(
	plan Plan,
	loadMetadata LoadMetadata,
	schedule ScheduleProject,
) (*Selection, error) {
	selection := &Selection{
		plan: Plan{
			Targets:  append([]Target(nil), plan.Targets...),
			Projects: append([]int(nil), plan.Projects...),
		},
		bindings:          make([]Binding, len(plan.Targets)),
		metadataByProject: make(map[int]metadataEvidence, len(plan.Projects)),
		loadedProjects:    make(map[int]bool),
		directValidated:   make([]bool, len(plan.Targets)),
	}
	for index := range selection.bindings {
		selection.bindings[index].Project = NoProject
	}
	if len(plan.Targets) == 0 || len(plan.Projects) == 0 || loadMetadata == nil {
		return selection, nil
	}

	unresolved := len(plan.Targets)
	scheduled := make(map[int]struct{})
	for _, project := range plan.Projects {
		evidence, loaded := selection.metadataByProject[project]
		if !loaded {
			metadata, available, err := loadMetadata(project)
			if err != nil {
				return nil, err
			}
			evidence = metadataEvidence{metadata: metadata, available: available && metadata != nil}
			selection.metadataByProject[project] = evidence
		}
		if !evidence.available {
			continue
		}

		selected := false
		for targetIndex, target := range plan.Targets {
			if selection.bindings[targetIndex].Tier != TierNone || !evidence.metadata.DirectRoot(target) {
				continue
			}
			selection.bindings[targetIndex] = Binding{Project: project, Tier: TierDirect}
			unresolved--
			selected = true
		}
		if selected && schedule != nil {
			if _, exists := scheduled[project]; !exists {
				schedule(project)
				scheduled[project] = struct{}{}
			}
		}
		if unresolved == 0 {
			break
		}
	}
	return selection, nil
}

// DirectTarget identifies one target within an independently resolved owner
// plan. Batch adapters may interleave these references in their existing
// request order without merging the plans' declaration orders.
type DirectTarget struct {
	Selection *Selection
	Index     int
}

// ValidateDirectTargets preserves the two-stage direct validation frontier:
// every selected Program is loaded in reference order before any declared root
// is checked in that same order. This lets batch adapters retain their existing
// cross-owner error order while every owner still resolves independently.
func ValidateDirectTargets(
	targets []DirectTarget,
	load LoadProject,
	contains ContainsTarget,
) error {
	for _, target := range targets {
		selection := target.Selection
		if selection == nil || target.Index < 0 || target.Index >= len(selection.bindings) ||
			selection.bindings[target.Index].Tier != TierDirect ||
			selection.directValidated[target.Index] {
			continue
		}
		binding := selection.bindings[target.Index]
		available, err := selection.loadProject(binding.Project, load)
		if err != nil {
			return err
		}
		if !available {
			return &DirectProjectUnavailableError{
				Project: binding.Project,
				Target:  selection.plan.Targets[target.Index],
			}
		}
	}
	for _, target := range targets {
		selection := target.Selection
		if selection == nil || target.Index < 0 || target.Index >= len(selection.bindings) ||
			selection.bindings[target.Index].Tier != TierDirect ||
			selection.directValidated[target.Index] {
			continue
		}
		binding := selection.bindings[target.Index]
		selectedTarget := selection.plan.Targets[target.Index]
		if contains == nil || !contains(binding.Project, selectedTarget) {
			return &DirectRootAbsentError{Project: binding.Project, Target: selectedTarget}
		}
		selection.directValidated[target.Index] = true
	}
	return nil
}

func (selection *Selection) loadProject(project int, load LoadProject) (bool, error) {
	if available, loaded := selection.loadedProjects[project]; loaded {
		return available, nil
	}
	if load == nil {
		selection.loadedProjects[project] = false
		return false, nil
	}
	available, err := load(project)
	if err != nil {
		return false, err
	}
	selection.loadedProjects[project] = available
	return available, nil
}

// ValidateDirect observes Program construction in target order, matching the
// existing batch loader error frontier, then verifies every declared root.
func (selection *Selection) ValidateDirect(
	load LoadProject,
	contains ContainsTarget,
) error {
	if selection == nil {
		return nil
	}
	targets := make([]DirectTarget, 0, len(selection.bindings))
	for targetIndex := range selection.bindings {
		targets = append(targets, DirectTarget{Selection: selection, Index: targetIndex})
	}
	return ValidateDirectTargets(targets, load, contains)
}

// Complete applies the second tier only to targets with no direct owner. It
// visits projects and exposes load errors strictly in declaration order.
func (selection *Selection) Complete(
	load LoadProject,
	contains ContainsTarget,
) ([]Binding, error) {
	if selection == nil {
		return nil, nil
	}
	if err := selection.ValidateDirect(load, contains); err != nil {
		return nil, err
	}

	pending := len(selection.bindings)
	for _, binding := range selection.bindings {
		if binding.Tier != TierNone {
			pending--
		}
	}
	if pending == 0 {
		return selection.Bindings(), nil
	}

	for _, project := range selection.plan.Projects {
		evidence := selection.metadataByProject[project]
		eligible := make([]int, 0, pending)
		for targetIndex, binding := range selection.bindings {
			if binding.Tier != TierNone {
				continue
			}
			if evidence.available && !evidence.metadata.Supports(selection.plan.Targets[targetIndex]) {
				continue
			}
			eligible = append(eligible, targetIndex)
		}
		if len(eligible) == 0 {
			continue
		}

		available, err := selection.loadProject(project, load)
		if err != nil {
			return nil, err
		}
		if !available {
			continue
		}
		for _, targetIndex := range eligible {
			target := selection.plan.Targets[targetIndex]
			if contains != nil && contains(project, target) {
				selection.bindings[targetIndex] = Binding{Project: project, Tier: TierImport}
				pending--
			}
		}
		if pending == 0 {
			break
		}
	}
	return selection.Bindings(), nil
}

func (selection *Selection) Bindings() []Binding {
	if selection == nil {
		return nil
	}
	return append([]Binding(nil), selection.bindings...)
}

func Resolve(
	plan Plan,
	loadMetadata LoadMetadata,
	loadProject LoadProject,
	contains ContainsTarget,
) ([]Binding, error) {
	selection, err := ResolveDirect(plan, loadMetadata, nil)
	if err != nil {
		return nil, err
	}
	return selection.Complete(loadProject, contains)
}

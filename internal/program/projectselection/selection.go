// Package projectselection owns the entry-neutral policy that binds lint
// targets to configured TypeScript projects. Callers retain responsibility for
// config resolution, filesystem evidence, Program lifetime, caching, and
// concurrency.
package projectselection

import "fmt"

const NoProject = -1

type Tier uint8

const (
	TierNone Tier = iota
	TierDirect
	TierImport
)

// Target carries the two source identities established by discovery and its
// own ordered project declarations. Project indexes are adapter-owned exact
// lexical identities and may be shared by unrelated targets.
type Target struct {
	Path          string
	CanonicalPath string
	Projects      []int
}

// Plan is one complete lint request. Target order, followed by each target's
// project declaration order, defines the observable metadata/Program error
// frontier. All direct-root decisions finish before import fallback starts.
type Plan struct {
	Targets []Target
}

type Binding struct {
	Project int
	Tier    Tier
}

// Metadata is the immutable project evidence available without constructing a
// Program. DirectRoot must use exact lexical identity plus authoritative
// canonical identity. Supports applies the parsed compiler options.
type Metadata interface {
	DirectRoot(target Target) bool
	Supports(target Target) bool
}

type LoadMetadata func(project int) (metadata Metadata, available bool, err error)

// ScheduleProject may start constructing a confirmed direct winner. It must
// never publish errors or ownership; LoadProject owns that logical frontier.
type ScheduleProject func(project int)
type LoadProject func(project int) (available bool, err error)
type ContainsTarget func(project int, target Target) bool

type DirectProjectUnavailableError struct {
	Project int
	Target  Target
}

func (err *DirectProjectUnavailableError) Error() string {
	return fmt.Sprintf("direct project %d is unavailable for %q", err.Project, err.Target.Path)
}

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

// Selection is the stable boundary between metadata-only direct selection and
// Program-backed validation/import fallback.
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
		plan:              Plan{Targets: make([]Target, len(plan.Targets))},
		bindings:          make([]Binding, len(plan.Targets)),
		metadataByProject: make(map[int]metadataEvidence),
		loadedProjects:    make(map[int]bool),
		directValidated:   make([]bool, len(plan.Targets)),
	}
	for index, target := range plan.Targets {
		selection.plan.Targets[index] = Target{
			Path:          target.Path,
			CanonicalPath: target.CanonicalPath,
			Projects:      append([]int(nil), target.Projects...),
		}
		selection.bindings[index].Project = NoProject
	}
	if loadMetadata == nil {
		return selection, nil
	}

	scheduled := make(map[int]struct{})
	// Resolve every target's direct tier before any import membership is
	// consulted. Per-target candidates intentionally remain independent.
	for targetIndex, target := range selection.plan.Targets {
		for _, project := range target.Projects {
			evidence, err := selection.metadata(project, loadMetadata)
			if err != nil {
				return nil, err
			}
			if !evidence.available || !evidence.metadata.DirectRoot(target) {
				continue
			}
			selection.bindings[targetIndex] = Binding{Project: project, Tier: TierDirect}
			if schedule != nil {
				if _, exists := scheduled[project]; !exists {
					schedule(project)
					scheduled[project] = struct{}{}
				}
			}
			break
		}
	}
	return selection, nil
}

func (selection *Selection) metadata(project int, load LoadMetadata) (metadataEvidence, error) {
	if evidence, ok := selection.metadataByProject[project]; ok {
		return evidence, nil
	}
	if load == nil {
		evidence := metadataEvidence{}
		selection.metadataByProject[project] = evidence
		return evidence, nil
	}
	metadata, available, err := load(project)
	if err != nil {
		return metadataEvidence{}, err
	}
	evidence := metadataEvidence{metadata: metadata, available: available && metadata != nil}
	selection.metadataByProject[project] = evidence
	return evidence, nil
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

// ValidateDirect preserves the established two-stage direct frontier: all
// selected Programs are observed in request target order before any declared
// root is checked in that same order.
func (selection *Selection) ValidateDirect(load LoadProject, contains ContainsTarget) error {
	if selection == nil {
		return nil
	}
	for targetIndex, binding := range selection.bindings {
		if binding.Tier != TierDirect || selection.directValidated[targetIndex] {
			continue
		}
		available, err := selection.loadProject(binding.Project, load)
		if err != nil {
			return err
		}
		if !available {
			return &DirectProjectUnavailableError{
				Project: binding.Project,
				Target:  selection.plan.Targets[targetIndex],
			}
		}
	}
	for targetIndex, binding := range selection.bindings {
		if binding.Tier != TierDirect || selection.directValidated[targetIndex] {
			continue
		}
		target := selection.plan.Targets[targetIndex]
		if contains == nil || !contains(binding.Project, target) {
			return &DirectRootAbsentError{Project: binding.Project, Target: target}
		}
		selection.directValidated[targetIndex] = true
	}
	return nil
}

// Complete resolves import-only ownership for targets with no direct owner.
// Programs are observed strictly in target order and declaration order.
func (selection *Selection) Complete(
	loadMetadata LoadMetadata,
	loadProject LoadProject,
	contains ContainsTarget,
) ([]Binding, error) {
	if selection == nil {
		return nil, nil
	}
	if err := selection.ValidateDirect(loadProject, contains); err != nil {
		return nil, err
	}
	for targetIndex, target := range selection.plan.Targets {
		if selection.bindings[targetIndex].Tier != TierNone {
			continue
		}
		for _, project := range target.Projects {
			evidence, err := selection.metadata(project, loadMetadata)
			if err != nil {
				return nil, err
			}
			if !evidence.available || !evidence.metadata.Supports(target) {
				continue
			}
			available, err := selection.loadProject(project, loadProject)
			if err != nil {
				return nil, err
			}
			if !available || contains == nil || !contains(project, target) {
				continue
			}
			selection.bindings[targetIndex] = Binding{Project: project, Tier: TierImport}
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
	return selection.Complete(loadMetadata, loadProject, contains)
}

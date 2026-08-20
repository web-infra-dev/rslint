package projectselection_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/web-infra-dev/rslint/internal/program/projectselection"
)

type fakeMetadata struct {
	direct  map[string]bool
	support map[string]bool
}

func (metadata fakeMetadata) DirectRoot(target projectselection.Target) bool {
	return metadata.direct[target.Path]
}

func (metadata fakeMetadata) Supports(target projectselection.Target) bool {
	supported, configured := metadata.support[target.Path]
	return !configured || supported
}

type fakeEvidence struct {
	metadata      map[int]projectselection.Metadata
	metadataErr   map[int]error
	programErr    map[int]error
	contains      map[int]map[string]bool
	metadataTrace []int
	programTrace  []int
	programLoads  map[int]int
}

func newFakeEvidence() *fakeEvidence {
	return &fakeEvidence{
		metadata:     make(map[int]projectselection.Metadata),
		metadataErr:  make(map[int]error),
		programErr:   make(map[int]error),
		contains:     make(map[int]map[string]bool),
		programLoads: make(map[int]int),
	}
}

func (evidence *fakeEvidence) loadMetadata(project int) (projectselection.Metadata, bool, error) {
	evidence.metadataTrace = append(evidence.metadataTrace, project)
	if err := evidence.metadataErr[project]; err != nil {
		return nil, false, err
	}
	metadata, available := evidence.metadata[project]
	return metadata, available, nil
}

func (evidence *fakeEvidence) loadProject(project int) (bool, error) {
	evidence.programTrace = append(evidence.programTrace, project)
	evidence.programLoads[project]++
	if err := evidence.programErr[project]; err != nil {
		return false, err
	}
	return true, nil
}

func (evidence *fakeEvidence) containsTarget(project int, target projectselection.Target) bool {
	return evidence.contains[project][target.Path]
}

func bindingProjects(bindings []projectselection.Binding) []int {
	projects := make([]int, len(bindings))
	for index, binding := range bindings {
		projects[index] = binding.Project
	}
	return projects
}

func bindingTiers(bindings []projectselection.Binding) []projectselection.Tier {
	tiers := make([]projectselection.Tier, len(bindings))
	for index, binding := range bindings {
		tiers[index] = binding.Tier
	}
	return tiers
}

func TestResolveDirectOutranksEarlierImport(t *testing.T) {
	const targetPath = "/repo/target.ts"
	evidence := newFakeEvidence()
	evidence.metadata[0] = fakeMetadata{}
	evidence.metadata[1] = fakeMetadata{direct: map[string]bool{targetPath: true}}
	evidence.contains[0] = map[string]bool{targetPath: true}
	evidence.contains[1] = map[string]bool{targetPath: true}

	bindings, err := projectselection.Resolve(
		projectselection.Plan{
			Targets:  []projectselection.Target{{Path: targetPath}},
			Projects: []int{0, 1},
		},
		evidence.loadMetadata,
		evidence.loadProject,
		evidence.containsTarget,
	)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if got, want := bindingProjects(bindings), []int{1}; !reflect.DeepEqual(got, want) {
		t.Fatalf("projects = %v, want %v", got, want)
	}
	if got, want := bindingTiers(bindings), []projectselection.Tier{projectselection.TierDirect}; !reflect.DeepEqual(got, want) {
		t.Fatalf("tiers = %v, want %v", got, want)
	}
	if got, want := evidence.programTrace, []int{1}; !reflect.DeepEqual(got, want) {
		t.Fatalf("Program loads = %v, want %v", got, want)
	}
}

func TestResolvePreservesDeclarationOrderWithinTier(t *testing.T) {
	const targetPath = "/repo/target.ts"
	tests := []struct {
		name      string
		metadata  map[int]projectselection.Metadata
		contains  map[int]map[string]bool
		wantTier  projectselection.Tier
		wantLoads []int
	}{
		{
			name: "direct",
			metadata: map[int]projectselection.Metadata{
				0: fakeMetadata{direct: map[string]bool{targetPath: true}},
				1: fakeMetadata{direct: map[string]bool{targetPath: true}},
			},
			contains:  map[int]map[string]bool{0: {targetPath: true}, 1: {targetPath: true}},
			wantTier:  projectselection.TierDirect,
			wantLoads: []int{0},
		},
		{
			name: "import",
			metadata: map[int]projectselection.Metadata{
				0: fakeMetadata{},
				1: fakeMetadata{},
			},
			contains:  map[int]map[string]bool{0: {targetPath: true}, 1: {targetPath: true}},
			wantTier:  projectselection.TierImport,
			wantLoads: []int{0},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			evidence := newFakeEvidence()
			evidence.metadata = test.metadata
			evidence.contains = test.contains
			bindings, err := projectselection.Resolve(
				projectselection.Plan{
					Targets:  []projectselection.Target{{Path: targetPath}},
					Projects: []int{0, 1},
				},
				evidence.loadMetadata,
				evidence.loadProject,
				evidence.containsTarget,
			)
			if err != nil {
				t.Fatalf("Resolve: %v", err)
			}
			if got := bindings[0]; got.Project != 0 || got.Tier != test.wantTier {
				t.Fatalf("binding = %+v, want project 0 tier %v", got, test.wantTier)
			}
			if !reflect.DeepEqual(evidence.programTrace, test.wantLoads) {
				t.Fatalf("Program loads = %v, want %v", evidence.programTrace, test.wantLoads)
			}
		})
	}
}

func TestResolveReturnsCompleteMixedBinding(t *testing.T) {
	const (
		directPath   = "/repo/direct.ts"
		importedPath = "/repo/imported.ts"
	)
	evidence := newFakeEvidence()
	evidence.metadata[0] = fakeMetadata{}
	evidence.metadata[1] = fakeMetadata{}
	evidence.metadata[2] = fakeMetadata{direct: map[string]bool{directPath: true}}
	evidence.contains[0] = map[string]bool{importedPath: true}
	evidence.contains[2] = map[string]bool{directPath: true}

	bindings, err := projectselection.Resolve(
		projectselection.Plan{
			Targets: []projectselection.Target{
				{Path: directPath},
				{Path: importedPath},
				{Path: "/repo/gap.ts"},
			},
			Projects: []int{0, 1, 2},
		},
		evidence.loadMetadata,
		evidence.loadProject,
		evidence.containsTarget,
	)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if got, want := bindingProjects(bindings), []int{2, 0, projectselection.NoProject}; !reflect.DeepEqual(got, want) {
		t.Fatalf("projects = %v, want %v", got, want)
	}
	if got, want := bindingTiers(bindings), []projectselection.Tier{
		projectselection.TierDirect,
		projectselection.TierImport,
		projectselection.TierNone,
	}; !reflect.DeepEqual(got, want) {
		t.Fatalf("tiers = %v, want %v", got, want)
	}
}

func TestResolveSkipsUnsupportedImportCandidate(t *testing.T) {
	const targetPath = "/repo/target.js"
	evidence := newFakeEvidence()
	evidence.metadata[0] = fakeMetadata{support: map[string]bool{targetPath: false}}
	evidence.metadata[1] = fakeMetadata{}
	evidence.contains[0] = map[string]bool{targetPath: true}
	evidence.contains[1] = map[string]bool{targetPath: true}

	bindings, err := projectselection.Resolve(
		projectselection.Plan{
			Targets:  []projectselection.Target{{Path: targetPath}},
			Projects: []int{0, 1},
		},
		evidence.loadMetadata,
		evidence.loadProject,
		evidence.containsTarget,
	)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if got := bindings[0]; got.Project != 1 || got.Tier != projectselection.TierImport {
		t.Fatalf("binding = %+v, want project 1 import", got)
	}
	if got, want := evidence.programTrace, []int{1}; !reflect.DeepEqual(got, want) {
		t.Fatalf("Program loads = %v, want %v", got, want)
	}
}

func TestResolvePreservesErrorFrontier(t *testing.T) {
	const targetPath = "/repo/target.ts"
	metadataFailure := errors.New("metadata failure")
	programFailure := errors.New("program failure")
	tests := []struct {
		name         string
		configure    func(*fakeEvidence)
		wantErr      error
		wantMetadata []int
		wantPrograms []int
	}{
		{
			name: "unreached metadata error after direct winner",
			configure: func(evidence *fakeEvidence) {
				evidence.metadata[0] = fakeMetadata{direct: map[string]bool{targetPath: true}}
				evidence.metadataErr[1] = metadataFailure
				evidence.contains[0] = map[string]bool{targetPath: true}
			},
			wantMetadata: []int{0},
			wantPrograms: []int{0},
		},
		{
			name: "reached metadata error",
			configure: func(evidence *fakeEvidence) {
				evidence.metadata[0] = fakeMetadata{}
				evidence.metadataErr[1] = metadataFailure
			},
			wantErr:      metadataFailure,
			wantMetadata: []int{0, 1},
		},
		{
			name: "unreached Program error after import winner",
			configure: func(evidence *fakeEvidence) {
				evidence.metadata[0] = fakeMetadata{}
				evidence.metadata[1] = fakeMetadata{}
				evidence.contains[0] = map[string]bool{targetPath: true}
				evidence.programErr[1] = programFailure
			},
			wantMetadata: []int{0, 1},
			wantPrograms: []int{0},
		},
		{
			name: "reached Program error",
			configure: func(evidence *fakeEvidence) {
				evidence.metadata[0] = fakeMetadata{}
				evidence.metadata[1] = fakeMetadata{}
				evidence.programErr[0] = programFailure
			},
			wantErr:      programFailure,
			wantMetadata: []int{0, 1},
			wantPrograms: []int{0},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			evidence := newFakeEvidence()
			test.configure(evidence)
			_, err := projectselection.Resolve(
				projectselection.Plan{
					Targets:  []projectselection.Target{{Path: targetPath}},
					Projects: []int{0, 1},
				},
				evidence.loadMetadata,
				evidence.loadProject,
				evidence.containsTarget,
			)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("error = %v, want %v", err, test.wantErr)
			}
			if !reflect.DeepEqual(evidence.metadataTrace, test.wantMetadata) {
				t.Fatalf("metadata trace = %v, want %v", evidence.metadataTrace, test.wantMetadata)
			}
			if !reflect.DeepEqual(evidence.programTrace, test.wantPrograms) {
				t.Fatalf("Program trace = %v, want %v", evidence.programTrace, test.wantPrograms)
			}
		})
	}
}

func TestResolveErrorFrontierWaitsForEveryPendingTarget(t *testing.T) {
	const (
		importedPath = "/repo/imported.ts"
		gapPath      = "/repo/gap.ts"
	)
	programFailure := errors.New("program failure")
	tests := []struct {
		name          string
		firstContains map[string]bool
		wantErr       error
		wantPrograms  []int
	}{
		{
			name:          "later error reached while one target remains",
			firstContains: map[string]bool{importedPath: true},
			wantErr:       programFailure,
			wantPrograms:  []int{0, 1},
		},
		{
			name:          "later error hidden after all targets resolve",
			firstContains: map[string]bool{importedPath: true, gapPath: true},
			wantPrograms:  []int{0},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			evidence := newFakeEvidence()
			evidence.metadata[0] = fakeMetadata{}
			evidence.metadata[1] = fakeMetadata{}
			evidence.contains[0] = test.firstContains
			evidence.programErr[1] = programFailure

			_, err := projectselection.Resolve(
				projectselection.Plan{
					Targets: []projectselection.Target{
						{Path: importedPath},
						{Path: gapPath},
					},
					Projects: []int{0, 1},
				},
				evidence.loadMetadata,
				evidence.loadProject,
				evidence.containsTarget,
			)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("error = %v, want %v", err, test.wantErr)
			}
			if !reflect.DeepEqual(evidence.programTrace, test.wantPrograms) {
				t.Fatalf("Program trace = %v, want %v", evidence.programTrace, test.wantPrograms)
			}
		})
	}
}

func TestValidateDirectTargetsLoadsEveryProgramBeforeCheckingRoots(t *testing.T) {
	const (
		firstPath  = "/repo/first.ts"
		secondPath = "/repo/second.ts"
	)
	programFailure := errors.New("second Program failure")
	makeSelection := func(project int, targetPath string) *projectselection.Selection {
		selection, err := projectselection.ResolveDirect(
			projectselection.Plan{
				Targets:  []projectselection.Target{{Path: targetPath}},
				Projects: []int{project},
			},
			func(int) (projectselection.Metadata, bool, error) {
				return fakeMetadata{direct: map[string]bool{targetPath: true}}, true, nil
			},
			nil,
		)
		if err != nil {
			t.Fatalf("ResolveDirect(%q): %v", targetPath, err)
		}
		return selection
	}
	first := makeSelection(0, firstPath)
	second := makeSelection(1, secondPath)
	containsCalls := 0
	err := projectselection.ValidateDirectTargets(
		[]projectselection.DirectTarget{
			{Selection: first, Index: 0},
			{Selection: second, Index: 0},
		},
		func(project int) (bool, error) {
			if project == 1 {
				return false, programFailure
			}
			return true, nil
		},
		func(int, projectselection.Target) bool {
			containsCalls++
			return false
		},
	)
	if !errors.Is(err, programFailure) {
		t.Fatalf("error = %v, want %v", err, programFailure)
	}
	if containsCalls != 0 {
		t.Fatalf("root checks = %d, want none before every direct Program loads", containsCalls)
	}
}

func TestResolveSharedProjectCanServeIndependentOwnerOrders(t *testing.T) {
	const (
		targetA = "/repo/a.ts"
		targetB = "/repo/b.ts"
	)
	evidence := newFakeEvidence()
	evidence.metadata[0] = fakeMetadata{}
	evidence.metadata[1] = fakeMetadata{}
	evidence.contains[0] = map[string]bool{targetA: true, targetB: true}
	evidence.contains[1] = map[string]bool{targetA: true, targetB: true}

	first, err := projectselection.Resolve(
		projectselection.Plan{
			Targets:  []projectselection.Target{{Path: targetA}},
			Projects: []int{0, 1},
		},
		evidence.loadMetadata,
		evidence.loadProject,
		evidence.containsTarget,
	)
	if err != nil {
		t.Fatalf("first Resolve: %v", err)
	}
	second, err := projectselection.Resolve(
		projectselection.Plan{
			Targets:  []projectselection.Target{{Path: targetB}},
			Projects: []int{1, 0},
		},
		evidence.loadMetadata,
		func(project int) (bool, error) {
			if evidence.programLoads[project] > 0 {
				return true, nil
			}
			return evidence.loadProject(project)
		},
		evidence.containsTarget,
	)
	if err != nil {
		t.Fatalf("second Resolve: %v", err)
	}
	if first[0].Project != 0 || second[0].Project != 1 {
		t.Fatalf("owners = %d and %d, want owner-local order 0 and 1", first[0].Project, second[0].Project)
	}
	if evidence.programLoads[0] != 1 || evidence.programLoads[1] != 1 {
		t.Fatalf("Program loads = %v, want each shared project once", evidence.programLoads)
	}
}

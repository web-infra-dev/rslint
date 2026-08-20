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

type fakeProvider struct {
	metadata      map[int]projectselection.Metadata
	metadataErr   map[int]error
	programErr    map[int]error
	contains      map[int]map[string]bool
	metadataTrace []int
	programTrace  []int
	loads         map[int]int
}

func newFakeProvider() *fakeProvider {
	return &fakeProvider{
		metadata:    make(map[int]projectselection.Metadata),
		metadataErr: make(map[int]error),
		programErr:  make(map[int]error),
		contains:    make(map[int]map[string]bool),
		loads:       make(map[int]int),
	}
}

func (provider *fakeProvider) loadMetadata(project int) (projectselection.Metadata, bool, error) {
	provider.metadataTrace = append(provider.metadataTrace, project)
	if err := provider.metadataErr[project]; err != nil {
		return nil, false, err
	}
	metadata, available := provider.metadata[project]
	return metadata, available, nil
}

func (provider *fakeProvider) loadProject(project int) (bool, error) {
	provider.programTrace = append(provider.programTrace, project)
	provider.loads[project]++
	if err := provider.programErr[project]; err != nil {
		return false, err
	}
	return true, nil
}

func (provider *fakeProvider) containsTarget(project int, target projectselection.Target) bool {
	return provider.contains[project][target.Path]
}

func TestResolveUsesTargetSpecificCandidatesAndDirectFirst(t *testing.T) {
	const (
		a = "/repo/a.ts"
		b = "/repo/b.ts"
	)
	provider := newFakeProvider()
	provider.metadata[0] = fakeMetadata{}
	provider.metadata[1] = fakeMetadata{direct: map[string]bool{a: true}}
	provider.metadata[2] = fakeMetadata{direct: map[string]bool{b: true}}
	provider.contains[0] = map[string]bool{a: true, b: true}
	provider.contains[1] = map[string]bool{a: true}
	provider.contains[2] = map[string]bool{a: true, b: true}

	bindings, err := projectselection.Resolve(
		projectselection.Plan{Targets: []projectselection.Target{
			{Path: a, Projects: []int{0, 1}},
			{Path: b, Projects: []int{2}},
		}},
		provider.loadMetadata,
		provider.loadProject,
		provider.containsTarget,
	)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if got, want := bindings, []projectselection.Binding{
		{Project: 1, Tier: projectselection.TierDirect},
		{Project: 2, Tier: projectselection.TierDirect},
	}; !reflect.DeepEqual(got, want) {
		t.Fatalf("bindings = %+v, want %+v", got, want)
	}
	if provider.loads[0] != 0 {
		t.Fatalf("earlier import candidate was built before direct selection")
	}
}

func TestResolvePreservesTargetAndDeclarationErrorFrontier(t *testing.T) {
	const (
		imported = "/repo/imported.ts"
		gap      = "/repo/gap.ts"
	)
	sentinel := errors.New("reached program failure")
	tests := []struct {
		name          string
		firstContains map[string]bool
		wantErr       error
		wantTrace     []int
	}{
		{
			name:          "later error reached by remaining target",
			firstContains: map[string]bool{imported: true},
			wantErr:       sentinel,
			wantTrace:     []int{0, 1},
		},
		{
			name:          "later error hidden after first target resolves both",
			firstContains: map[string]bool{imported: true, gap: true},
			wantTrace:     []int{0},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			provider := newFakeProvider()
			provider.metadata[0] = fakeMetadata{}
			provider.metadata[1] = fakeMetadata{}
			provider.contains[0] = test.firstContains
			provider.programErr[1] = sentinel
			_, err := projectselection.Resolve(
				projectselection.Plan{Targets: []projectselection.Target{
					{Path: imported, Projects: []int{0, 1}},
					{Path: gap, Projects: []int{0, 1}},
				}},
				provider.loadMetadata,
				provider.loadProject,
				provider.containsTarget,
			)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("error = %v, want %v", err, test.wantErr)
			}
			if !reflect.DeepEqual(provider.programTrace, test.wantTrace) {
				t.Fatalf("Program trace = %v, want %v", provider.programTrace, test.wantTrace)
			}
		})
	}
}

func TestResolveSkipsUnsupportedImportCandidate(t *testing.T) {
	const target = "/repo/target.js"
	provider := newFakeProvider()
	provider.metadata[0] = fakeMetadata{support: map[string]bool{target: false}}
	provider.metadata[1] = fakeMetadata{}
	provider.contains[0] = map[string]bool{target: true}
	provider.contains[1] = map[string]bool{target: true}

	bindings, err := projectselection.Resolve(
		projectselection.Plan{Targets: []projectselection.Target{{Path: target, Projects: []int{0, 1}}}},
		provider.loadMetadata,
		provider.loadProject,
		provider.containsTarget,
	)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if got := bindings[0]; got.Project != 1 || got.Tier != projectselection.TierImport {
		t.Fatalf("binding = %+v, want project 1 import", got)
	}
	if got, want := provider.programTrace, []int{1}; !reflect.DeepEqual(got, want) {
		t.Fatalf("Program trace = %v, want %v", got, want)
	}
}

func TestResolveReturnsCompleteMixedBinding(t *testing.T) {
	const (
		direct   = "/repo/direct.ts"
		imported = "/repo/imported.ts"
		gap      = "/repo/gap.ts"
	)
	provider := newFakeProvider()
	provider.metadata[0] = fakeMetadata{}
	provider.metadata[1] = fakeMetadata{direct: map[string]bool{direct: true}}
	provider.contains[0] = map[string]bool{imported: true}
	provider.contains[1] = map[string]bool{direct: true}

	bindings, err := projectselection.Resolve(
		projectselection.Plan{Targets: []projectselection.Target{
			{Path: direct, Projects: []int{0, 1}},
			{Path: imported, Projects: []int{0, 1}},
			{Path: gap, Projects: []int{0, 1}},
		}},
		provider.loadMetadata,
		provider.loadProject,
		provider.containsTarget,
	)
	if err != nil {
		t.Fatal(err)
	}
	want := []projectselection.Binding{
		{Project: 1, Tier: projectselection.TierDirect},
		{Project: 0, Tier: projectselection.TierImport},
		{Project: projectselection.NoProject, Tier: projectselection.TierNone},
	}
	if !reflect.DeepEqual(bindings, want) {
		t.Fatalf("bindings = %+v, want %+v", bindings, want)
	}
	if provider.loads[0] != 1 || provider.loads[1] != 1 {
		t.Fatalf("Program loads = %v, want each selected project once", provider.loads)
	}
}

func TestResolveTreatsMissingMetadataProviderAsUnavailable(t *testing.T) {
	bindings, err := projectselection.Resolve(
		projectselection.Plan{Targets: []projectselection.Target{{Path: "/repo/a.ts", Projects: []int{0}}}},
		nil,
		nil,
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(bindings) != 1 || bindings[0] != (projectselection.Binding{Project: projectselection.NoProject, Tier: projectselection.TierNone}) {
		t.Fatalf("bindings = %+v, want one no-project binding", bindings)
	}
}

func TestResolvePreservesTargetSpecificDeclarationOrder(t *testing.T) {
	const (
		rootTarget  = "/repo/root.ts"
		childTarget = "/repo/child.ts"
	)
	provider := newFakeProvider()
	provider.metadata[0] = fakeMetadata{}
	provider.metadata[1] = fakeMetadata{}
	provider.contains[0] = map[string]bool{rootTarget: true, childTarget: true}
	provider.contains[1] = map[string]bool{rootTarget: true, childTarget: true}

	bindings, err := projectselection.Resolve(
		projectselection.Plan{Targets: []projectselection.Target{
			{Path: rootTarget, Projects: []int{0, 1}},
			{Path: childTarget, Projects: []int{1, 0}},
		}},
		provider.loadMetadata,
		provider.loadProject,
		provider.containsTarget,
	)
	if err != nil {
		t.Fatal(err)
	}
	want := []projectselection.Binding{
		{Project: 0, Tier: projectselection.TierImport},
		{Project: 1, Tier: projectselection.TierImport},
	}
	if !reflect.DeepEqual(bindings, want) {
		t.Fatalf("bindings = %+v, want %+v", bindings, want)
	}
	if provider.loads[0] != 1 || provider.loads[1] != 1 {
		t.Fatalf("shared project loads = %v, want once per project", provider.loads)
	}
}

func TestValidateDirectLoadsAllWinnersBeforeCheckingRoots(t *testing.T) {
	const (
		first  = "/repo/first.ts"
		second = "/repo/second.ts"
	)
	sentinel := errors.New("second direct Program failed")
	provider := newFakeProvider()
	provider.metadata[0] = fakeMetadata{direct: map[string]bool{first: true}}
	provider.metadata[1] = fakeMetadata{direct: map[string]bool{second: true}}
	provider.contains[0] = map[string]bool{first: false}
	provider.contains[1] = map[string]bool{second: true}
	provider.programErr[1] = sentinel

	_, err := projectselection.Resolve(
		projectselection.Plan{Targets: []projectselection.Target{
			{Path: first, Projects: []int{0}},
			{Path: second, Projects: []int{1}},
		}},
		provider.loadMetadata,
		provider.loadProject,
		provider.containsTarget,
	)
	if !errors.Is(err, sentinel) {
		t.Fatalf("error = %v, want later Program error before earlier root absence", err)
	}
	if got, want := provider.programTrace, []int{0, 1}; !reflect.DeepEqual(got, want) {
		t.Fatalf("Program trace = %v, want %v", got, want)
	}
}

func TestResolveMetadataErrorFrontier(t *testing.T) {
	const target = "/repo/target.ts"
	sentinel := errors.New("metadata failure")

	t.Run("reached earlier error is visible", func(t *testing.T) {
		provider := newFakeProvider()
		provider.metadataErr[0] = sentinel
		provider.metadata[1] = fakeMetadata{direct: map[string]bool{target: true}}
		_, err := projectselection.Resolve(
			projectselection.Plan{Targets: []projectselection.Target{{Path: target, Projects: []int{0, 1}}}},
			provider.loadMetadata,
			provider.loadProject,
			provider.containsTarget,
		)
		if !errors.Is(err, sentinel) {
			t.Fatalf("error = %v, want %v", err, sentinel)
		}
	})

	t.Run("later error after direct winner is hidden", func(t *testing.T) {
		provider := newFakeProvider()
		provider.metadata[0] = fakeMetadata{direct: map[string]bool{target: true}}
		provider.metadataErr[1] = sentinel
		provider.contains[0] = map[string]bool{target: true}
		bindings, err := projectselection.Resolve(
			projectselection.Plan{Targets: []projectselection.Target{{Path: target, Projects: []int{0, 1}}}},
			provider.loadMetadata,
			provider.loadProject,
			provider.containsTarget,
		)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := provider.metadataTrace, []int{0}; !reflect.DeepEqual(got, want) {
			t.Fatalf("metadata trace = %v, want %v", got, want)
		}
		if bindings[0].Project != 0 || bindings[0].Tier != projectselection.TierDirect {
			t.Fatalf("binding = %+v", bindings[0])
		}
	})
}

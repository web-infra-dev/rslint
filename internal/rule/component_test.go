package rule

import (
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// countingFS records how many times a path's bytes were read, which is what
// "costs nothing" has to mean for a store that exists to avoid reads.
type countingFS struct {
	vfs.FS
	reads int
}

func (f *countingFS) ReadFile(path string) (string, bool) {
	f.reads++
	return f.FS.ReadFile(path)
}

func newComponentFS(path, contents string) *countingFS {
	return &countingFS{FS: utils.NewOverlayVFSForFile(path, contents)}
}

const componentFixture = "<template>\n  <div/>\n</template>\n\n" +
	"<script>\nexport const shared = 1;\n</script>\n\n" +
	"<script setup>\nconst local = 2;\n</script>\n"

// TestComponentReadsNothingUnlessAsked locks in the performance claim the store
// rests on: a component's bytes are read only when a rule asks about the
// component, and then only once for every rule on that file.
func TestComponentReadsNothingUnlessAsked(t *testing.T) {
	t.Parallel()

	t.Run("a store nobody asks reads nothing", func(t *testing.T) {
		t.Parallel()
		fileSystem := newComponentFS("/App.vue", componentFixture)
		NewComponent(fileSystem, "/App.vue")

		if fileSystem.reads != 0 {
			t.Errorf("read the component %d times without being asked", fileSystem.reads)
		}
	})

	t.Run("many questions read once", func(t *testing.T) {
		t.Parallel()
		fileSystem := newComponentFS("/App.vue", componentFixture)
		component := NewComponent(fileSystem, "/App.vue")

		component.IsComponent()
		component.ScriptSetupRange()
		component.ScriptSetupRange()

		if fileSystem.reads != 1 {
			t.Errorf("read the component %d times, want 1", fileSystem.reads)
		}
	})

	t.Run("a file that is not a component reads nothing", func(t *testing.T) {
		t.Parallel()
		fileSystem := newComponentFS("/main.ts", "export const value = 1;\n")
		component := NewComponent(fileSystem, "/main.ts")

		if component.IsComponent() {
			t.Error("a .ts file reported itself a component")
		}
		if fileSystem.reads != 0 {
			t.Errorf("read a non-component %d times", fileSystem.reads)
		}
	})
}

func TestComponentScriptSetupRange(t *testing.T) {
	t.Parallel()

	component := NewComponent(newComponentFS("/App.vue", componentFixture), "/App.vue")
	setup, ok := component.ScriptSetupRange()
	if !ok {
		t.Fatal("the setup block was not found")
	}
	if got := componentFixture[setup.Pos():setup.End()]; got != "\nconst local = 2;\n" {
		t.Errorf("setup range covers %q", got)
	}

	withoutSetup := NewComponent(newComponentFS("/App.vue", "<script>\nconst a = 1;\n</script>\n"), "/App.vue")
	if _, ok := withoutSetup.ScriptSetupRange(); ok {
		t.Error("a component with only a plain script reported a setup block")
	}
}

func TestComponentInScriptSetup(t *testing.T) {
	t.Parallel()

	component := NewComponent(newComponentFS("/App.vue", componentFixture), "/App.vue")

	// The plain script's export sits outside the setup block, which is the
	// distinction no-export-in-script-setup is built on.
	if component.InScriptSetup(strings.Index(componentFixture, "export const shared")) {
		t.Error("an export in the plain block was placed inside the setup block")
	}
	if !component.InScriptSetup(strings.Index(componentFixture, "const local")) {
		t.Error("the setup block's own code was placed outside it")
	}

	withoutSetup := NewComponent(newComponentFS("/App.vue", "<script>\nconst a = 1;\n</script>\n"), "/App.vue")
	if withoutSetup.InScriptSetup(10) {
		t.Error("a component with no setup block placed a position inside one")
	}
}

// TestComponentNilStore covers the shape a manually assembled rule context hands
// a rule, which must answer rather than panic.
func TestComponentNilStore(t *testing.T) {
	t.Parallel()

	var component *Component
	if component.IsComponent() {
		t.Error("a nil store reported a component")
	}
	if _, ok := component.ScriptSetupRange(); ok {
		t.Error("a nil store returned a setup range")
	}
	if component.InScriptSetup(0) {
		t.Error("a nil store placed a position inside a setup block")
	}
	if (&RuleContext{}).IsExposedToTemplate(nil, "x") {
		t.Error("an empty context reported a binding exposed to a template")
	}
}

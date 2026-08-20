package loader

import (
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/tspath"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
)

// ProjectSet is the stable, deduplicated set of configured project generations
// built for one load pass. Its compiler backing and config associations are
// private assembly details; consumers use Programs after loading or binding.
type ProjectSet struct {
	compilerPrograms  []*compiler.Program
	programs          []*lintprogram.Program
	typeCheckPrograms []*lintprogram.Program
	targetBinding     *projectTargetBinding
}

// Programs returns the configured rslint Programs in stable declaration order.
// The slice is read-only and remains owned by the ProjectSet.
func (projects ProjectSet) Programs() []*lintprogram.Program {
	return projects.programs
}

func (projects ProjectSet) Len() int {
	return len(projects.programs)
}

// TypeCheckPrograms returns only the configured projects admitted by the
// program-wide type-check catalog. Lint-only effective candidates are excluded.
func (projects ProjectSet) TypeCheckPrograms() []*lintprogram.Program {
	return projects.typeCheckPrograms
}

type projectSpec struct {
	tsconfigPath string
	programCwd   string
}

type projectPlan struct {
	specs []projectSpec
}

func exactPathID(filePath string) string {
	return string(tspath.ToPath(tspath.NormalizePath(filePath), "", true))
}

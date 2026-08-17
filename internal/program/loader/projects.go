package loader

import (
	"fmt"
	"runtime"
	"sort"
	"sync"

	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
)

type configOrders map[string]int

// ProjectSet is the stable, deduplicated set of configured project generations
// built for one load pass. Its compiler backing and config associations are
// private assembly details; consumers use Programs after loading or binding.
type ProjectSet struct {
	compilerPrograms []*compiler.Program
	programs         []*lintprogram.Program
	configOrders     []configOrders
}

// Programs returns the configured rslint Programs in stable declaration order.
// The slice is read-only and remains owned by the ProjectSet.
func (projects ProjectSet) Programs() []*lintprogram.Program {
	return projects.programs
}

func (projects ProjectSet) Len() int {
	return len(projects.programs)
}

type projectSpec struct {
	tsconfigPath string
	programCwd   string
	configOrders configOrders
}

type projectPlan struct {
	specs       []projectSpec
	terminalErr error
}

func exactPathID(filePath string) string {
	return string(tspath.ToPath(tspath.NormalizePath(filePath), "", true))
}

func buildProjectPlan(configMap map[string]rslintconfig.RslintConfig, fsys vfs.FS) projectPlan {
	if len(configMap) == 0 {
		return projectPlan{}
	}

	configDirs := make([]string, 0, len(configMap))
	for configDir := range configMap {
		configDirs = append(configDirs, configDir)
	}
	sort.Strings(configDirs)

	plan := projectPlan{}
	programByTsconfig := make(map[string]int)
	for _, configDir := range configDirs {
		entries := configMap[configDir]
		normalizedConfigDir := tspath.NormalizePath(configDir)
		configDirID := exactPathID(normalizedConfigDir)
		tsconfigs, err := rslintconfig.ResolveTsConfigPaths(entries, normalizedConfigDir, fsys)
		if err != nil {
			plan.terminalErr = fmt.Errorf("resolve tsconfigs for %q: %w", configDir, err)
			return plan
		}

		for order, tsconfigPath := range tsconfigs {
			tsconfigPath = tspath.NormalizePath(tsconfigPath)
			tsconfigID := exactPathID(tsconfigPath)
			if programIndex, ok := programByTsconfig[tsconfigID]; ok {
				if _, associated := plan.specs[programIndex].configOrders[configDirID]; !associated {
					plan.specs[programIndex].configOrders[configDirID] = order
				}
				continue
			}

			programByTsconfig[tsconfigID] = len(plan.specs)
			plan.specs = append(plan.specs, projectSpec{
				tsconfigPath: tsconfigPath,
				programCwd:   tspath.GetDirectoryPath(tsconfigPath),
				configOrders: configOrders{configDirID: order},
			})
		}
	}
	return plan
}

func (s *Session) executeProjectPlan(plan projectPlan, singleThreaded bool) (ProjectSet, error) {
	if err := s.validate(); err != nil {
		return ProjectSet{}, err
	}
	if len(plan.specs) == 0 {
		if plan.terminalErr != nil {
			return ProjectSet{}, plan.terminalErr
		}
		return ProjectSet{}, nil
	}
	compilerPrograms := make([]*compiler.Program, len(plan.specs))
	errs := make([]error, len(plan.specs))
	workerCount := min(runtime.GOMAXPROCS(0), len(plan.specs))
	parallel := !singleThreaded && workerCount > 1
	if parallel {
		s.context.enableConcurrentProgramQueries()
	}
	build := func(index int) {
		spec := plan.specs[index]
		compilerPrograms[index], errs[index] = s.context.createProjectProgram(
			singleThreaded,
			spec.programCwd,
			spec.tsconfigPath,
		)
	}

	if !parallel {
		for index := range plan.specs {
			build(index)
			if errs[index] != nil {
				break
			}
		}
	} else {
		jobs := make(chan int, workerCount)
		var workers sync.WaitGroup
		workers.Add(workerCount)
		for range workerCount {
			go func() {
				defer workers.Done()
				for index := range jobs {
					build(index)
				}
			}()
		}
		for index := range plan.specs {
			jobs <- index
		}
		close(jobs)
		workers.Wait()
	}

	for index, err := range errs {
		if err != nil {
			return ProjectSet{}, fmt.Errorf(
				"create TypeScript Program from %q: %w",
				plan.specs[index].tsconfigPath,
				err,
			)
		}
	}
	if plan.terminalErr != nil {
		return ProjectSet{}, plan.terminalErr
	}

	orders := make([]configOrders, len(plan.specs))
	for index := range plan.specs {
		orders[index] = plan.specs[index].configOrders
	}
	return ProjectSet{
		compilerPrograms: compilerPrograms,
		programs:         lintprogram.NewFromCompilers(compilerPrograms),
		configOrders:     orders,
	}, nil
}

// BuildProjects constructs every unique tsconfig declared by configs in
// stable config/project order.
func (s *Session) BuildProjects(
	configs map[string]rslintconfig.RslintConfig,
	singleThreaded bool,
) (ProjectSet, error) {
	if err := s.validate(); err != nil {
		return ProjectSet{}, err
	}
	plan := buildProjectPlan(configs, s.FS())
	return s.executeProjectPlan(plan, singleThreaded)
}

func (s *Session) BuildProject(
	configDirectory string,
	config rslintconfig.RslintConfig,
	singleThreaded bool,
) (ProjectSet, error) {
	return s.BuildProjects(
		map[string]rslintconfig.RslintConfig{configDirectory: config},
		singleThreaded,
	)
}

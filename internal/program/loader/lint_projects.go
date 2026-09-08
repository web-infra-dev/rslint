package loader

import (
	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/tsoptions"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/config/target"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/program/projectservice"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// appendServiceProjects retains the complete Programs already used to prove
// service membership. They never enter ordinary owner candidate lists.
func (s *Session) appendServiceProjects(
	set *ProjectSet,
	plan target.Plan,
	policies map[target.File]rslintconfig.ProjectPolicy,
	singleThreaded bool,
) error {
	if len(policies) == 0 {
		return nil
	}
	selector := projectservice.New(projectservice.Host{
		FS: s.FS(),
		ParseConfig: func(path string) (*tsoptions.ParsedCommandLine, error) {
			_, parsed, err := s.context.parseConfig(tspath.GetDirectoryPath(path), path)
			return parsed, err
		},
		CreateProgram: func(path string, parsed *tsoptions.ParsedCommandLine) (*compiler.Program, error) {
			return utils.CreateProgramFromParsedConfigLenientWithProjectReferences(
				singleThreaded, parsed, s.context.newCompilerHostWithCache(tspath.GetDirectoryPath(path)),
			)
		},
	})
	candidatesByProgram := make(map[*compiler.Program][]int)
	for _, file := range plan.Files {
		policy := policies[file]
		if !policy.ProjectService {
			continue
		}
		root := policy.TsconfigRootDir
		if root == "" {
			root = file.ConfigDirectory
		}
		selected, err := selector.Select(file.Path, root)
		if err != nil {
			return err
		}
		if set.targetProjects == nil {
			set.targetProjects = make(map[target.File][]int)
		}
		set.targetProjects[file] = nil
		if selected.Program == nil {
			continue
		}
		candidates, exists := candidatesByProgram[selected.Program]
		if !exists {
			candidates = []int{len(set.compilerPrograms)}
			candidatesByProgram[selected.Program] = candidates
			set.compilerPrograms = append(set.compilerPrograms, selected.Program)
			set.programs = append(set.programs, lintprogram.NewFromCompiler(selected.Program))
			set.configOrders = append(set.configOrders, nil)
		}
		set.targetProjects[file] = candidates
	}
	return nil
}

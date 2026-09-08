package loader

import (
	"github.com/microsoft/TypeScript/tsc/shim/tsoptions"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/config/target"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// discoverServiceProjects adds only directly selected configs to the existing
// construction plan. Unmatched targets get an empty candidate list and remain
// gaps even if some other target's Program imports their source files.
func (s *Session) discoverServiceProjects(
	plan *projectPlan,
	targets target.Plan,
	policies map[target.File]rslintconfig.ProjectPolicy,
) error {
	if len(policies) == 0 {
		return nil
	}
	var discovery *utils.TypeScriptProjectDiscovery
	var projects map[string][]int
	for _, file := range targets.Files {
		root := policies[file].ServiceRootDirectory
		if root == "" {
			continue
		}
		if discovery == nil {
			discovery = utils.NewTypeScriptProjectDiscovery(s.FS(), func(path string) (*tsoptions.ParsedCommandLine, error) {
				_, parsed, err := s.context.parseConfig(tspath.GetDirectoryPath(path), path)
				return parsed, err
			})
			projects = make(map[string][]int)
		}
		parsed, err := discovery.Find(file.Path, root)
		if err != nil {
			return err
		}
		if plan.targetProjects == nil {
			plan.targetProjects = make(map[target.File][]int)
		}
		plan.targetProjects[file] = nil
		if parsed == nil {
			continue
		}
		path := tspath.NormalizePath(parsed.ConfigName())
		indexes, exists := projects[path]
		if !exists {
			indexes = []int{len(plan.specs)}
			projects[path] = indexes
			plan.specs = append(plan.specs, projectSpec{
				tsconfigPath: path, programCwd: tspath.GetDirectoryPath(path),
				parsed: parsed, sourceReferences: true,
			})
		}
		plan.targetProjects[file] = indexes
	}
	return nil
}

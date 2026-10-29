package contentmappers

import (
	"sync"

	"github.com/microsoft/TypeScript/tsc/shim/contentmapper"
	"github.com/microsoft/TypeScript/tsc/shim/tsoptions"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
)

// ProjectCache keeps at most one mapper project per tsconfig for a long-lived
// owner such as the language server, so rebuilding a Program does not
// transform its mapped files again. A project is reused while its tsconfig
// keeps the same parse, and replaced when the tsconfig is parsed again.
type ProjectCache struct {
	runtime *Runtime

	mu       sync.Mutex
	byConfig map[tspath.Path]cachedProject
}

type cachedProject struct {
	config  *tsoptions.ParsedCommandLine
	project contentmapper.Project
	release func() error
}

// NewProjectCache returns a cache whose projects come from r. Closing r
// releases them.
func (r *Runtime) NewProjectCache() *ProjectCache {
	return &ProjectCache{runtime: r}
}

// Project returns the mapper project for config, or nil when config uses none.
func (c *ProjectCache) Project(config *tsoptions.ParsedCommandLine) contentmapper.Project {
	if c == nil || config == nil || config.ConfigName() == "" {
		return nil
	}
	key := tspath.ToPath(config.ConfigName(), "", true)
	c.mu.Lock()
	defer c.mu.Unlock()
	cached, ok := c.byConfig[key]
	if ok && cached.config == config {
		return cached.project
	}
	if ok {
		_ = cached.release()
	}
	project, release := c.runtime.Open(config)
	if c.byConfig == nil {
		c.byConfig = make(map[tspath.Path]cachedProject)
	}
	c.byConfig[key] = cachedProject{config: config, project: project, release: release}
	return project
}

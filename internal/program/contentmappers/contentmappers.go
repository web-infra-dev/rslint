// Package contentmappers runs the content mappers that tsconfig files declare
// in `contentMappers`, so mapped files (for example .vue) join a Program and
// provide their types to the files that import them.
//
// A mapper is an npm package that ts-go runs as a child process, so it only
// runs when the caller opts in, like tsc's --runExternalCode. Each entry point
// owns one Runtime: the CLI per run, the API per request, and the language
// server per session.
package contentmappers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"slices"
	"sync"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/contentmapper"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/diagnostics"
	"github.com/microsoft/TypeScript/tsc/shim/locale"
	"github.com/microsoft/TypeScript/tsc/shim/tsoptions"
)

// Options configures a Runtime.
type Options struct {
	// Enabled lets declared mappers run. Without it ts-go drops them, and the
	// files they would map are left out of the Program.
	Enabled bool
	// Spawner starts mapper processes. Nil starts real child processes.
	Spawner contentmapper.Spawner
}

// Runtime decides whether declared content mappers run and owns the mapper
// processes and projects they need. A nil Runtime runs no mappers and records
// nothing. All methods are safe for concurrent use.
type Runtime struct {
	enabled bool
	spawner contentmapper.Spawner

	mu      sync.Mutex
	host    contentmapper.Host
	leases  map[*lease]struct{}
	ignored map[string]struct{}
	closed  bool
}

type lease struct {
	project contentmapper.Project
	once    sync.Once
	err     error
}

func (l *lease) close() error {
	l.once.Do(func() { l.err = l.project.Close() })
	return l.err
}

// New returns a Runtime with the given options.
func New(options Options) *Runtime {
	spawner := options.Spawner
	if spawner == nil {
		spawner = contentmapper.SpawnerFunc(spawnProcess)
	}
	return &Runtime{enabled: options.Enabled, spawner: spawner}
}

// Enabled reports whether declared mappers run.
func (r *Runtime) Enabled() bool {
	return r != nil && r.enabled
}

// ParseOptions returns the existing options to parse a tsconfig with. ts-go
// keeps a config's mappers only when these carry RunExternalCode.
func (r *Runtime) ParseOptions() *core.CompilerOptions {
	options := &core.CompilerOptions{}
	if r.Enabled() {
		options.RunExternalCode = core.TSTrue
	}
	return options
}

// Observe records a parsed config whose mappers were dropped because the
// Runtime is not enabled, and reports whether this is the first time the
// config is recorded. Call it after every tsconfig parse.
func (r *Runtime) Observe(config *tsoptions.ParsedCommandLine) bool {
	if r == nil || config == nil || config.ConfigName() == "" {
		return false
	}
	if !slices.ContainsFunc(config.GetConfigFileParsingDiagnostics(), func(diagnostic *ast.Diagnostic) bool {
		return diagnostic.Code() == diagnostics.Content_mappers_require_the_runExternalCode_command_line_flag_to_be_enabled.Code()
	}) {
		return false
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.ignored[config.ConfigName()]; ok {
		return false
	}
	if r.ignored == nil {
		r.ignored = make(map[string]struct{})
	}
	r.ignored[config.ConfigName()] = struct{}{}
	return true
}

// IgnoredConfigs returns, sorted, the tsconfig files that declare mappers the
// Runtime did not run.
func (r *Runtime) IgnoredConfigs() []string {
	if r == nil {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	configs := make([]string, 0, len(r.ignored))
	for config := range r.ignored {
		configs = append(configs, config)
	}
	slices.Sort(configs)
	return configs
}

// Open returns the mapper project for config, or nil when the Runtime is not
// enabled, is closed, or config declares no mappers. The project stays open
// until release is called or the Runtime closes. release is never nil and may
// be called more than once.
func (r *Runtime) Open(config *tsoptions.ParsedCommandLine) (project contentmapper.Project, release func() error) {
	if !r.Enabled() || config == nil || len(config.ContentMappers()) == 0 {
		return nil, func() error { return nil }
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return nil, func() error { return nil }
	}
	if r.host == nil {
		r.host = contentmapper.NewHost(context.Background(), r.spawner, locale.Default)
	}
	l := &lease{project: r.host.Project(contentmapper.ProjectSpec{
		ConfigFileName:  config.ConfigName(),
		Mappers:         config.ContentMappers(),
		CompilerOptions: config.CompilerOptions(),
	})}
	if r.leases == nil {
		r.leases = make(map[*lease]struct{})
	}
	r.leases[l] = struct{}{}
	return l.project, func() error {
		r.mu.Lock()
		delete(r.leases, l)
		r.mu.Unlock()
		return l.close()
	}
}

// Close releases every open project and stops every mapper process. Open
// returns nil afterwards.
func (r *Runtime) Close() error {
	if r == nil {
		return nil
	}
	r.mu.Lock()
	r.closed = true
	leases := r.leases
	r.leases = nil
	host := r.host
	r.mu.Unlock()
	var errs []error
	for l := range leases {
		errs = append(errs, l.close())
	}
	if host != nil {
		errs = append(errs, host.Close())
	}
	return errors.Join(errs...)
}

// WriteIgnoredWarnings writes one warning per config in IgnoredConfigs. optIn
// names the switch that enables mappers for the caller, such as a CLI flag.
func (r *Runtime) WriteIgnoredWarnings(w io.Writer, optIn string) {
	for _, config := range r.IgnoredConfigs() {
		fmt.Fprintf(w, "Warning: %s declares contentMappers, which run only with %s; the files they map provide no types.\n", config, optIn)
	}
}

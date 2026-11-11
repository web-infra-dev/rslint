package moduleresolver

import (
	"encoding/json"
	"maps"
	"slices"

	"github.com/web-infra-dev/rslint/internal/program"
)

// PreparedOptions is an immutable snapshot of runtime resolution options. It
// owns its lists and maps and can be shared across requests and goroutines.
// It retains no Program or filesystem state.
type PreparedOptions struct {
	options Options
	key     string
}

// PrepareOptions copies options and encodes their cache identity once. Nil
// and empty lists retain distinct identities, including when defaults apply.
// Later mutations to the caller's options do not change this snapshot.
func PrepareOptions(options Options) *PreparedOptions {
	options.Extensions = slices.Clone(options.Extensions)
	options.Modules = slices.Clone(options.Modules)
	options.Paths = slices.Clone(options.Paths)
	options.Conditions = slices.Clone(options.Conditions)
	options.MainFiles = slices.Clone(options.MainFiles)
	options.ExtensionAliases = maps.Clone(options.ExtensionAliases)
	for name, extensions := range options.ExtensionAliases {
		options.ExtensionAliases[name] = slices.Clone(extensions)
	}
	options.Aliases = cloneAliases(options.Aliases)
	options.Fallbacks = cloneAliases(options.Fallbacks)
	options.MainFields = slices.Clone(options.MainFields)
	for i := range options.MainFields {
		options.MainFields[i].Name = slices.Clone(options.MainFields[i].Name)
	}
	options.AliasFields = slices.Clone(options.AliasFields)
	for i := range options.AliasFields {
		options.AliasFields[i] = slices.Clone(options.AliasFields[i])
	}
	// Options contains only JSON-compatible value types.
	encoded, err := json.Marshal(options)
	if err != nil {
		return nil
	}
	return &PreparedOptions{options: defaultOptions(options), key: string(encoded)}
}

func cloneAliases(aliases []Alias) []Alias {
	aliases = slices.Clone(aliases)
	for i := range aliases {
		aliases[i].Targets = slices.Clone(aliases[i].Targets)
	}
	return aliases
}

func defaultOptions(options Options) Options {
	if options.Extensions == nil {
		options.Extensions = DefaultExtensions()
	}
	if options.Modules == nil {
		options.Modules = []string{"node_modules"}
	}
	if options.Conditions == nil {
		options.Conditions = []string{"node", "require", "import"}
	}
	return options
}

// PreserveSymlinks reports the snapshot's importer and target path policy.
func (options *PreparedOptions) PreserveSymlinks() bool {
	return options != nil && options.options.PreserveSymlinks
}

// Resolve uses the same generation cache and resolution semantics as Resolve,
// without encoding options again. A nil snapshot returns an empty result.
func (options *PreparedOptions) Resolve(p *program.Program, name, containingFile string) Result {
	if options == nil || p.FS() == nil {
		return Result{}
	}
	return resolveWithOptions(p, name, containingFile, options.options, options.key)
}

package moduleresolver

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/module"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/modules"
	"github.com/web-infra-dev/rslint/internal/utils/pathutil"
)

// PathAliasResolver probes compiler paths independently of package lookup.
// Its lazy TypeScript resolver belongs to one rule invocation.
type PathAliasResolver struct {
	program  *program.Program
	patterns *module.ParsedPatterns
	resolver *module.DefaultResolver
}

func NewPathAliasResolver(p *program.Program) *PathAliasResolver {
	return &PathAliasResolver{program: p}
}

// HasAliasAt reports whether the selected paths pattern resolves a candidate
// name in the original reference's context, including resolution-mode attributes.
// A missing substitution does not suppress ordinary package lookup.
func (r *PathAliasResolver) HasAliasAt(sourceFile *ast.SourceFile, name string, source modules.Source) bool {
	p := r.program
	if !p.OwnsSourceFile(sourceFile) || source.Specifier() == nil || !ast.IsStringLiteralLike(source.Specifier()) {
		return false
	}
	options := p.Options()
	if options.Paths == nil || options.Paths.Size() == 0 {
		return false
	}
	if r.patterns == nil {
		r.patterns = module.TryParsePatterns(options.Paths)
	}
	pattern := module.MatchPatternOrExact(r.patterns, name)
	if !pattern.IsValid() {
		return false
	}
	if r.resolver == nil {
		// Mapped targets must not reapply paths or rootDirs. Only a successful
		// substitution from the selected pattern establishes an alias.
		aliasOptions := options.Clone()
		aliasOptions.Paths = nil
		aliasOptions.RootDirs = nil
		host := utils.CreateCompilerHost(p.CurrentDirectory(), p.FS())
		r.resolver = module.NewResolver(module.ResolverOptions{Host: host, CompilerOptions: aliasOptions})
	}
	fileName := sourceFile.FileName()
	if options.PreserveSymlinks != core.TSTrue {
		if realPath := p.FS().Realpath(fileName.AsPath()); realPath != "" {
			fileName = tspath.RootedFilePathFromPath(realPath)
		}
	}
	mode := p.GetModeForUsageLocation(sourceFile, source.Specifier())
	base := options.GetPathsBasePath(pathutil.Directory(p.CurrentDirectory()))
	for _, target := range options.Paths.GetOrZero(pattern.Text) {
		candidate := tspath.ResolvePath(base.AsString(), strings.Replace(target, "*", pattern.MatchedText(name), 1))
		if resolved, _, _ := r.resolver.ResolveModuleName(candidate, fileName, mode, nil); resolved.IsResolved() {
			return true
		}
	}
	return false
}

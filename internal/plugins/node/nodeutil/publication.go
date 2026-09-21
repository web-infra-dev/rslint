package nodeutil

import (
	"path/filepath"
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils/packagejson"
)

// PublicationCheck shares publication policy between import and require rules.
// Callers choose module resolution and the nodes on which to report.
type PublicationCheck struct {
	program   *program.Program
	pkg       *packagejson.Package
	converter pathConverter
	cwd       string
	allowed   []string
}

// NewPublicationCheck returns nil when the source is unpublished, belongs to an
// ignored private package, or has no usable package metadata.
func NewPublicationCheck(ctx rule.RuleContext, options map[string]any) *PublicationCheck {
	p, fileName := ctx.Program(), ctx.SourceFile.FileName()
	if p == nil || fileName == "<input>" {
		return nil
	}
	pkg := packagejson.FindNearestValid(p, fileName)
	if pkg == nil || options["ignorePrivate"] != false && pkg.Field("private") == true {
		return nil
	}
	cwd := ctx.ProcessCurrentDirectory()
	if cwd == "" {
		cwd = p.CurrentDirectory()
	}
	check := &PublicationCheck{program: p, pkg: pkg, converter: compilePathConverter(options, ctx.Settings), cwd: cwd}
	converted, ok := check.relativePath(fileName)
	if !ok || IsUnpublished(p, pkg, tspath.ResolvePath(pkg.Directory(), converted)) {
		return nil
	}
	check.allowed = StringListSetting("allowModules", options, ctx.Settings)
	return check
}

func (check *PublicationCheck) relativePath(target string) (string, bool) {
	// Like Node's path.relative, unresolved targets and URL spellings use the
	// process directory. Normalize only at the filesystem boundary.
	if !IsAbsolutePath(target) {
		target = filepath.Join(filepath.FromSlash(check.cwd), target)
	}
	relative := tspath.GetRelativePathFromDirectory(check.pkg.Directory(), tspath.NormalizePath(target), tspath.ComparePathsOptions{UseCaseSensitiveFileNames: check.program.FS().UseCaseSensitiveFileNames()})
	return check.converter.convert(relative)
}

// IsUnpublishedFile checks a resolved or lexical target in the source package.
func (check *PublicationCheck) IsUnpublishedFile(target string) bool {
	relative, ok := check.relativePath(target)
	return ok && relative != "" && IsUnpublished(check.program, check.pkg, tspath.ResolvePath(check.pkg.Directory(), relative))
}

// IsUnpublishedDependency checks this package's own development dependencies.
// Workspace ancestors do not supply production dependencies for publication.
func (check *PublicationCheck) IsUnpublishedDependency(name string) bool {
	development, _ := check.pkg.Field("devDependencies").(map[string]any)
	if _, exists := development[name]; !exists || slices.Contains(check.allowed, name) {
		return false
	}
	for _, field := range []string{"dependencies", "peerDependencies", "optionalDependencies"} {
		dependencies, _ := check.pkg.Field(field).(map[string]any)
		if _, exists := dependencies[name]; exists {
			return false
		}
	}
	return true
}

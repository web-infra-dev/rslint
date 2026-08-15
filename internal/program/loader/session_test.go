package loader

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/vfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestSessionRejectsNilFilesystemBeforeLoading(t *testing.T) {
	var typedNil *programMetadataFS
	tests := []struct {
		name string
		fs   vfs.FS
	}{
		{name: "nil"},
		{name: "typed nil", fs: typedNil},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			session := NewSession(test.fs)
			if session.FS() != nil {
				t.Fatal("invalid session unexpectedly exposed a filesystem")
			}
			if _, err := session.BuildProjects(nil, true); err == nil {
				t.Fatal("BuildProjects accepted an invalid session")
			}
			if _, err := session.LoadCLI(ProjectSet{}, rslintconfig.LintTargetPlan{}, "/repo", true); err == nil {
				t.Fatal("LoadCLI accepted an invalid session")
			}
			if _, err := session.LoadAPI(ProjectSet{}, rslintconfig.LintTargetPlan{}, "/repo", true); err == nil {
				t.Fatal("LoadAPI accepted an invalid session")
			}
		})
	}
}

func TestSessionRecordsEmptyInitialCompilerGeneration(t *testing.T) {
	session := &Session{context: &buildContext{parseCache: utils.NewParseCache()}}
	session.retainCompilerPrograms(nil)
	if !session.initialProgramsSet {
		t.Fatal("empty initial compiler generation was not recorded")
	}
}

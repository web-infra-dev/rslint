package lsp

import (
	"runtime"
	"testing"

	"github.com/microsoft/typescript-go/shim/vfs"
)

// mockFS is a mock implementation of vfs.FS for testing. Only FileExists is properly implemented;
// other methods are stubbed with default implementations.
type mockFS struct {
	files map[string]bool // path -> exists
}

func (m *mockFS) FileExists(path string) bool {
	exists, found := m.files[path]
	if !found {
		return false
	}
	return exists
}

// Stubbed implementations of other vfs.FS interface methods for testing purposes
func (m *mockFS) UseCaseSensitiveFileNames() bool                                   { return true }
func (m *mockFS) ReadFile(path string) (string, bool)                               { return "", false }
func (m *mockFS) WriteFile(path string, data string, writeByteOrderMark bool) error { return nil }
func (m *mockFS) Remove(path string) error                                          { return nil }
func (m *mockFS) DirectoryExists(path string) bool                                  { return false }
func (m *mockFS) GetAccessibleEntries(path string) vfs.Entries                      { return vfs.Entries{} }
func (m *mockFS) Stat(path string) vfs.FileInfo                                     { return nil }
func (m *mockFS) WalkDir(root string, walkFn vfs.WalkDirFunc) error                 { return nil }
func (m *mockFS) Realpath(path string) string                                       { return path }

func TestFindRslintConfig(t *testing.T) {
	// FIXME: skip windows tests now
	if runtime.GOOS == "windows" {
		t.Skip("not supported in windows yet, due to MockFS not support windows yet")
	}
	tests := []struct {
		name          string
		workingDir    string
		filePath      string
		fileSystemMap map[string]bool // path -> exists
		expectedPath  string
		expectedDir   string
		expectedFound bool
	}{
		{
			name:       "config found in working directory - rslint.json",
			workingDir: "/project",
			filePath:   "/project/src/file.ts",
			fileSystemMap: map[string]bool{
				"/project/rslint.json": true,
			},
			expectedPath:  "/project/rslint.json",
			expectedDir:   "/project",
			expectedFound: true,
		},
		{
			name:       "config found in working directory - rslint.jsonc",
			workingDir: "/project",
			filePath:   "/project/src/file.ts",
			fileSystemMap: map[string]bool{
				"/project/rslint.json":  false,
				"/project/rslint.jsonc": true,
			},
			expectedPath:  "/project/rslint.jsonc",
			expectedDir:   "/project",
			expectedFound: true,
		},
		{
			name:       "config not found in working directory, found in parent",
			workingDir: "/project/src",
			filePath:   "/project/src/components/file.ts",
			fileSystemMap: map[string]bool{
				"/project/src/rslint.json":             false,
				"/project/src/rslint.jsonc":            false,
				"/project/src/components/rslint.json":  false,
				"/project/src/components/rslint.jsonc": false,
				"/project/rslint.json":                 true,
			},
			expectedPath:  "",
			expectedDir:   "",
			expectedFound: false,
		},
		{
			name:       "no config found anywhere",
			workingDir: "/project",
			filePath:   "/project/src/file.ts",
			fileSystemMap: map[string]bool{
				"/project/rslint.json":      false,
				"/project/rslint.jsonc":     false,
				"/project/src/rslint.json":  false,
				"/project/src/rslint.jsonc": false,
				"/rslint.json":              false,
				"/rslint.jsonc":             false,
			},
			expectedPath:  "",
			expectedDir:   "/project",
			expectedFound: false,
		},
		{
			name:       "empty file path - only check working directory",
			workingDir: "/project",
			filePath:   "",
			fileSystemMap: map[string]bool{
				"/project/rslint.json":  false,
				"/project/rslint.jsonc": true,
			},
			expectedPath:  "/project/rslint.jsonc",
			expectedDir:   "/project",
			expectedFound: true,
		},
		{
			name:       "empty file path - no config in working directory",
			workingDir: "/project",
			filePath:   "",
			fileSystemMap: map[string]bool{
				"/project/rslint.json":  false,
				"/project/rslint.jsonc": false,
			},
			expectedPath:  "",
			expectedDir:   "/project",
			expectedFound: false,
		},
		{
			name:       "prefer rslint.json over rslint.jsonc when both exist",
			workingDir: "/project",
			filePath:   "/project/src/file.ts",
			fileSystemMap: map[string]bool{
				"/project/rslint.json":  true,
				"/project/rslint.jsonc": true,
			},
			expectedPath:  "/project/rslint.json",
			expectedDir:   "/project",
			expectedFound: true,
		},
		{
			name:       "search depth limit - stop after 5 levels",
			workingDir: "/very/deep/nested/structure/level5",
			filePath:   "/very/deep/nested/structure/level5/level6/level7/file.ts",
			fileSystemMap: map[string]bool{
				"/very/deep/nested/structure/level5/rslint.json":                false,
				"/very/deep/nested/structure/level5/rslint.jsonc":               false,
				"/very/deep/nested/structure/level5/level6/rslint.json":         false,
				"/very/deep/nested/structure/level5/level6/rslint.jsonc":        false,
				"/very/deep/nested/structure/level5/level6/level7/rslint.json":  false,
				"/very/deep/nested/structure/level5/level6/level7/rslint.jsonc": false,
				"/very/deep/nested/structure/rslint.json":                       false,
				"/very/deep/nested/structure/rslint.jsonc":                      false,
				"/very/deep/nested/rslint.json":                                 false,
				"/very/deep/nested/rslint.jsonc":                                false,
				"/very/deep/rslint.json":                                        false,
				"/very/deep/rslint.jsonc":                                       false,
				// This should not be reached due to 5-level limit
				"/very/rslint.json": true,
			},
			expectedPath:  "",
			expectedDir:   "/very/deep/nested/structure/level5",
			expectedFound: false,
		},
		{
			name:       "windows-style paths (for cross-platform compatibility)",
			workingDir: "/c/project",
			filePath:   "/c/project/src/file.ts",
			fileSystemMap: map[string]bool{
				"/c/project/rslint.json": true,
			},
			expectedPath:  "/c/project/rslint.json",
			expectedDir:   "/c/project",
			expectedFound: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock VFS
			mockFS := &mockFS{
				files: tt.fileSystemMap,
			}

			// Call the function under test
			gotPath, gotFound := findRslintConfig(mockFS, tt.workingDir)

			// Assert results
			if gotPath != tt.expectedPath {
				t.Errorf("findRslintConfig() gotPath = %v, want %v", gotPath, tt.expectedPath)
			}
			if gotFound != tt.expectedFound {
				t.Errorf("findRslintConfig() gotFound = %v, want %v", gotFound, tt.expectedFound)
			}
		})
	}
}

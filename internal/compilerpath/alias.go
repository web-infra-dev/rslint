package compilerpath

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
)

// CanRepresent reports whether ts-go's cross-platform path model can preserve
// this host path exactly. On POSIX, ts-go treats a legal backslash filename
// byte as a directory separator, so the source needs a compiler-only alias.
func CanRepresent(filePath string) bool {
	return runtime.GOOS == "windows" || !strings.ContainsRune(filePath, '\\')
}

// Alias returns a unique compiler-representable path for a host source. The
// caller overlays contents at this path and retains the lexical host path for
// config matching, diagnostics, and writes.
func Alias(filePath string, fsys vfs.FS, reserved map[string]struct{}) string {
	digest := sha256.Sum256([]byte(filePath))
	suffix := sourceSuffix(filePath)
	for salt := 0; ; salt++ {
		name := fmt.Sprintf("%x-%d%s", digest[:16], salt, suffix)
		candidate := filepath.Join(os.TempDir(), ".rslint-compiler-path", name)
		identity := string(tspath.ToPath(tspath.NormalizePath(candidate), "", true))
		if _, collision := reserved[identity]; collision || (fsys != nil && fsys.FileExists(candidate)) {
			continue
		}
		if reserved != nil {
			reserved[identity] = struct{}{}
		}
		return candidate
	}
}

func sourceSuffix(filePath string) string {
	lower := strings.ToLower(filePath)
	for _, suffix := range []string{
		".d.mts", ".d.cts", ".d.ts",
		".tsx", ".mts", ".cts", ".jsx", ".mjs", ".cjs", ".ts", ".js",
	} {
		if strings.HasSuffix(lower, suffix) {
			return filePath[len(filePath)-len(suffix):]
		}
	}
	return filepath.Ext(filePath)
}

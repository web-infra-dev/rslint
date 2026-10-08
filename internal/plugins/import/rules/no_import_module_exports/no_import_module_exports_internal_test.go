package no_import_module_exports

import "testing"

func TestCompileExceptionMatchersWindowsSeparators(t *testing.T) {
	const (
		pattern  = `C:\repo\src\bridge.js`
		fileName = "C:/repo/src/bridge.js"
	)

	if !compileExceptionMatchers([]string{pattern}, true)[0].Match(fileName) {
		t.Fatal("Windows-native exception did not match the normalized filename")
	}
	if compileExceptionMatchers([]string{pattern}, false)[0].Match(fileName) {
		t.Fatal("POSIX matching unexpectedly treated backslashes as separators")
	}
}

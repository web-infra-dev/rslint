//go:build (!unix || aix) && !windows

package suppressions

func lockBaselineFile(string, bool) (func(), error) {
	return func() {}, nil
}

//go:build linux || windows

package suppressions

// Lock a single byte beyond EOF at the same offset on Linux (OFD locks) and
// Windows (LockFileEx) so writers serialize across shared SMB volumes without
// overlapping [0, EOF) reads and while fitting inside signed 32-bit off_t.
const (
	baselineLockOffset = 0x7ffffffe
	baselineLockLength = 1
)

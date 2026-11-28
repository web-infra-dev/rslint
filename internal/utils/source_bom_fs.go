package utils

import (
	"encoding/binary"
	"fmt"
	"io/fs"
	"os"
	"strings"
	"sync"
	"unicode/utf16"
	"unsafe"

	"github.com/microsoft/TypeScript/tsc/shim/bundled"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/cachedvfs"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
)

// diskReadSemaphore bounds concurrent raw disk reads to match osvfs's read
// semaphore (128), preventing file descriptor exhaustion during parallel
// Program loading.
var diskReadSemaphore = make(chan struct{}, 128)

// NewDiskFSWithBOM returns a fresh OS-backed [vfs.FS] that captures each file's
// byte order mark during ReadFile and implements [BOMSource]. Recorded BOM
// observations live only for the returned instance's lifetime; create a fresh
// instance per lint run or request so state stays bounded and never leaks
// across unrelated disk generations.
func NewDiskFSWithBOM() vfs.FS {
	return newBOMTrackingDiskFS(osvfs.FS(), nil)
}

// CachedFS wraps base with [cachedvfs.From] while preserving [BOMSource].
func CachedFS(base vfs.FS) vfs.FS {
	return &cachedBOMSourceFS{
		FS:   cachedvfs.From(base),
		base: base,
	}
}

// BundledFS wraps base with [bundled.WrapFS] while preserving [BOMSource].
// Embedded bundled library paths report no BOM without consulting disk or base.
func BundledFS(base vfs.FS) vfs.FS {
	return &bundledBOMSourceFS{
		FS:   bundled.WrapFS(base),
		base: base,
	}
}

// NewCachedDiskFS returns a fresh bundled, metadata-cached OS filesystem that
// captures byte order marks on ReadFile and preserves [BOMSource] across every
// wrapper layer.
func NewCachedDiskFS() vfs.FS {
	return BundledFS(CachedFS(NewDiskFSWithBOM()))
}

// ReadDiskFileWithBOM reads an OS disk file once, decodes its text with the
// same UTF-8/UTF-16 handling as [NewDiskFSWithBOM], and preserves a leading
// encoding byte order mark in the returned string without allocating a tracking
// filesystem.
func ReadDiskFileWithBOM(path string) (string, bool) {
	_, rootName, rest := splitDiskPath(path)
	b, err := readDiskFileBytes(rootName, rest)
	if err != nil {
		return "", false
	}
	contents, hasBOM := decodeSourceBytes(b)
	if hasBOM {
		return BOM + contents, true
	}
	return contents, true
}

type bomTrackingDiskFS struct {
	vfs.FS
	readRaw func(rootName string, rest string) ([]byte, error)

	mu           sync.RWMutex
	epoch        uint64
	hasBOMByPath map[string]bool
}

var (
	_ vfs.FS    = (*bomTrackingDiskFS)(nil)
	_ BOMSource = (*bomTrackingDiskFS)(nil)
)

func newBOMTrackingDiskFS(
	base vfs.FS,
	readRaw func(rootName string, rest string) ([]byte, error),
) *bomTrackingDiskFS {
	return &bomTrackingDiskFS{
		FS:      base,
		readRaw: readRaw,
	}
}

func (f *bomTrackingDiskFS) ReadFile(path string) (string, bool) {
	normalized, rootName, rest := splitDiskPath(path)

	startEpoch := f.currentEpoch()
	b, err := f.readBytes(rootName, rest)
	key := f.pathKey(normalized)
	if err != nil {
		// Drop any previously recorded BOM if a subsequent read fails within
		// the same generation, and do not cache failures so fallback
		// ReadSourceBOM preserves real errors.
		f.mu.Lock()
		if f.epoch == startEpoch {
			delete(f.hasBOMByPath, key)
		}
		f.mu.Unlock()
		return "", false
	}

	contents, hasBOM := decodeSourceBytes(b)
	f.storeBOM(startEpoch, key, hasBOM)
	return contents, true
}

func (f *bomTrackingDiskFS) ReadSourceBOM(path string) (bool, error) {
	if hasBOM, ok := f.loadBOM(f.pathKey(path)); ok {
		return hasBOM, nil
	}
	return ReadSourceBOM(f.FS, path)
}

// Writes and removals invalidate the tracked generation both before and after
// the mutation so concurrent in-flight reads from an earlier epoch cannot
// publish stale observations, symlink or hard-link aliases do not retain stale
// BOM entries, and relative removals evict cached absolute keys.
func (f *bomTrackingDiskFS) WriteFile(path string, data string) error {
	f.invalidate()
	defer f.invalidate()
	return f.FS.WriteFile(path, data)
}

func (f *bomTrackingDiskFS) AppendFile(path string, data string) error {
	f.invalidate()
	defer f.invalidate()
	return f.FS.AppendFile(path, data)
}

func (f *bomTrackingDiskFS) Remove(path string) error {
	f.invalidate()
	defer f.invalidate()
	return f.FS.Remove(path)
}

func (f *bomTrackingDiskFS) readBytes(rootName string, rest string) ([]byte, error) {
	if f.readRaw != nil {
		return f.readRaw(rootName, rest)
	}
	return readDiskFileBytes(rootName, rest)
}

func splitDiskPath(path string) (normalized string, rootName string, rest string) {
	normalized = tspath.NormalizePath(path)
	rootLen := diskRootLength(normalized)
	rootName = normalized[:rootLen]
	rest = tspath.RemoveTrailingDirectorySeparator(normalized[rootLen:])
	if rest == "" {
		rest = "."
	}
	return normalized, rootName, rest
}

func readDiskFileBytes(rootName string, rest string) ([]byte, error) {
	diskReadSemaphore <- struct{}{}
	defer func() { <-diskReadSemaphore }()
	return fs.ReadFile(os.DirFS(rootName), rest)
}

// pathKey preserves exact directory and file name casing while normalizing
// separators, dot segments, and Windows drive letters. This prevents distinct
// sibling files on a case-sensitive volume mounted on a case-insensitive host
// (such as a case-sensitive volume mounted on macOS) from overwriting each
// other's cached BOM; alternate-case queries on case-insensitive volumes safely
// miss the cache and fall back to ReadSourceBOM(f.FS, path).
func (f *bomTrackingDiskFS) pathKey(path string) string {
	p := ExactPathID(path)
	if len(p) > tspath.GetRootLength(p) {
		p = tspath.RemoveTrailingDirectorySeparator(p)
	}
	return p
}

func (f *bomTrackingDiskFS) currentEpoch() uint64 {
	f.mu.RLock()
	epoch := f.epoch
	f.mu.RUnlock()
	return epoch
}

func (f *bomTrackingDiskFS) invalidate() {
	f.mu.Lock()
	f.epoch++
	f.hasBOMByPath = nil
	f.mu.Unlock()
}

func (f *bomTrackingDiskFS) loadBOM(key string) (bool, bool) {
	f.mu.RLock()
	hasBOM, ok := f.hasBOMByPath[key]
	f.mu.RUnlock()
	return hasBOM, ok
}

func (f *bomTrackingDiskFS) storeBOM(startEpoch uint64, key string, hasBOM bool) {
	f.mu.Lock()
	if f.epoch == startEpoch {
		if f.hasBOMByPath == nil {
			f.hasBOMByPath = make(map[string]bool)
		}
		f.hasBOMByPath[strings.Clone(key)] = hasBOM
	}
	f.mu.Unlock()
}

// diskRootLength matches typescript-go/tsc/internal/vfs/internal.RootLength.
func diskRootLength(normalized string) int {
	l := tspath.GetEncodedRootLength(normalized)
	if l == 0 {
		panic(fmt.Sprintf("vfs: path %q is not absolute", normalized))
	}
	if l < 0 {
		return ^l
	}
	return l
}

// decodeSourceBytes mirrors typescript-go/tsc/internal/vfs/internal.decodeBytes
// and internal/utils.readBOMHeader so ReadFile and BOM tracking stay in exact
// parity for UTF-8, UTF-16LE, UTF-16BE, and unmarked inputs.
//
// Caller transfers ownership of b; b must not be mutated after this call.
func decodeSourceBytes(b []byte) (contents string, hasBOM bool) {
	if len(b) == 0 {
		return "", false
	}
	if len(b) >= 2 {
		switch {
		case b[0] == 0xFF && b[1] == 0xFE:
			return decodeUTF16(b[2:], binary.LittleEndian), true
		case b[0] == 0xFE && b[1] == 0xFF:
			return decodeUTF16(b[2:], binary.BigEndian), true
		}
	}
	if len(b) >= 3 && b[0] == 0xEF && b[1] == 0xBB && b[2] == 0xBF {
		if len(b) == 3 {
			return "", true
		}
		return unsafe.String(&b[3], len(b)-3), true
	}
	return unsafe.String(&b[0], len(b)), false
}

func decodeUTF16(b []byte, order binary.ByteOrder) string {
	n := len(b) / 2
	if n == 0 {
		return ""
	}
	ints := make([]uint16, n)
	for i := range n {
		ints[i] = order.Uint16(b[i*2:])
	}
	return string(utf16.Decode(ints))
}

type cachedBOMSourceFS struct {
	vfs.FS
	base vfs.FS
}

var (
	_ vfs.FS    = (*cachedBOMSourceFS)(nil)
	_ BOMSource = (*cachedBOMSourceFS)(nil)
)

func (f *cachedBOMSourceFS) ReadSourceBOM(path string) (bool, error) {
	return ReadSourceBOM(f.base, path)
}

type bundledBOMSourceFS struct {
	vfs.FS
	base vfs.FS
}

var (
	_ vfs.FS    = (*bundledBOMSourceFS)(nil)
	_ BOMSource = (*bundledBOMSourceFS)(nil)
)

func (f *bundledBOMSourceFS) ReadSourceBOM(path string) (bool, error) {
	if bundled.IsBundled(path) {
		return false, nil
	}
	return ReadSourceBOM(f.base, path)
}

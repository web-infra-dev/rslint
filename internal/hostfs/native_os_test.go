package hostfs

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
)

func TestNativeOSKeepsPOSIXBackslashDistinctFromSlash(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows uses backslash as a path separator")
	}
	root := t.TempDir()
	backslash := filepath.Join(root, `a\b.js`)
	slash := filepath.Join(root, "a", "b.js")
	if err := os.WriteFile(backslash, []byte("backslash"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(slash), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(slash, []byte("slash"), 0o644); err != nil {
		t.Fatal(err)
	}

	fsys := NativeOS(osvfs.FS())
	if content, ok := fsys.ReadFile(backslash); !ok || content != "backslash" {
		t.Fatalf("ReadFile(backslash) = (%q, %t)", content, ok)
	}
	if content, ok := fsys.ReadFile(slash); !ok || content != "slash" {
		t.Fatalf("ReadFile(slash) = (%q, %t)", content, ok)
	}
	if !fsys.FileExists(backslash) || fsys.Stat(backslash) == nil {
		t.Fatal("native backslash file was not stat'ed exactly")
	}
}

func TestNativeOSReadFileMatchesOSVFSEncoding(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("native POSIX path adapter is not active on Windows")
	}
	root := t.TempDir()
	fsys := NativeOS(osvfs.FS())
	tests := []struct {
		name    string
		bytes   []byte
		content string
	}{
		{name: "utf8-bom", bytes: append([]byte{0xef, 0xbb, 0xbf}, []byte("hello")...), content: "hello"},
		{name: "utf16le", bytes: encodeUTF16Test([]uint16{'h', 'i'}, binary.LittleEndian, []byte{0xff, 0xfe}), content: "hi"},
		{name: "utf16be", bytes: encodeUTF16Test([]uint16{'h', 'i'}, binary.BigEndian, []byte{0xfe, 0xff}), content: "hi"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			filePath := filepath.Join(root, test.name+`\\file.ts`)
			if err := os.WriteFile(filePath, test.bytes, 0o644); err != nil {
				t.Fatal(err)
			}
			if content, ok := fsys.ReadFile(filePath); !ok || content != test.content {
				t.Fatalf("ReadFile() = (%q, %t), want %q", content, ok, test.content)
			}
		})
	}
}

func encodeUTF16Test(units []uint16, order binary.ByteOrder, prefix []byte) []byte {
	result := append([]byte(nil), prefix...)
	for _, unit := range units {
		var encoded [2]byte
		order.PutUint16(encoded[:], unit)
		result = append(result, encoded[:]...)
	}
	return result
}

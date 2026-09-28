// cspell:words READWRITE
package ipc

import (
	"os"
	"strconv"
	"sync/atomic"
	"testing"
	"unsafe"

	"golang.org/x/sys/windows"
)

func TestPoolWindowsCommitsOnlyUsedSlots(t *testing.T) {
	// SEC_RESERVE matches the native arena. RSS cannot detect eager pagefile
	// commitment, so inspect the section's actual page states instead.
	const secReserve = 0x04000000
	handle, err := windows.CreateFileMapping(windows.InvalidHandle, nil, windows.PAGE_READWRITE|secReserve, 0, MemoryCapacity, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer windows.CloseHandle(handle)
	view, err := windows.MapViewOfFile(handle, windows.FILE_MAP_READ, 0, 0, MemoryCapacity)
	if err != nil {
		t.Fatal(err)
	}
	defer windows.UnmapViewOfFile(view)
	state := func(offset int) uint32 {
		t.Helper()
		var info windows.MemoryBasicInformation
		if err := windows.VirtualQuery(view+uintptr(offset), &info, unsafe.Sizeof(info)); err != nil {
			t.Fatal(err)
		}
		return info.State
	}
	pool, err := openMemoryPool(MemoryMapping{Version: MemoryVersion, Handle: strconv.FormatUint(uint64(handle), 10), ProcessID: uint32(os.Getpid())})
	if err != nil {
		t.Fatal(err)
	}
	defer pool.close()
	if state(0) != windows.MEM_COMMIT || state(MemoryHeaderSize+MemorySlotSize/2) != windows.MEM_RESERVE {
		t.Fatal("opening the mapping committed unused memory pages")
	}
	text := "snapshot 😀"
	batches, _ := pool.store([]Attachment{Text(text)})
	if len(batches) != 1 {
		t.Fatal("could not commit the memory slot")
	}
	if state(MemoryHeaderSize+MemorySlotSize/2) != windows.MEM_COMMIT || state(MemoryHeaderSize+MemorySlotSize+MemorySlotSize/2) != windows.MEM_RESERVE {
		t.Fatal("memory backing was not committed one slot at a time")
	}
	// A second, read-only view must see both publication and exact attachment bytes.
	if got := atomic.LoadUint32((*uint32)(unsafe.Pointer(view))); got != batches[0].Generation {
		t.Fatalf("publication was not shared: %d", got)
	}
	bytes := unsafe.Slice((*byte)(unsafe.Pointer(view+MemoryHeaderSize)), len(text))
	if string(bytes) != text {
		t.Fatalf("snapshot was not shared: %q", bytes)
	}
	pool.release(batches)
	if next, _ := pool.store([]Attachment{Text("next")}); len(next) != 1 || state(MemoryHeaderSize+MemorySlotSize+MemorySlotSize/2) != windows.MEM_RESERVE {
		t.Fatal("reuse committed another slot")
	}
}

// cspell:words psapi JOBOBJECT Nonpaged
package main

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	psapi         = windows.NewLazySystemDLL("psapi.dll")
	performance   = psapi.NewProc("GetPerformanceInfo")
	processMemory = psapi.NewProc("GetProcessMemoryInfo")
)

// Layouts from psapi.h. SIZE_T fields use uintptr, including on Windows ARM64.
type performanceInformation struct {
	Size              uint32
	CommitTotal       uintptr
	CommitLimit       uintptr
	CommitPeak        uintptr
	PhysicalTotal     uintptr
	PhysicalAvailable uintptr
	SystemCache       uintptr
	KernelTotal       uintptr
	KernelPaged       uintptr
	KernelNonpaged    uintptr
	PageSize          uintptr
	HandleCount       uint32
	ProcessCount      uint32
	ThreadCount       uint32
}

type processMemoryCounters struct {
	Size                       uint32
	PageFaultCount             uint32
	PeakWorkingSetSize         uintptr
	WorkingSetSize             uintptr
	QuotaPeakPagedPoolUsage    uintptr
	QuotaPagedPoolUsage        uintptr
	QuotaPeakNonPagedPoolUsage uintptr
	QuotaNonPagedPoolUsage     uintptr
	PagefileUsage              uintptr
	PeakPagefileUsage          uintptr
	PrivateUsage               uintptr
}

func availableCommitMiB() (int64, error) {
	var info performanceInformation
	info.Size = uint32(unsafe.Sizeof(info))
	ok, _, err := performance.Call(uintptr(unsafe.Pointer(&info)), uintptr(info.Size))
	if ok == 0 {
		return 0, err
	}
	if info.CommitLimit < info.CommitTotal {
		return 0, nil
	}
	return int64((info.CommitLimit - info.CommitTotal) * info.PageSize / mib), nil
}

func trackTree() (func() (usage, error), error) {
	job, err := windows.CreateJobObject(nil, nil)
	if err != nil {
		return nil, err
	}
	var limits windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION
	limits.BasicLimitInformation.LimitFlags = windows.JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE
	if _, err = windows.SetInformationJobObject(job, windows.JobObjectExtendedLimitInformation, uintptr(unsafe.Pointer(&limits)), uint32(unsafe.Sizeof(limits))); err != nil {
		windows.CloseHandle(job)
		return nil, err
	}
	if err = windows.AssignProcessToJobObject(job, windows.CurrentProcess()); err != nil {
		windows.CloseHandle(job)
		return nil, fmt.Errorf("create nested process accounting job: %w", err)
	}
	// Intentionally retain the handle until process exit. Closing it here would
	// terminate this wrapper as well as its descendants. It is not inheritable.
	return func() (usage, error) {
		var info windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION
		if err := windows.QueryInformationJobObject(job, windows.JobObjectExtendedLimitInformation, uintptr(unsafe.Pointer(&info)), uint32(unsafe.Sizeof(info)), nil); err != nil {
			return usage{}, err
		}
		var ids struct {
			Assigned uint32
			Count    uint32
			IDs      [4096]uintptr
		}
		if err := windows.QueryInformationJobObject(job, windows.JobObjectBasicProcessIdList, uintptr(unsafe.Pointer(&ids)), uint32(unsafe.Sizeof(ids)), nil); err != nil {
			return usage{}, err
		}
		current := uint64(0)
		for _, id := range ids.IDs[:ids.Count] {
			process, err := windows.OpenProcess(windows.PROCESS_QUERY_INFORMATION|windows.PROCESS_VM_READ, false, uint32(id))
			if err == windows.ERROR_INVALID_PARAMETER {
				continue // The child exited between enumeration and opening it.
			}
			if err != nil {
				return usage{}, err
			}
			var counters processMemoryCounters
			counters.Size = uint32(unsafe.Sizeof(counters))
			ok, _, readErr := processMemory.Call(uintptr(process), uintptr(unsafe.Pointer(&counters)), uintptr(counters.Size))
			windows.CloseHandle(process)
			if ok == 0 {
				return usage{}, readErr
			}
			current += uint64(counters.PrivateUsage)
		}
		return usage{CurrentMiB: int64((current + mib - 1) / mib), PeakMiB: int64((info.PeakJobMemoryUsed + mib - 1) / mib)}, nil
	}, nil
}

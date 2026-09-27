param([string]$LogPath, [string]$StopPath, [int]$ParentProcessId)

$ErrorActionPreference = 'Stop'
# Temporary, read-only diagnostics. System counters can describe the host;
# the current Windows job object can have a separate, lower memory limit.
# cspell:ignore psapi DllImport StructLayout UIntPtr kernel32 Nonpaged pscustomobject
Add-Type @'
using System;
using System.Runtime.InteropServices;
public static class WindowsMemory {
    [StructLayout(LayoutKind.Sequential)]
    public struct Performance {
        public uint Size;
        public UIntPtr CommitTotal, CommitLimit, CommitPeak;
        public UIntPtr PhysicalTotal, PhysicalAvailable, SystemCache;
        public UIntPtr KernelTotal, KernelPaged, KernelNonpaged, PageSize;
        public uint HandleCount, ProcessCount, ThreadCount;
    }
    [StructLayout(LayoutKind.Sequential)]
    public struct BasicLimits {
        public long ProcessUserTimeLimit, JobUserTimeLimit;
        public uint LimitFlags;
        public UIntPtr MinimumWorkingSetSize, MaximumWorkingSetSize;
        public uint ActiveProcessLimit;
        public UIntPtr Affinity;
        public uint PriorityClass, SchedulingClass;
    }
    [StructLayout(LayoutKind.Sequential)]
    public struct IoCounters {
        public ulong ReadOperationCount, WriteOperationCount, OtherOperationCount;
        public ulong ReadTransferCount, WriteTransferCount, OtherTransferCount;
    }
    [StructLayout(LayoutKind.Sequential)]
    public struct JobLimits {
        public BasicLimits Basic;
        public IoCounters Io;
        public UIntPtr ProcessMemoryLimit, JobMemoryLimit;
        public UIntPtr PeakProcessMemoryUsed, PeakJobMemoryUsed;
    }
    [DllImport("psapi.dll", SetLastError = true)]
    public static extern bool GetPerformanceInfo(ref Performance info, uint size);
    [DllImport("kernel32.dll", SetLastError = true)]
    public static extern bool QueryInformationJobObject(IntPtr job, int type,
        ref JobLimits info, uint size, IntPtr returnedLength);
}
'@

function Write-MemoryRecord($record) {
    $record | ConvertTo-Json -Compress -Depth 6 | Add-Content -LiteralPath $LogPath
}

$performance = [WindowsMemory+Performance]::new()
$performance.Size = [Runtime.InteropServices.Marshal]::SizeOf($performance)
$limits = [WindowsMemory+JobLimits]::new()
$limitsSize = [Runtime.InteropServices.Marshal]::SizeOf($limits)
$peaks = @{}
$samples = 0
$minimumCommitAvailable = $null
$maximumVisibleGoPrivate = 0
Write-MemoryRecord @{kind = 'start'; utc = [DateTime]::UtcNow.ToString('o'); sample_interval_seconds = 1}
try {
    while (!(Test-Path -LiteralPath $StopPath) -and (Get-Process -Id $ParentProcessId -ErrorAction SilentlyContinue)) {
        $system = $null
        $systemError = 0
        if ([WindowsMemory]::GetPerformanceInfo([ref]$performance, $performance.Size)) {
            $pageMiB = $performance.PageSize.ToUInt64() / 1MB
            $available = ($performance.CommitLimit.ToUInt64() - $performance.CommitTotal.ToUInt64()) * $pageMiB
            if ($null -eq $minimumCommitAvailable -or $available -lt $minimumCommitAvailable) { $minimumCommitAvailable = $available }
            $system = @{
                commit_mib = $performance.CommitTotal.ToUInt64() * $pageMiB
                commit_limit_mib = $performance.CommitLimit.ToUInt64() * $pageMiB
                commit_available_mib = $available
                physical_available_mib = $performance.PhysicalAvailable.ToUInt64() * $pageMiB
            }
        } else { $systemError = [Runtime.InteropServices.Marshal]::GetLastWin32Error() }
        $job = $null
        $jobError = 0
        if ([WindowsMemory]::QueryInformationJobObject([IntPtr]::Zero, 9, [ref]$limits, $limitsSize, [IntPtr]::Zero)) {
            $job = @{
                limit_flags = $limits.Basic.LimitFlags
                memory_limit_mib = $(if ($limits.Basic.LimitFlags -band 0x200) { $limits.JobMemoryLimit.ToUInt64() / 1MB } else { $null })
                process_memory_limit_mib = $(if ($limits.Basic.LimitFlags -band 0x100) { $limits.ProcessMemoryLimit.ToUInt64() / 1MB } else { $null })
                peak_process_mib = $limits.PeakProcessMemoryUsed.ToUInt64() / 1MB
                peak_job_mib = $limits.PeakJobMemoryUsed.ToUInt64() / 1MB
            }
        } else { $jobError = [Runtime.InteropServices.Marshal]::GetLastWin32Error() }
        $processes = @()
        $visiblePrivate = 0
        $unreadable = 0
        foreach ($process in Get-Process) {
            try {
                $private = $process.PrivateMemorySize64 / 1MB
                $visiblePrivate += $private
                if ($process.ProcessName -in @('go', 'compile', 'link', 'vet') -or $process.ProcessName -like '*.test') {
                    $record = [pscustomobject]@{
                        id = $process.Id; name = $process.ProcessName
                        started_at = $process.StartTime.ToUniversalTime().ToString('o')
                        private_mib = [Math]::Round($private, 2)
                    }
                    $processes += $record
                    $key = "$($record.id)/$($record.started_at)"
                    if (!$peaks.ContainsKey($key) -or $private -gt $peaks[$key].private_mib) { $peaks[$key] = $record }
                }
            } catch { $unreadable++ }
            finally { $process.Dispose() }
        }
        $goPrivate = ($processes | Measure-Object -Property private_mib -Sum).Sum
        $maximumVisibleGoPrivate = [Math]::Max($maximumVisibleGoPrivate, $goPrivate)
        Write-MemoryRecord @{
            kind = 'sample'; utc = [DateTime]::UtcNow.ToString('o')
            system = $system; job_object = $job
            system_query_error = $systemError; job_query_error = $jobError
            visible_private_mib = [Math]::Round($visiblePrivate, 2)
            visible_go_private_mib = $goPrivate; unreadable_processes = $unreadable
            processes = $processes
        }
        $samples++
        Start-Sleep -Seconds 1
    }
} finally {
    Write-MemoryRecord @{
        kind = 'summary'; utc = [DateTime]::UtcNow.ToString('o'); samples = $samples
        minimum_system_commit_available_mib = $minimumCommitAvailable
        maximum_visible_go_private_mib = $maximumVisibleGoPrivate
        largest_sampled_processes = @($peaks.Values | Sort-Object -Property private_mib -Descending | Select-Object -First 20)
    }
}

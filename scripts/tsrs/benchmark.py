#!/usr/bin/env python3
"""Validate or run a full-rslint, paired tsgo/tsrs benchmark (Python 3.9+).

The default action validates inputs and never executes benchmark commands.
  python3 scripts/tsrs/benchmark.py --manifest benchmark.json
  python3 scripts/tsrs/benchmark.py --manifest benchmark.json --output results \
      --lock /path/to/shared-machine.lock --run
  python3 scripts/tsrs/benchmark.py --example > benchmark.example.json

All manifest paths are relative to the manifest containing them. argv is an
argument array, never shell text. A command accepts cwd, env, inherit_env,
expected_exit_codes, timeout_seconds, and process_expectations. Only explicitly
listed environment variables are inherited. {manifest_dir}, {output_dir},
{work_manifest}, {observation} and {case_id} are supported argv placeholders.
Executable paths containing '/' are relative to command cwd. Bare executable
names use that command's explicit PATH. Artifacts pin the built executables and
JS/bridge entry points; versions/build_conditions document their provenance.

Each case's separate work_manifest contains schema_version=1, roots for both
backends, nonempty files/configs arrays of {path,sha256}, a rules array covering
every file ({file,rules:[{name,options,requires_type_information}]}), and check_scope.
Paths in files/configs are relative to each backend root. The complete manifest
is hashed; actual input bytes are checked before and after every invocation.

Normalization compares the ENTIRE parsed output and exit code. Its only allowed
changes are explicit path_fields (JSON pointers, '*' matches one array/object
level), named path_roots replaced by '<name>', and unordered_arrays sorted by
complete JSON values without deduplication. Required_json_pointers must exist.
No key, position, message, fix, or other output field can be dropped. stderr is
compared byte-for-byte. json and jsonl are supported. The verifier must expose
all relevant diagnostics/fix results in these structured outputs.

Each fixture and measured case requires a proof_command. It receives the saved
observation via {observation}, verifies actual file/rule/check scope and fixes,
and emits ONE JSON object with schema_version=1, status='passed',
scope='complete-lint', observation_sha256, expected_work_sha256, and backends.
Each backends.{baseline,candidate} object must contain the observation's exact
output_sha256, exit_code, and work_fingerprint. checks must contain true values
for workload, diagnostics, positions, messages, fixes, type_information and
no_tsgo_fallback. Fixture proofs additionally attest hit, no_hit,
type_aware_hit and type_aware_no_hit in fixture_coverage. A proof is not inferred
from an empty output or an exit code; its command, output and hashes are saved.
The harness cannot itself establish that a separately authored verifier is
correct. The verifier and backend instrumentation must be reviewed with the PR.

RSS is the maximum SAMPLE of the sum of resident bytes observed in the
owned process tree, including children (shared pages may be counted repeatedly).
It is NOT the sum of per-process peaks and NOT a kernel-measured true tree peak.
Process RSS reads within one sample are sequential, not an atomic OS snapshot.
PID start times guard against PID reuse. Process-group discovery and retained
live children cover observed reparenting; a child that spawns, changes session
and exits entirely between samples can be missed. Each sample records processes,
coverage errors and gaps. Wall time includes process start, compiler startup,
IPC/conversion and exit; parent wait is measured independently of the sampler.
No historical PoC result is consumed or extrapolated by this harness.
"""

import argparse
import ctypes
import errno
import fcntl
import hashlib
import json
import os
from pathlib import Path
import platform
import re
import shutil
import signal
import statistics
import subprocess
import sys
import threading
import tempfile
import time

BACKENDS = ("baseline", "candidate")
PROOF_CHECKS = ("workload", "diagnostics", "positions", "messages", "fixes",
                "type_information", "no_tsgo_fallback")
FIXTURE_CHECKS = ("hit", "no_hit", "type_aware_hit", "type_aware_no_hit")
MEMORY_METHOD = (
    "maximum sampled sum of contemporaneous RSS observations of root and descendants; includes "
    "known reparented descendants and the private process group; shared pages may "
    "be counted repeatedly; excludes harness; short-lived detached processes and "
    "peaks between samples may be missed; never sum individual process peaks"
)


class BenchmarkError(RuntimeError):
    pass


def require(condition, message):
    if not condition:
        raise BenchmarkError(message)


def canonical(value):
    return json.dumps(value, sort_keys=True, separators=(",", ":"),
                      ensure_ascii=False, allow_nan=False).encode("utf-8")


def digest_bytes(value):
    return hashlib.sha256(value).hexdigest()


def file_digest(path):
    h = hashlib.sha256()
    with Path(path).open("rb") as source:
        for chunk in iter(lambda: source.read(1 << 20), b""):
            h.update(chunk)
    return h.hexdigest()


def save_json(path, value):
    Path(path).write_text(json.dumps(value, indent=2, ensure_ascii=False,
                                    allow_nan=False) + "\n", encoding="utf-8")


def unique_object(pairs):
    result = {}
    for key, value in pairs:
        require(key not in result, "duplicate JSON object key: " + key)
        result[key] = value
    return result


def parse_json(raw):
    def invalid_constant(value):
        raise BenchmarkError("non-finite JSON number: " + value)
    return json.loads(raw, object_pairs_hook=unique_object,
                      parse_constant=invalid_constant)


def read_json(path):
    return parse_json(Path(path).read_text(encoding="utf-8"))


def relative_path(base, value, label):
    require(isinstance(value, str) and value != "" and not Path(value).is_absolute(),
            label + " must be a nonempty relative path")
    return (base / value).resolve()


def valid_sha(value):
    return isinstance(value, str) and re.fullmatch(r"[a-f0-9]{64}", value) is not None


def load_work(path):
    work = read_json(path)
    require(work.get("schema_version") == 1, "work manifest schema_version must be 1")
    roots = work.get("roots", {})
    resolved = {backend: relative_path(path.parent, roots.get(backend), "work root")
                for backend in BACKENDS}
    for category in ("files", "configs"):
        rows = work.get(category)
        require(isinstance(rows, list) and rows, "work manifest needs nonempty " + category)
        seen = set()
        for row in rows:
            name = row.get("path")
            require(isinstance(name, str) and name and not Path(name).is_absolute()
                    and ".." not in Path(name).parts and name not in seen,
                    "work file paths must be unique relative paths within each root")
            require(valid_sha(row.get("sha256")), "work file needs a SHA-256: " + name)
            seen.add(name)
    rules = work.get("rules")
    require(isinstance(rules, list) and rules, "work manifest needs per-file effective rules")
    by_file = {}
    for entry in rules:
        name = entry.get("file")
        require(name not in by_file, "duplicate per-file rule entry")
        enabled = entry.get("rules")
        require(isinstance(enabled, list), "each rule entry requires a rules array")
        names = set()
        for rule in enabled:
            require(isinstance(rule.get("name"), str) and rule["name"]
                    and rule["name"] not in names and "options" in rule
                    and isinstance(rule.get("requires_type_information"), bool),
                    "each enabled rule needs a unique name, options and type-information flag")
            names.add(rule["name"])
        by_file[name] = enabled
    require(set(by_file) == {row["path"] for row in work["files"]},
            "effective rules must cover exactly the expected checked file set")
    require(isinstance(work.get("check_scope"), dict) and work["check_scope"],
            "work manifest needs explicit check_scope")
    fingerprint = {
        "files_sha256": digest_bytes(canonical(sorted(work["files"], key=lambda r: r["path"]))),
        "configs_sha256": digest_bytes(canonical(sorted(work["configs"], key=lambda r: r["path"]))),
        "rules_sha256": digest_bytes(canonical(sorted(rules, key=lambda r: r["file"]))),
        "check_scope_sha256": digest_bytes(canonical(work["check_scope"])),
        "checked_file_count": len(work["files"]),
    }
    return {"path": path, "data": work, "roots": resolved,
            "sha256": file_digest(path), "fingerprint": fingerprint}


def verify_work(work):
    require(file_digest(work["path"]) == work["sha256"], "work manifest changed during batch")
    for backend, root in work["roots"].items():
        for category in ("files", "configs"):
            for row in work["data"][category]:
                path = root / row["path"]
                require(path.is_file() and file_digest(path) == row["sha256"],
                        "input bytes differ: {} {} {}".format(backend, category, row["path"]))


def command_env(command):
    inherited = command.get("inherit_env")
    additions = command.get("env")
    require(isinstance(inherited, list) and all(isinstance(k, str) for k in inherited),
            "command inherit_env must explicitly list inherited variables")
    require(isinstance(additions, dict) and all(isinstance(k, str) and isinstance(v, str)
                                              for k, v in additions.items()),
            "command env must be an explicit string map")
    env = {key: os.environ[key] for key in inherited if key in os.environ}
    env.update(additions)
    return env


def expand_argument(value, variables):
    require(isinstance(value, str) and "\0" not in value, "argv must contain non-NUL strings")
    for name, replacement in variables.items():
        value = value.replace("{" + name + "}", str(replacement))
    require(not re.search(r"\{[a-z_]+\}", value), "unknown argv placeholder: " + value)
    return value


def prepare_command(command, base, variables):
    argv = command.get("argv")
    require(isinstance(argv, list) and argv and all(isinstance(a, str) for a in argv),
            "command argv must be a nonempty string array")
    argv = [expand_argument(arg, variables) for arg in argv]
    cwd = relative_path(base, command.get("cwd"), "command cwd")
    require(cwd.is_dir(), "command cwd does not exist: " + str(cwd))
    env = command_env(command)
    executable = (cwd / argv[0]).resolve() if "/" in argv[0] else shutil.which(argv[0], path=env.get("PATH", ""))
    require(executable is not None and Path(executable).is_file(), "executable is unavailable: " + argv[0])
    argv[0] = str(executable)
    require(not any(label in arg for arg in argv for label in ("rslint-poc", "tsrs-poc", "rslint-facts")),
            "historical partial-rule PoC commands cannot be a full-backend benchmark")
    codes = command.get("expected_exit_codes")
    require(isinstance(codes, list) and codes and all(type(c) is int and 0 <= c < 126 for c in codes),
            "command needs explicit expected_exit_codes")
    timeout = command.get("timeout_seconds")
    require(isinstance(timeout, (int, float)) and 0 < timeout <= 86400, "command needs a bounded timeout_seconds")
    return argv, cwd, env


def parse_pointer(pointer):
    require(isinstance(pointer, str) and (pointer == "" or pointer.startswith("/")),
            "normalization paths must be JSON pointers")
    if not pointer:
        return []
    parts = pointer[1:].split("/")
    require(all(re.search(r"~(?![01])", part) is None for part in parts), "invalid JSON pointer escape")
    return [part.replace("~1", "/").replace("~0", "~") for part in parts]


def map_pointer(value, parts, transform):
    if not parts:
        return transform(value), 1
    head, tail = parts[0], parts[1:]
    count = 0
    if isinstance(value, list):
        indexes = range(len(value)) if head == "*" else ([int(head)] if head.isdigit() and int(head) < len(value) else [])
        for index in indexes:
            value[index], found = map_pointer(value[index], tail, transform)
            count += found
    elif isinstance(value, dict):
        keys = list(value) if head == "*" else ([head] if head in value else [])
        for key in keys:
            value[key], found = map_pointer(value[key], tail, transform)
            count += found
    return value, count


def normalize_output(raw, policy, backend, base):
    require(set(policy) <= {"format", "path_fields", "path_roots", "unordered_arrays", "required_json_pointers"},
            "normalization may not remove fields or apply undeclared transforms")
    fmt = policy.get("format")
    if fmt == "json":
        data = parse_json(raw.decode("utf-8"))
    elif fmt == "jsonl":
        data = [parse_json(line) for line in raw.decode("utf-8").splitlines() if line.strip()]
    else:
        raise BenchmarkError("normalization format must be json or jsonl")
    required = policy.get("required_json_pointers")
    require(isinstance(required, list) and required, "declare required structured diagnostic/fix fields")
    for pointer in required:
        data, count = map_pointer(data, parse_pointer(pointer), lambda value: value)
        require(count > 0, "required output field is missing: " + pointer)
    roots = policy.get("path_roots", {}).get(backend, {})
    prefixes = [(str(relative_path(base, path, "normalization root")), "<" + name + ">")
                for name, path in roots.items()]
    prefixes.sort(key=lambda pair: len(pair[0]), reverse=True)
    def path_value(value):
        require(isinstance(value, str), "declared path output field is not a string")
        for prefix, replacement in prefixes:
            if value == prefix or value.startswith(prefix.rstrip("/") + "/"):
                return replacement + value[len(prefix):]
        return value
    for pointer in policy.get("path_fields", []):
        data, _ = map_pointer(data, parse_pointer(pointer), path_value)
    def unordered(value):
        require(isinstance(value, list), "unordered_arrays pointer does not identify an array")
        return sorted(value, key=canonical)
    for pointer in policy.get("unordered_arrays", []):
        data, _ = map_pointer(data, parse_pointer(pointer), unordered)
    return data


class MacProcessProbe:
    """libproc ABI from Apple's SDK; convenience pid-list calls return PID counts.

    Audited against apple-oss-distributions/xnu, libsyscall/wrappers/libproc/libproc.c:
    https://github.com/apple-oss-distributions/xnu/blob/main/libsyscall/wrappers/libproc/libproc.c
    They divide proc_listpids' byte count by sizeof(int); buffersize stays bytes.
    """
    class BSDInfo(ctypes.Structure):
        _fields_ = [(name, ctypes.c_uint32) for name in (
            "flags", "status", "xstatus", "pid", "ppid", "uid", "gid", "ruid", "rgid", "svuid", "svgid", "reserved")]
        _fields_ += [("comm", ctypes.c_char * 16), ("name", ctypes.c_char * 32)]
        _fields_ += [(name, ctypes.c_uint32) for name in ("nfiles", "pgid", "jobc", "tdev", "tpgid")]
        _fields_ += [("nice", ctypes.c_int32), ("start_sec", ctypes.c_uint64), ("start_usec", ctypes.c_uint64)]
    class TaskInfo(ctypes.Structure):
        _fields_ = [(name, ctypes.c_uint64) for name in ("virtual", "resident", "user", "system", "threads_user", "threads_system")]
        _fields_ += [("counter" + str(index), ctypes.c_int32) for index in range(12)]

    def __init__(self):
        self.lib = ctypes.CDLL("/usr/lib/libproc.dylib", use_errno=True)
        self.lib.proc_pidinfo.argtypes = [ctypes.c_int, ctypes.c_int, ctypes.c_uint64, ctypes.c_void_p, ctypes.c_int]
        self.lib.proc_pidinfo.restype = ctypes.c_int
        self.lib.proc_pidpath.argtypes = [ctypes.c_int, ctypes.c_void_p, ctypes.c_uint32]
        self.lib.proc_pidpath.restype = ctypes.c_int
        for name in ("proc_listchildpids", "proc_listpgrppids"):
            function = getattr(self.lib, name)
            function.argtypes = [ctypes.c_int, ctypes.c_void_p, ctypes.c_int]
            function.restype = ctypes.c_int

    def list_pids(self, name, owner):
        function = getattr(self.lib, name)
        ctypes.set_errno(0)
        count = function(owner, None, 0)
        error = ctypes.get_errno()
        if error not in (0, errno.ESRCH, errno.ENOENT):
            raise OSError(error, name)
        capacity = max(64, count + 32)
        while capacity <= (1 << 20):
            buffer = (ctypes.c_int * capacity)()
            ctypes.set_errno(0)
            count = function(owner, buffer, ctypes.sizeof(buffer))
            error = ctypes.get_errno()
            if error not in (0, errno.ESRCH, errno.ENOENT):
                raise OSError(error, name)
            require(count >= 0, name + " failed")
            if count < capacity:
                return [pid for pid in buffer[:count] if pid > 0]
            capacity *= 2
        raise BenchmarkError("libproc PID list exceeded explicit memory bound")

    def info(self, pid):
        bsd = self.BSDInfo()
        ctypes.set_errno(0)
        count = self.lib.proc_pidinfo(pid, 3, 0, ctypes.byref(bsd), ctypes.sizeof(bsd))
        error = ctypes.get_errno()
        if count == 0 and error in (0, errno.ESRCH, errno.ENOENT):
            return None
        if count != ctypes.sizeof(bsd):
            raise OSError(error or errno.EIO, "proc_pidinfo BSD " + str(pid))
        if bsd.status == 5:  # SZOMB; it owns no resident address space.
            return None
        task = self.TaskInfo()
        ctypes.set_errno(0)
        count = self.lib.proc_pidinfo(pid, 4, 0, ctypes.byref(task), ctypes.sizeof(task))
        error = ctypes.get_errno()
        if count == 0 and error in (0, errno.ESRCH, errno.ENOENT):
            return None
        if count != ctypes.sizeof(task):
            raise OSError(error or errno.EIO, "proc_pidinfo TASK " + str(pid))
        check = self.BSDInfo()
        count = self.lib.proc_pidinfo(pid, 3, 0, ctypes.byref(check), ctypes.sizeof(check))
        if count != ctypes.sizeof(check) or (check.start_sec, check.start_usec) != (bsd.start_sec, bsd.start_usec):
            return None
        return {"pid": pid, "ppid": bsd.ppid, "pgid": bsd.pgid,
                "identity": "{}:{}:{}".format(pid, bsd.start_sec, bsd.start_usec),
                "rss_bytes": int(task.resident), "comm": bytes(bsd.comm).split(b"\0", 1)[0].decode("utf-8", "replace")}

    def executable(self, pid):
        buffer = ctypes.create_string_buffer(4096)
        count = self.lib.proc_pidpath(pid, buffer, len(buffer))
        return os.path.basename(os.fsdecode(buffer.value)) if count > 0 else None

    def refresh(self):
        pass

    def group_pids(self, root):
        return self.list_pids("proc_listpgrppids", root)

    def children(self, pid):
        return self.list_pids("proc_listchildpids", pid)


class LinuxProcessProbe:
    def __init__(self):
        self.page_size = os.sysconf("SC_PAGE_SIZE")
        self.rows = {}

    def refresh(self):
        rows = {}
        for entry in Path("/proc").iterdir():
            if not entry.name.isdigit():
                continue
            try:
                data = (entry / "stat").read_text()
                start, end = data.index("("), data.rindex(")")
                fields = data[end + 2:].split()
                if fields[0] in ("Z", "X"):
                    continue
                pid = int(entry.name)
                rows[pid] = {"pid": pid, "ppid": int(fields[1]), "pgid": int(fields[2]),
                             "identity": "{}:{}".format(pid, fields[19]),
                             "rss_bytes": int(fields[21]) * self.page_size,
                             "comm": data[start + 1:end]}
            except (FileNotFoundError, ProcessLookupError):
                pass
            except PermissionError:
                # Access is required only for the target tree. Missing unrelated
                # process rows do not count as a target RSS observation.
                rows[int(entry.name)] = {"permission_denied": True}
        self.rows = rows

    def group_pids(self, root):
        return [pid for pid, row in self.rows.items() if row.get("pgid") == root]

    def children(self, parent):
        return [pid for pid, row in self.rows.items() if row.get("ppid") == parent]

    def info(self, pid):
        # Re-read immediately rather than trusting enumeration's PID identity.
        try:
            data = Path("/proc/{}/stat".format(pid)).read_text()
            start, end = data.index("("), data.rindex(")")
            fields = data[end + 2:].split()
            if fields[0] in ("Z", "X"):
                return None
            return {"pid": pid, "ppid": int(fields[1]), "pgid": int(fields[2]),
                    "identity": "{}:{}".format(pid, fields[19]),
                    "rss_bytes": int(fields[21]) * self.page_size,
                    "comm": data[start + 1:end]}
        except (FileNotFoundError, ProcessLookupError):
            return None

    def executable(self, pid):
        try:
            return os.path.basename(os.readlink("/proc/{}/exe".format(pid)))
        except (FileNotFoundError, ProcessLookupError):
            return None


class ProcessTree:
    def __init__(self, pid):
        self.root = pid
        self.probe = MacProcessProbe() if sys.platform == "darwin" else LinuxProcessProbe()
        self.known = {}
        self.executables = {}
        self.reparented = set()
        self.reused = set()

    def sample(self):
        self.probe.refresh()
        root = self.probe.info(self.root)
        root_reused = (root is not None and self.root in self.known
                       and self.known[self.root] != root["identity"])
        # Never traverse a reused PID's children or a new unrelated process
        # group that reused the old leader's number.
        group = [] if root_reused else self.probe.group_pids(self.root)
        pending = [(pid, None) for pid in self.known]
        pending += [(pid, "group") for pid in group]
        pending.append((self.root, "root"))
        visited = set()
        rows = []
        while pending:
            pid, origin = pending.pop()
            if pid in visited:
                continue
            visited.add(pid)
            row = self.probe.info(pid)
            if row is None:
                continue
            old = self.known.get(pid)
            if old is not None and old != row["identity"]:
                self.reused.add(pid)
                continue
            if old is None:
                if origin == "group" and row["pgid"] != self.root:
                    continue
                if isinstance(origin, int) and row["ppid"] != origin and row["pgid"] != self.root:
                    # It may have detached before its first observation, or
                    # the enumerated PID may already have been reused. It
                    # cannot safely be attributed without a prior identity.
                    continue
            self.known[pid] = row["identity"]
            executable = self.probe.executable(pid)
            if executable:
                self.executables[row["identity"]] = executable
            if pid != self.root and row["ppid"] not in self.known:
                self.reparented.add(row["identity"])
            rows.append(row)
            pending.extend((child, pid) for child in self.probe.children(pid))
        return rows

    def terminate_owned(self):
        # Never kill a PID merely because an older process had that number.
        for sig in (signal.SIGTERM, signal.SIGKILL):
            rows = self.sample()
            for row in rows:
                if self.known.get(row["pid"]) == row["identity"]:
                    try:
                        os.kill(row["pid"], sig)
                    except ProcessLookupError:
                        pass
            if sig == signal.SIGTERM:
                time.sleep(0.1)


def clock_observation(monotonic_elapsed, realtime_elapsed):
    # macOS monotonic time can exclude system sleep. A clock step is also not a
    # trustworthy end-to-end observation; preserve both durations and reject it.
    skew = realtime_elapsed - monotonic_elapsed
    tolerance = max(0.1, monotonic_elapsed * 0.01)
    return {"monotonic_seconds": monotonic_elapsed, "realtime_seconds": realtime_elapsed,
            "skew_seconds": skew, "tolerance_seconds": tolerance,
            "consistent": abs(skew) <= tolerance}


def run_once(command, base, variables, prefix, interval, lock_fd=None, require_rss=True):
    argv, cwd, env = prepare_command(command, base, variables)
    prefix.parent.mkdir(parents=True, exist_ok=True)
    record = {"argv": argv, "cwd": str(cwd), "env": env, "started_unix": time.time(),
              "load_before": list(os.getloadavg()), "sample_interval_seconds": interval,
              "memory_method": MEMORY_METHOD}
    realtime_started = time.time()
    started = time.perf_counter()
    done = threading.Event()
    wait_result = {}
    tree = None
    process = None
    samples = 0
    peak = 0
    count_peak = 0
    previous = 0.0
    gap_max = 0.0
    sample_duration_max = 0.0
    individual_peaks = {}
    child_after_parent = False
    failure = None
    try:
        with Path(str(prefix) + ".stdout").open("wb") as stdout, Path(str(prefix) + ".stderr").open("wb") as stderr, Path(str(prefix) + ".rss.jsonl").open("w", encoding="utf-8") as rss_log:
            process = subprocess.Popen(argv, cwd=cwd, env=env, stdout=stdout, stderr=stderr,
                                       start_new_session=True, pass_fds=(() if lock_fd is None else (lock_fd,)))
            tree = ProcessTree(process.pid)
            def wait_parent():
                wait_result["exit_code"] = process.wait()
                wait_result["wall_seconds"] = time.perf_counter() - started
                done.set()
            waiter = threading.Thread(target=wait_parent, daemon=True)
            waiter.start()
            while True:
                sample_started = time.perf_counter() - started
                rows = tree.sample()
                now = time.perf_counter() - started
                sample_duration_max = max(sample_duration_max, now - sample_started)
                total = sum(row["rss_bytes"] for row in rows)
                for row in rows:
                    key = row["identity"]
                    individual_peaks[key] = max(individual_peaks.get(key, 0), row["rss_bytes"])
                peak, count_peak = max(peak, total), max(count_peak, len(rows))
                gap_max = max(gap_max, now - previous)
                previous = now
                samples += 1
                rss_log.write(json.dumps({"seconds": now, "sample_duration_seconds": now - sample_started,
                                          "rss_bytes": total, "processes": rows}, separators=(",", ":")) + "\n")
                if done.is_set():
                    alive_children = [row for row in rows if row["pid"] != process.pid]
                    if not alive_children:
                        break
                    child_after_parent = True
                if now > command["timeout_seconds"]:
                    raise BenchmarkError("command exceeded timeout_seconds")
                done.wait(interval) if not done.is_set() else time.sleep(interval)
            waiter.join()
            wall = previous if child_after_parent else wait_result["wall_seconds"]
            record.update(exit_code=wait_result["exit_code"], wall_seconds=wall,
                          root_wall_seconds=wait_result["wall_seconds"],
                          wall_observation_uncertainty_seconds=(gap_max if child_after_parent else sample_duration_max))
    except BaseException as error:
        failure = error
        record["error"] = str(error)
        if tree is not None:
            try:
                tree.terminate_owned()
            except BaseException as cleanup_error:
                record["cleanup_error"] = str(cleanup_error)
        if process is not None:
            if process.poll() is None:
                process.kill()
            process.wait(timeout=10)
            record["exit_code"] = process.returncode
        record["wall_seconds"] = time.perf_counter() - started
    finally:
        clocks = clock_observation(time.perf_counter() - started, time.time() - realtime_started)
        if not clocks["consistent"] and failure is None:
            failure = BenchmarkError("realtime/monotonic divergence; possible system sleep or clock step: " + str(prefix))
            record["error"] = str(failure)
        record.update(sampled_tree_peak_rss_bytes=peak, sample_count=samples,
                      peak_process_count=count_peak, max_sample_gap_seconds=gap_max,
                      max_sample_duration_seconds=sample_duration_max,
                      observed_children_after_parent_exit=child_after_parent,
                      individual_sampled_peaks_bytes=individual_peaks,
                      executables=({} if tree is None else tree.executables),
                      reparented_processes=([] if tree is None else sorted(tree.reparented)),
                      rejected_reused_pids=([] if tree is None else sorted(tree.reused)),
                      load_after=list(os.getloadavg()), clocks=clocks)
        save_json(str(prefix) + ".record.json", record)
    if failure is not None:
        raise failure
    require(record["exit_code"] in command["expected_exit_codes"], "unexpected exit code: " + str(prefix))
    require(not require_rss or (peak > 0 and tree.root in tree.known),
            "sampler did not observe root and resident target processes: " + str(prefix))
    expected = command.get("process_expectations", {})
    require(count_peak >= expected.get("minimum_peak_count", 1 if require_rss else 0), "not all expected simultaneous processes were observed")
    seen = set(record["executables"].values())
    require(set(expected.get("required_executables", [])) <= seen,
            "sampler missed required executables: " + str(set(expected.get("required_executables", [])) - seen))
    return record


def observed_output(prefix, policy, backend, base, record):
    path = Path(str(prefix) + ".stdout")
    require(path.stat().st_size <= (256 << 20), "structured output exceeds 256 MiB bound")
    value = normalize_output(path.read_bytes(), policy, backend, base)
    normalized = canonical(value)
    Path(str(prefix) + ".normalized.json").write_bytes(normalized + b"\n")
    return {"output_sha256": digest_bytes(normalized), "exit_code": record["exit_code"],
            "stderr_sha256": file_digest(str(prefix) + ".stderr")}


def verify_artifacts(manifest, base):
    for backend in BACKENDS:
        rows = manifest["backends"][backend].get("artifacts")
        require(isinstance(rows, list) and rows, "pin executable and bridge artifacts for " + backend)
        for row in rows:
            path = relative_path(base, row.get("path"), "built artifact")
            require(valid_sha(row.get("sha256")) and path.is_file() and file_digest(path) == row["sha256"],
                    "built artifact differs: " + str(path))


def validate_manifest(path, output):
    manifest = read_json(path)
    base = path.parent
    require(manifest.get("schema_version") == 1 and manifest.get("scope") == "complete-lint", "manifest must declare schema_version=1 and scope=complete-lint")
    require(manifest.get("warmups", 2) >= 2 and manifest.get("repeats", 5) >= 5,
            "at least two warmups and five interleaved repetitions are required")
    require(type(manifest.get("warmups", 2)) is int and type(manifest.get("repeats", 5)) is int,
            "warmups and repeats must be integers")
    require(0.0005 <= manifest.get("sample_interval_seconds", 0.002) <= 0.1, "sample interval must be between 0.5 and 100 ms")
    require(sys.platform in ("darwin", "linux"), "process-tree sampling supports macOS and Linux")
    for backend, compiler in (("baseline", "tsgo"), ("candidate", "tsrs")):
        info = manifest.get("backends", {}).get(backend, {})
        require(info.get("compiler") == compiler and isinstance(info.get("versions"), dict) and info["versions"],
                "backend compiler and exact versions/pins must be declared")
        require(all(isinstance(info["versions"].get(key), str)
                    and re.fullmatch(r"[a-f0-9]{40}", info["versions"][key])
                    for key in ("rslint", "compiler")),
                "versions.rslint and versions.compiler must be exact 40-digit commit pins")
        conditions = info.get("build_conditions", {})
        require(conditions.get("optimization") == "release" and "threads" in conditions and "cache" in conditions,
                "declare release optimization, threading and cache conditions")
    verify_artifacts(manifest, base)
    fixtures, cases = manifest.get("fixtures"), manifest.get("cases")
    require(isinstance(fixtures, list) and fixtures and isinstance(cases, list) and cases,
            "fixture correctness cases and representative real-project cases are both required")
    ids = set()
    loaded = []
    for kind, group in (("fixture", fixtures), ("case", cases)):
        for case in group:
            name = case.get("id")
            require(isinstance(name, str) and re.fullmatch(r"[a-zA-Z0-9][a-zA-Z0-9_-]*", name) and name not in ids,
                    "case IDs must be unique simple path-safe names")
            ids.add(name)
            work_path = relative_path(base, case.get("work_manifest"), "work manifest")
            work = load_work(work_path)
            verify_work(work)
            policy = case.get("normalization")
            require(isinstance(policy, dict) and policy.get("format") in ("json", "jsonl"), "case needs an explicit normalization policy")
            require(not (set(policy) - {"format", "path_fields", "path_roots", "unordered_arrays", "required_json_pointers"}),
                    "normalization contains unsupported field-dropping rules")
            for field in ("path_fields", "unordered_arrays", "required_json_pointers"):
                require(isinstance(policy.get(field), list), "normalization must explicitly declare " + field)
                for pointer in policy[field]:
                    parts = parse_pointer(pointer)
                    if field == "unordered_arrays":
                        require(not any(part.lower() in ("fix", "fixes", "edit", "edits", "suggestions") for part in parts),
                                "fix/edit/suggestion order must not be normalized away")
            require(policy["required_json_pointers"], "required structured output pointers cannot be empty")
            variables = {"manifest_dir": base, "output_dir": output, "work_manifest": work_path,
                         "observation": output / name / "observation.json", "case_id": name}
            for backend in BACKENDS:
                require(backend in case, "case requires both backend commands")
                prepare_command(case[backend], base, variables)
                expectation = case[backend].get("process_expectations")
                require(isinstance(expectation, dict) and type(expectation.get("minimum_peak_count")) is int
                        and expectation["minimum_peak_count"] >= 1
                        and isinstance(expectation.get("required_executables"), list)
                        and expectation["required_executables"], "declare expected process coverage for both backends")
            proof = case.get("proof_command")
            require(isinstance(proof, dict) and any("{observation}" in arg for arg in proof.get("argv", [])),
                    "a proof command consuming {observation} is required")
            prepare_command(proof, base, variables)
            require(proof["expected_exit_codes"] == [0], "correctness proof command must require exit zero")
            loaded.append((kind, case, work))
    require(any(rule["requires_type_information"] for kind, _, work in loaded if kind == "case"
                for file in work["data"]["rules"] for rule in file["rules"]),
            "real-project cases must include type-aware work")
    return manifest, loaded


def equivalence(kind, case, work, base, output, interval, lock_fd):
    case_dir = output / case["id"]
    case_dir.mkdir()
    variables = {"manifest_dir": base, "output_dir": output, "work_manifest": work["path"],
                 "observation": case_dir / "observation.json", "case_id": case["id"]}
    observed, records = {}, []
    for backend in BACKENDS:
        verify_work(work)
        prefix = case_dir / ("equivalence-" + backend)
        record = run_once(case[backend], base, variables, prefix, interval, lock_fd)
        verify_work(work)
        record.update(case=case["id"], backend=backend, phase="equivalence")
        records.append(record)
        observed[backend] = observed_output(prefix, case["normalization"], backend, base, record)
        observed[backend]["work_fingerprint"] = work["fingerprint"]
    require(observed["baseline"] == observed["candidate"], "full structured diagnostic/fix/exit equivalence failed: " + case["id"])
    observation = {"schema_version": 1, "case": case["id"], "kind": kind,
                   "expected_work_sha256": work["sha256"], "backends": observed,
                   "stdout_paths": {backend: str(case_dir / ("equivalence-" + backend + ".stdout")) for backend in BACKENDS},
                   "normalization": case["normalization"]}
    save_json(variables["observation"], observation)
    observation_hash = file_digest(variables["observation"])
    prefix = case_dir / "proof"
    record = run_once(case["proof_command"], base, variables, prefix, interval, lock_fd, require_rss=False)
    require(record["exit_code"] == 0, "proof failed")
    require(Path(str(prefix) + ".stdout").stat().st_size <= (256 << 20), "proof output exceeds 256 MiB bound")
    proof = read_json(str(prefix) + ".stdout")
    require(proof.get("schema_version") == 1 and proof.get("status") == "passed" and proof.get("scope") == "complete-lint", "proof did not establish complete-lint equivalence")
    require(proof.get("observation_sha256") == observation_hash and proof.get("expected_work_sha256") == work["sha256"], "proof is not bound to this observation/work manifest")
    for backend in BACKENDS:
        answer = proof.get("backends", {}).get(backend, {})
        for field in ("output_sha256", "exit_code", "work_fingerprint"):
            require(answer.get(field) == observed[backend][field], "proof workload/output mismatch: " + field)
    require(all(proof.get("checks", {}).get(name) is True for name in PROOF_CHECKS), "proof is missing a required functional/workload check")
    if kind == "fixture":
        require(all(proof.get("fixture_coverage", {}).get(name) is True for name in FIXTURE_CHECKS), "fixture proof requires hit/non-hit and true type-aware coverage")
    verify_work(work)
    save_json(case_dir / "equivalence-evidence.json", {"observation_sha256": observation_hash,
              "proof_sha256": file_digest(str(prefix) + ".stdout"), "work_sha256": work["sha256"], "status": "passed"})
    return observed["baseline"], records, variables


def stats(values):
    mid = statistics.median(values)
    return {"count": len(values), "median": mid, "mad": statistics.median(abs(value - mid) for value in values),
            "min": min(values), "max": max(values)}


def summarize(records):
    summary = {}
    for case in sorted({row["case"] for row in records if row["phase"] == "measurement"}):
        rows = [row for row in records if row["case"] == case and row["phase"] == "measurement"]
        by_backend = {backend: [row for row in rows if row["backend"] == backend] for backend in BACKENDS}
        result = {"backends": {backend: {} for backend in BACKENDS}, "comparisons": {}}
        for metric in ("wall_seconds", "sampled_tree_peak_rss_bytes"):
            for backend in BACKENDS:
                result["backends"][backend][metric] = stats([row[metric] for row in by_backend[backend]])
            baseline, candidate = [result["backends"][backend][metric] for backend in BACKENDS]
            absolute = candidate["median"] - baseline["median"]
            noise = 3 * (baseline["mad"] + candidate["mad"])
            if metric == "wall_seconds":
                noise = max(noise, 2 * max(row.get("wall_observation_uncertainty_seconds", 0) for row in rows))
            result["comparisons"][metric] = {
                "candidate_minus_baseline": absolute,
                "candidate_change_percent": 100 * absolute / baseline["median"],
                "reduction_percent": -100 * absolute / baseline["median"],
                "candidate_over_baseline": candidate["median"] / baseline["median"],
                "baseline_over_candidate": baseline["median"] / candidate["median"],
                "heuristic_noise_threshold": noise,
                "directional_signal": ("inconclusive" if abs(absolute) <= noise else "lower" if absolute < 0 else "higher"),
                "signal_method": "median difference greater than 3*(baseline MAD+candidate MAD); heuristic, not a confidence interval",
                "paired_candidate_over_baseline": stats([candidate_row[metric] / baseline_row[metric]
                    for baseline_row, candidate_row in zip(by_backend["baseline"], by_backend["candidate"])])}
        summary[case] = result
    return summary


def lock_descriptor(lock_path):
    target = lock_path.stat()
    fd_dir = Path("/dev/fd") if sys.platform == "darwin" else Path("/proc/self/fd")
    inherited = []
    for entry in fd_dir.iterdir():
        if not entry.name.isdigit():
            continue
        fd = int(entry.name)
        try:
            stat = os.fstat(fd)
            if (stat.st_dev, stat.st_ino) == (target.st_dev, target.st_ino):
                inherited.append(fd)
        except OSError:
            pass
    require(inherited, "locked worker requires the wrapper's inherited lock descriptor")
    with lock_path.open("a+") as independent:
        try:
            fcntl.flock(independent, fcntl.LOCK_EX | fcntl.LOCK_NB)
        except BlockingIOError:
            pass
        else:
            fcntl.flock(independent, fcntl.LOCK_UN)
            raise BenchmarkError("machine lock is not held by the wrapper")
    for fd in inherited:
        try:
            fcntl.flock(fd, fcntl.LOCK_EX | fcntl.LOCK_NB)
            return fd
        except BlockingIOError:
            pass
    raise BenchmarkError("inherited descriptor does not own the machine lock")


def batch(manifest_path, output, lock_path):
    lock_fd = lock_descriptor(lock_path)
    manifest, loaded = validate_manifest(manifest_path, output)
    manifest_hash = file_digest(manifest_path)
    save_json(output / "manifest.json", manifest)
    save_json(output / "environment.json", {"platform": platform.uname()._asdict(), "python": sys.version,
              "cpu_count": os.cpu_count(), "load_before": list(os.getloadavg()),
              "started_unix": time.time(), "manifest_sha256": manifest_hash,
              "memory_method": MEMORY_METHOD, "lock": str(lock_path),
              "note": "raw logs may contain local paths; publish reviewed, portable summaries separately"})
    ps = subprocess.run(["ps", "-Ao", "pid,ppid,pgid,pcpu,pmem,etime,comm"], stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    (output / "process-load-before.txt").write_bytes(ps.stdout + ps.stderr)
    interval = manifest.get("sample_interval_seconds", 0.002)
    records, comparisons = [], {}
    prepared = []
    try:
        # Every correctness gate completes before any warmup or timing phase.
        for kind, case, work in loaded:
            expected, preflight, variables = equivalence(kind, case, work, manifest_path.parent, output, interval, lock_fd)
            records.extend(preflight)
            prepared.append((kind, case, work, expected, variables))
            save_json(output / "records.json", records)
            print("equivalence passed: " + case["id"], flush=True)
        for case_index, (kind, case, work, expected, variables) in enumerate(prepared):
            if kind == "fixture":
                continue
            for phase, repetitions in (("warmup", manifest.get("warmups", 2)), ("measurement", manifest.get("repeats", 5))):
                for repetition in range(repetitions):
                    order = BACKENDS if (case_index + repetition) % 2 == 0 else tuple(reversed(BACKENDS))
                    for backend in order:
                        require(file_digest(manifest_path) == manifest_hash, "benchmark manifest changed during batch")
                        verify_artifacts(manifest, manifest_path.parent)
                        verify_work(work)
                        prefix = output / case["id"] / ("{}-{:02d}-{}".format(phase, repetition + 1, backend))
                        record = run_once(case[backend], manifest_path.parent, variables, prefix, interval, lock_fd)
                        verify_work(work)
                        actual = observed_output(prefix, case["normalization"], backend, manifest_path.parent, record)
                        actual["work_fingerprint"] = work["fingerprint"]
                        require(actual == expected, "measured invocation changed diagnostics, fixes, exit or work: " + str(prefix))
                        record.update(case=case["id"], backend=backend, phase=phase, repetition=repetition + 1)
                        records.append(record)
                        save_json(output / "records.json", records)
                        print("{} {} {} {}: {:.6f}s {:.2f} MiB".format(case["id"], phase, repetition + 1,
                              backend, record["wall_seconds"], record["sampled_tree_peak_rss_bytes"] / 1048576), flush=True)
        comparisons = summarize(records)
        save_json(output / "summary.json", {"status": "measured", "scope": "complete-lint", "cases": comparisons,
                  "memory_method": MEMORY_METHOD, "load_after": list(os.getloadavg()),
                  "note": "all values are specific to these verified inputs/builds; no PoC values are included"})
        lines = ["case,metric,baseline_median,candidate_median,absolute_change,change_percent,candidate_over_baseline,signal"]
        for case, result in comparisons.items():
            for metric, compare in result["comparisons"].items():
                lines.append(",".join(map(str, [case, metric, result["backends"]["baseline"][metric]["median"],
                    result["backends"]["candidate"][metric]["median"], compare["candidate_minus_baseline"],
                    compare["candidate_change_percent"], compare["candidate_over_baseline"], compare["directional_signal"]])))
        (output / "summary.csv").write_text("\n".join(lines) + "\n")
    except BaseException as error:
        save_json(output / "failure.json", {"status": "failed", "error": str(error), "completed_records": len(records),
                  "no_performance_conclusion": True})
        raise


def example():
    command = {"argv": ["node", "../baseline/packages/rslint/bin/rslint.js", ".", "--format", "jsonline"],
               "cwd": "project", "env": {}, "inherit_env": ["PATH", "HOME", "TMPDIR"],
               "expected_exit_codes": [0, 1], "timeout_seconds": 120,
               "process_expectations": {"minimum_peak_count": 2, "required_executables": ["node", "rslint"]}}
    case = {"id": "project", "work_manifest": "work.json", "baseline": command,
            "candidate": dict(command, argv=["node", "../candidate/packages/rslint/bin/rslint.js", ".", "--format", "jsonline"],
                              process_expectations={"minimum_peak_count": 3, "required_executables": ["node", "rslint", "rslint-tsrs-service"]}),
            "normalization": {"format": "jsonl", "path_fields": ["/*/file"], "path_roots": {"baseline": {"project": "project"}, "candidate": {"project": "project"}},
                              "unordered_arrays": [""], "required_json_pointers": [""]},
            "proof_command": {"argv": ["python3", "verify-full-workload.py", "{observation}", "{work_manifest}"],
                              "cwd": ".", "env": {}, "inherit_env": ["PATH"], "expected_exit_codes": [0], "timeout_seconds": 120}}
    return {"schema_version": 1, "scope": "complete-lint", "warmups": 2, "repeats": 5,
            "sample_interval_seconds": 0.002,
            "backends": {backend: {"compiler": compiler, "versions": {"rslint": "EXACT_COMMIT", "compiler": "EXACT_COMMIT"},
                                   "build_conditions": {"optimization": "release", "threads": "EXPLICIT_THREAD_SETTING", "cache": "new process, warm OS cache"},
                                   "artifacts": [{"path": backend + "/rslint", "sha256": "REPLACE_WITH_ACTUAL_SHA256"}]}
                         for backend, compiler in (("baseline", "tsgo"), ("candidate", "tsrs"))},
            "fixtures": [dict(case, id="fixture", work_manifest="fixture-work.json")], "cases": [case],
            "_example_note": "Replace placeholders and implement the real workload/fix proof; this template is deliberately not runnable."}



def self_test(output, lock_path):
    """Short owned children and protocol invariants, never a performance result."""
    lock_fd = lock_descriptor(lock_path)
    checks = []
    require(clock_observation(1.0, 1.002)["consistent"], "small clock read skew was rejected")
    require(not clock_observation(1.0, 61.0)["consistent"], "suspend-like forward divergence was accepted")
    require(not clock_observation(1.0, -2.0)["consistent"], "backward realtime clock step was accepted")
    checks.append("clock divergence rejects suspended or stepped observations")
    class ReusedPIDProbe:
        def __init__(self):
            self.visited = []
            self.rows = {
                1: {"pid": 1, "ppid": 0, "pgid": 1, "identity": "1:new", "rss_bytes": 999, "comm": "unrelated"},
                2: {"pid": 2, "ppid": 0, "pgid": 2, "identity": "2:owned", "rss_bytes": 12, "comm": "owned"},
                3: {"pid": 3, "ppid": 2, "pgid": 2, "identity": "3:child", "rss_bytes": 18, "comm": "child"},
                99: {"pid": 99, "ppid": 1, "pgid": 1, "identity": "99:unrelated", "rss_bytes": 999, "comm": "unrelated"}}
        def refresh(self):
            pass
        def info(self, pid):
            return self.rows.get(pid)
        def group_pids(self, root):
            return [1, 99]
        def children(self, pid):
            self.visited.append(pid)
            return [3] if pid == 2 else [99] if pid == 1 else []
        def executable(self, pid):
            return self.rows[pid]["comm"]
    tree = ProcessTree.__new__(ProcessTree)
    tree.root, tree.probe = 1, ReusedPIDProbe()
    tree.known, tree.executables, tree.reparented, tree.reused = {1: "1:old", 2: "2:owned"}, {}, set(), set()
    rows = tree.sample()
    require({row["pid"] for row in rows} == {2, 3} and 1 not in tree.probe.visited,
            "PID reuse allowed traversal of unrelated children")
    checks.append("PID reuse excludes unrelated group and descendants")

    policy = {"format": "json", "path_fields": ["/diagnostics/*/file"],
              "path_roots": {"baseline": {"project": "baseline"}, "candidate": {"project": "candidate"}},
              "unordered_arrays": ["/diagnostics"], "required_json_pointers": ["/diagnostics", "/fixes"]}
    def result(backend, message="message", end=2, replacement="fixed"):
        return canonical({"diagnostics": [{"file": str(output / backend / "source.ts"), "rule": "test",
                          "message": message, "start": 1, "end": end}], "fixes": [{"text": replacement}]})
    baseline = normalize_output(result("baseline"), policy, "baseline", output)
    require(baseline == normalize_output(result("candidate"), policy, "candidate", output), "declared path normalization failed")
    for kwargs in ({"message": "changed"}, {"end": 3}, {"replacement": "changed"}):
        require(baseline != normalize_output(result("candidate", **kwargs), policy, "candidate", output), "normalization discarded diagnostic/fix information")
    try:
        normalize_output(result("baseline"), dict(policy, drop_fields=["message"]), "baseline", output)
    except BenchmarkError:
        pass
    else:
        raise BenchmarkError("field-dropping normalization was accepted")
    checks.append("normalization preserves messages, positions and fixes")

    # A detached grandchild outlives both its original parent and the root.
    # Its memory is deliberately small and this is not a benchmark workload.
    leaf = "import time; data=bytearray(16*1024*1024); time.sleep(0.65)"
    middle = "import subprocess,sys,time; subprocess.Popen([sys.executable,'-c',{}],start_new_session=True); time.sleep(0.25)".format(repr(leaf))
    root = "import subprocess,sys,time; subprocess.Popen([sys.executable,'-c',{}]); time.sleep(0.40)".format(repr(middle))
    command = {"argv": [sys.executable, "-c", root], "cwd": ".", "env": {},
               "inherit_env": ["PATH", "HOME", "TMPDIR"], "expected_exit_codes": [0], "timeout_seconds": 10,
               "process_expectations": {"minimum_peak_count": 3, "required_executables": []}}
    record = run_once(command, output, {}, output / "owned-tree", 0.003, lock_fd)
    require(record["observed_children_after_parent_exit"] and record["reparented_processes"], "detached reparented child was not retained")
    require(record["sampled_tree_peak_rss_bytes"] >= 16 * 1024 * 1024, "child allocation was omitted from tree RSS")
    require(record["wall_seconds"] > record["root_wall_seconds"], "child lifecycle was omitted from total wall time")
    checks.append("native sampler retains detached reparented grandchild through its exit")
    save_json(output / "self-test.json", {"status": "passed", "checks": checks,
              "scope": "harness correctness only", "no_performance_conclusion": True})
    print("Harness self-test passed; no rslint performance measurement was performed.", flush=True)


def main():
    parser = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    parser.add_argument("--manifest", type=Path)
    parser.add_argument("--output", type=Path)
    parser.add_argument("--lock", type=Path)
    parser.add_argument("--lock-timeout", type=float, default=1800)
    parser.add_argument("--batch-timeout", type=float, default=21600)
    parser.add_argument("--run", action="store_true", help="execute after validation under the shared machine lock")
    parser.add_argument("--example", action="store_true", help="print a deliberately incomplete portable manifest template")
    parser.add_argument("--self-test", action="store_true", help="test sampler/protocol with short owned children under a temporary lock")
    parser.add_argument("--_self-test-locked", action="store_true", help=argparse.SUPPRESS)
    parser.add_argument("--_locked", action="store_true", help=argparse.SUPPRESS)
    args = parser.parse_args()
    if args.example:
        print(json.dumps(example(), indent=2))
        return 0
    if args._self_test_locked:
        require(args.output is not None and args.lock is not None, "self-test worker needs output and lock")
        self_test(args.output.resolve(), args.lock.resolve())
        return 0
    if args.self_test:
        temporary = tempfile.TemporaryDirectory(prefix="rslint-ab-self-test-") if args.output is None else None
        output = Path(temporary.name) if temporary else args.output.resolve()
        require(not output.exists() or not any(output.iterdir()), "self-test output directory must be empty")
        output.mkdir(parents=True, exist_ok=True)
        wrapper = Path(__file__).with_name("with-machine-lock.py")
        test_lock = args.lock.resolve() if args.lock else output / "self-test.lock"
        try:
            return subprocess.run([sys.executable, str(wrapper), "--lock", str(test_lock),
                "--lock-timeout", str(args.lock_timeout), "--timeout", "30", "--cwd", str(output), "--log", str(output / "self-test.log"),
                "--", sys.executable, str(Path(__file__).resolve()), "--output", str(output),
                "--lock", str(test_lock), "--_self-test-locked"]).returncode
        finally:
            if temporary:
                temporary.cleanup()
    if args.manifest is None:
        parser.error("--manifest is required")
    manifest = args.manifest.resolve()
    output = args.output.resolve() if args.output else manifest.parent / "benchmark-results"
    if args._locked:
        require(args.lock is not None, "locked worker needs --lock")
        batch(manifest, output, args.lock.resolve())
        return 0
    validate_manifest(manifest, output)
    if not args.run:
        print("Manifest and input hashes validated; no commands executed. Pass --run to measure.")
        return 0
    if args.lock is None:
        parser.error("--run requires the shared --lock path")
    require(not output.exists() or not any(output.iterdir()), "output directory must be new or empty; preserve previous raw results")
    output.mkdir(parents=True, exist_ok=True)
    wrapper = Path(__file__).with_name("with-machine-lock.py")
    result = subprocess.run([sys.executable, str(wrapper), "--lock", str(args.lock.resolve()),
        "--lock-timeout", str(args.lock_timeout), "--timeout", str(args.batch_timeout),
        "--cwd", str(manifest.parent), "--log", str(output / "batch.log"), "--",
        sys.executable, str(Path(__file__).resolve()), "--manifest", str(manifest), "--output", str(output),
        "--lock", str(args.lock.resolve()), "--_locked"])
    return result.returncode


if __name__ == "__main__":
    try:
        sys.exit(main())
    except (BenchmarkError, OSError, ValueError) as error:
        print("benchmark: " + str(error), file=sys.stderr)
        sys.exit(1)

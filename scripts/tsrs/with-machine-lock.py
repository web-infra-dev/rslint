#!/usr/bin/env python3
"""Run a build or measurement while holding an advisory machine lock."""

import argparse
import fcntl
import json
import os
from pathlib import Path
import signal
import selectors
import subprocess
import sys
import time


def acquire(handle, timeout):
    # A separate process can block fairly in the kernel's flock queue without
    # making Python's main thread uninterruptible. A deadline kills only this
    # acquisition helper; the helper never launches the requested command.
    waiter = subprocess.Popen(
        [sys.executable, "-c",
         "import fcntl,sys; fcntl.flock(int(sys.argv[1]),fcntl.LOCK_EX); "
         "print('locked',flush=True); sys.stdin.buffer.read()",
         str(handle.fileno())],
        stdin=subprocess.PIPE, stdout=subprocess.PIPE,
        pass_fds=(handle.fileno(),),
    )
    started = time.monotonic()
    last_notice = started
    try:
        with selectors.DefaultSelector() as selector:
            selector.register(waiter.stdout, selectors.EVENT_READ)
            while True:
                elapsed = time.monotonic() - started
                if elapsed >= timeout:
                    raise TimeoutError(f"machine lock wait exceeded {timeout}s")
                if selector.select(min(1, timeout - elapsed)):
                    if waiter.stdout.readline() != b"locked\n":
                        raise RuntimeError("machine lock acquisition helper failed")
                    return time.monotonic() - started, waiter
                if time.monotonic() - last_notice >= 30:
                    print(f"waiting for machine lock ({elapsed:.0f}s)", flush=True)
                    last_notice = time.monotonic()
    except BaseException:
        waiter.kill()
        waiter.wait()
        waiter.stdin.close()
        waiter.stdout.close()
        raise


def stop_child(child):
    if child.poll() is None:
        os.killpg(child.pid, signal.SIGTERM)
        try:
            child.wait(timeout=10)
        except subprocess.TimeoutExpired:
            os.killpg(child.pid, signal.SIGKILL)
            child.wait()


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--lock", default=os.environ.get("RSLINT_MACHINE_LOCK"))
    parser.add_argument("--lock-timeout", type=float, default=1800)
    parser.add_argument("--timeout", type=float, default=1800)
    parser.add_argument("--cwd", default=".")
    parser.add_argument("--log", required=True)
    parser.add_argument("command", nargs=argparse.REMAINDER)
    args = parser.parse_args()
    if not args.lock:
        parser.error("provide --lock or RSLINT_MACHINE_LOCK")
    command = args.command[1:] if args.command[:1] == ["--"] else args.command
    if not command:
        parser.error("provide a command after --")
    if args.lock_timeout <= 0 or args.timeout <= 0:
        parser.error("timeouts must be positive")
    log = Path(args.log)
    log.parent.mkdir(parents=True, exist_ok=True)
    record = {"command": command, "cwd": args.cwd, "started_unix": time.time()}
    code = 1
    lock_waiter = None
    try:
        with open(args.lock, "a+") as lock:
            record["lock_wait_seconds"], lock_waiter = acquire(lock, args.lock_timeout)
            print(f"machine lock acquired ({record['lock_wait_seconds']:.2f}s)", flush=True)
            record["load_before"] = list(os.getloadavg())
            started = time.monotonic()
            with log.open("w") as output:
                child = subprocess.Popen(
                    command, cwd=args.cwd, stdout=output, stderr=subprocess.STDOUT,
                    start_new_session=True, pass_fds=(lock.fileno(),),
                )
                try:
                    code = child.wait(timeout=args.timeout)
                except subprocess.TimeoutExpired:
                    record["error"] = "child timeout"
                    code = 124
                finally:
                    stop_child(child)
            record["elapsed_seconds"] = time.monotonic() - started
            record["load_after"] = list(os.getloadavg())
    except TimeoutError as error:
        record["error"] = str(error)
        code = 124
    finally:
        if lock_waiter is not None:
            lock_waiter.stdin.close()
            lock_waiter.wait(timeout=10)
            lock_waiter.stdout.close()
        record["exit_code"] = code
        Path(str(log) + ".meta.json").write_text(json.dumps(record, indent=2) + "\n")
    print(f"exit={code}; log={log}", flush=True)
    return code


if __name__ == "__main__":
    sys.exit(main())

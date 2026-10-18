#!/usr/bin/env python3
"""Run existing Go tests with the repository's short-test time limits."""
import argparse
import collections
import json
import os
from pathlib import Path
import queue
import shutil
import signal
import subprocess
import sys
import threading
import time


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--package", default="./tlc")
    parser.add_argument("--run", help="Go test name expression")
    parser.add_argument("--log", default=".cache/test-results.jsonl")
    args = parser.parse_args()
    root = Path(__file__).resolve().parent.parent
    os.chdir(root)
    log_path = Path(args.log)
    log_path.parent.mkdir(parents=True, exist_ok=True)
    cmd = [os.environ.get("GO", "go"), "test", args.package,
           "-p=1", "-count=1", "-timeout=2m", "-json"]
    if args.run:
        cmd += ["-run", args.run]
    if sys.platform.startswith("linux") and shutil.which("taskset"):
        cmd = ["taskset", "-c", "0-7"] + cmd
    env = dict(os.environ, GOMAXPROCS="8")
    proc = subprocess.Popen(cmd, stdout=subprocess.PIPE, stderr=subprocess.STDOUT,
                            env=env, start_new_session=(os.name != "nt"))
    lines = queue.Queue()

    def read_output():
        for line in proc.stdout:
            lines.put(line.decode("utf-8", errors="replace"))
        lines.put(None)

    threading.Thread(target=read_output, daemon=True).start()
    active = {}
    counts = collections.Counter()
    tail = collections.deque(maxlen=30)
    started = None
    elapsed = 0
    reason = None
    next_progress = time.monotonic() + 10

    def stop():
        if proc.poll() is not None:
            return
        if os.name == "nt":
            subprocess.run(["taskkill", "/F", "/T", "/PID", str(proc.pid)],
                           stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
        else:
            os.killpg(proc.pid, signal.SIGKILL)

    try:
        with log_path.open("w", encoding="utf-8") as log:
            while True:
                try:
                    line = lines.get(timeout=0.05)
                except queue.Empty:
                    line = ""
                if line is None:
                    break
                if line:
                    log.write(line)
                    tail.append(line)
                    try:
                        event = json.loads(line)
                    except ValueError:
                        event = {}
                    action, name = event.get("Action"), event.get("Test")
                    if action in ("run", "cont") and name:
                        active[name] = time.monotonic()
                        started = started or time.monotonic()
                    elif action == "pause" and name:
                        active.pop(name, None)
                    elif action in ("pass", "fail", "skip"):
                        if name:
                            active.pop(name, None)
                            counts[action] += 1
                        else:
                            elapsed = event.get("Elapsed", elapsed)
                now = time.monotonic()
                slow = [name for name, began in active.items()
                        if now - began >= 9.5
                        and not any(child.startswith(name + "/") for child in active)]
                if reason is None and slow:
                    reason = "Test reached the 9.5-second safety cutoff: " + ", ".join(slow)
                    stop()
                if reason is None and started and now - started >= 118:
                    reason = "Suite reached the 118-second safety cutoff."
                    stop()
                if now >= next_progress:
                    print(f"Waiting: {counts['pass']} passed, {counts['skip']} skipped...", flush=True)
                    next_progress = now + 10
    except BaseException:
        stop()
        proc.wait()
        raise
    finally:
        proc.stdout.close()
    code = proc.wait()
    if not code and not reason and not sum(counts.values()):
        reason = "No tests matched; check the package and --run expression."
    if code or reason:
        print("".join(tail), file=sys.stderr, end="")
        print(reason or "Go tests failed.", file=sys.stderr)
    print(f"{'FAIL' if code or reason else 'PASS'}: {counts['pass']} passed, "
          f"{counts['skip']} skipped, {counts['fail']} failed; {elapsed:.3f}s. "
          f"Log: {log_path}")
    return 1 if code or reason else 0


if __name__ == "__main__":
    sys.exit(main())

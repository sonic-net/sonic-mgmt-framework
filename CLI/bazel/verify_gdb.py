"""Prove each installed ELF discovers its separate symbols, with a negative control.

GDB is an explicit CI inspection tool, not a build input. Runtime and symbol
TARs are extracted into separate temporary roots: no adjacent source binary,
host symbol directory or user GDB configuration can satisfy the lookup.
"""

import argparse
import hashlib
import json
from pathlib import Path
import re
import subprocess
import tarfile
import tempfile


CLI_SYMBOLS = {
    "usr/sbin/cli/clish": "main",
    "usr/sbin/cli/.libs/clish_plugin_clish.so": "clish_plugin_clish_init",
    "usr/sbin/cli/.libs/libclish.so.1.0.0": "clish_shell_new",
    "usr/sbin/cli/.libs/libkonf.so.1.0.0": "konf_tree_new",
    "usr/sbin/cli/.libs/liblub.so.1.0.0": "lub_string_dup",
    "usr/sbin/cli/.libs/libtinyrl.so.1.0.0": "tinyrl_new",
}
REST_SYMBOLS = {
    "usr/sbin/rest_server": "main.main",
    "usr/sbin/generate_cert": "main.main",
}


def verify(runtime, debug, gdb, cli_only=False):
    runtime, debug, gdb = (Path(value).resolve() for value in (runtime, debug, gdb))
    symbols = CLI_SYMBOLS if cli_only else CLI_SYMBOLS | REST_SYMBOLS
    with tempfile.TemporaryDirectory(prefix="framework-debug-lookup-") as directory:
        root = Path(directory)
        for archive_path, destination in ((runtime, root / "runtime"), (debug, root / "symbols")):
            with tarfile.open(archive_path) as archive:
                archive.extractall(destination, filter="data")
                if archive_path == debug:
                    for member in archive.getmembers():
                        if member.isfile():
                            if member.mode != 0o644:
                                raise AssertionError(f"Unexpected debug file mode {oct(member.mode)} for {member.name}")
                            if (member.uid, member.gid) != (0, 0):
                                raise AssertionError(f"Unexpected debug file ownership {member.uid}:{member.gid} for {member.name}")
        empty = root / "empty"
        empty.mkdir()
        environment = {"PATH": "/usr/bin:/bin", "HOME": str(empty),
                       "LANG": "C.UTF-8", "LC_ALL": "C.UTF-8", "DEBUGINFOD_URLS": ""}
        records = []
        for relative, symbol in symbols.items():
            binary = root / "runtime" / relative
            if not binary.is_file():
                raise AssertionError("Missing installed executable/library: " + relative)
            results = []
            for debug_root in (root / "symbols/usr/lib/debug", empty):
                result = subprocess.run(
                    [str(gdb), "-nx", "-nh", "--batch",
                     "-iex", "set auto-load off",
                     "-iex", "set debuginfod enabled off",
                     "-iex", "set debug-file-directory " + str(debug_root),
                     "-ex", "file " + str(binary),
                     "-ex", "info line " + symbol,
                     "-ex", "ptype " + symbol],
                    cwd=empty, env=environment, capture_output=True, text=True, timeout=30,
                )
                results.append({"returncode": result.returncode,
                                "output": result.stdout + result.stderr})
            positive, negative = results
            # Missing symbol names can make GDB return 1. A valid negative
            # control permits that diagnostic but must never find a line.
            line = r"Line [0-9]+ of"
            if positive["returncode"] != 0 or not re.search(line, positive["output"]):
                raise AssertionError({"path": relative, "positive": positive})
            # Pointer-to-struct return types expand the struct over several
            # lines before GDB prints the function's argument list.
            signature = re.search(r"type = ([\s\S]+)", positive["output"])
            if not signature or "(" not in signature[1] or "<unknown" in signature[1]:
                raise AssertionError({"path": relative, "positive_type": positive})
            if negative["returncode"] not in (0, 1) or re.search(line, negative["output"]):
                raise AssertionError({"path": relative, "negative": negative})
            records.append({"path": relative, "symbol": symbol,
                            "positive": positive, "negative": negative})
        return {"runtime_sha256": hashlib.sha256(runtime.read_bytes()).hexdigest(),
                "debug_sha256": hashlib.sha256(debug.read_bytes()).hexdigest(),
                "matching_symbol_lookups": len(records), "checks": records}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--runtime", required=True)
    parser.add_argument("--debug", required=True)
    parser.add_argument("--gdb", required=True, help="Absolute path to the CI inspection GDB")
    parser.add_argument("--cli-only", action="store_true")
    parser.add_argument("--output", type=Path)
    args = parser.parse_args()
    report = json.dumps(verify(args.runtime, args.debug, args.gdb, args.cli_only), indent=2, sort_keys=True) + "\n"
    if args.output:
        args.output.parent.mkdir(parents=True, exist_ok=True)
        args.output.write_text(report)
    print(report, end="")


if __name__ == "__main__":
    main()

"""Execute installed clish with its declared loader, libraries and command tree."""

import argparse
import hashlib
import json
import os
from pathlib import Path
import subprocess
import tarfile
import tempfile


def probe(runtime, libraries):
    runtime, libraries = Path(runtime).resolve(), Path(libraries).resolve()
    with tempfile.TemporaryDirectory(prefix="framework-cli-installed-") as directory:
        root = Path(directory)
        with tarfile.open(runtime) as archive:
            archive.extractall(root, filter="data")
        cli = root / "usr/sbin/cli"
        native = libraries / "usr/lib/x86_64-linux-gnu"
        loader = native / "ld-linux-x86-64.so.2"
        home = root / "home"
        home.mkdir()
        environment = {
            "PATH": "/usr/bin:/bin",
            "HOME": str(home),
            "LANG": "C.UTF-8",
            "LC_ALL": "C.UTF-8",
            "CLISH_PATH": str(cli / "command-tree"),
            "SONIC_CLI_ROOT": str(cli),
            "SONIC_MGMT_ROOT": str(cli.parent),
            "SYSTEM_NAME": "sonic",
            "LD_LIBRARY_PATH": os.pathsep.join((str(cli / ".libs"), str(native))),
            "PYTHONHOME": str(libraries / "usr"),
            "PYTHONPATH": os.pathsep.join((str(cli), str(cli / "scripts"))),
            "PYTHONNOUSERSITE": "1",
        }
        checks = []
        for arguments, expected_code, expected_text in (
            (["-v"], 0, "2.1.4"),
            (["-k", "-c", "show tech-support"], 0, None),
            (["-k", "-c", "not-a-real-sonic-command"], 255, "Syntax error"),
        ):
            result = subprocess.run(
                [str(loader), "--library-path", environment["LD_LIBRARY_PATH"], str(cli / "clish"), *arguments],
                cwd=home,
                env=environment,
                capture_output=True,
                text=True,
                timeout=30,
            )
            record = {"args": arguments, "returncode": result.returncode,
                      "stdout": result.stdout, "stderr": result.stderr}
            checks.append(record)
            if result.returncode != expected_code:
                raise AssertionError(record)
            if expected_text is not None and expected_text not in result.stdout + result.stderr:
                raise AssertionError(record)
        return {"runtime_sha256": hashlib.sha256(runtime.read_bytes()).hexdigest(),
                "checks": checks}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("runtime")
    parser.add_argument("libraries")
    args = parser.parse_args()
    print(json.dumps(probe(args.runtime, args.libraries), indent=2, sort_keys=True))


if __name__ == "__main__":
    main()

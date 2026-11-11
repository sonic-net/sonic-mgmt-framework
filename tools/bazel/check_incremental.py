"""Measure no-op and independent REST/UI/CLI edits against the public TARs."""

import hashlib
import json
import os
from pathlib import Path
import shlex
import subprocess
import time

BAZEL = shlex.split(os.environ.get("BAZEL_COMMAND", "bazel"))
OUTPUT = Path("bazel-evidence")
TARGETS = ["//dist:runtime", "//dist:runtime.debug_symbols", "//dist:rest_binaries", "//dist:ui",
           "//CLI:runtime", "//CLI:runtime.debug_symbols"]


def snapshot(name):
    start = time.monotonic()
    with (OUTPUT / (name + ".log")).open("w") as log:
        subprocess.run([*BAZEL, "build", *TARGETS], stdout=log, stderr=subprocess.STDOUT, check=True)
    seconds = time.monotonic() - start
    hashes = {}
    for target in TARGETS:
        path = subprocess.check_output([*BAZEL, "cquery", target, "--output=files"], text=True).strip()
        artifact = Path(path)
        assert artifact.is_file(), (target, path)
        hashes[target] = hashlib.sha256(artifact.read_bytes()).hexdigest()
    return {"seconds": seconds, "sha256": hashes}


def main():
    OUTPUT.mkdir(exist_ok=True)
    result = {"baseline": snapshot("baseline"), "noop": snapshot("noop")}
    assert result["baseline"]["sha256"] == result["noop"]["sha256"]
    edits = [
        ("rest", Path("rest/server/restconf.go"), b"2016-06-21", b"2016-06-22", {"//dist:runtime", "//dist:runtime.debug_symbols", "//dist:rest_binaries"}),
        ("ui", Path("tools/ui_gen/templates/landing_html.template"), None, b"\n<!-- Bazel incremental source probe -->\n", {"//dist:runtime", "//dist:ui"}),
        ("cli", Path("CLI/actioner/cli_client.py"), None, b"\n# Bazel incremental source probe\n", {"//dist:runtime", "//CLI:runtime"}),
    ]
    for name, path, old, new, changed in edits:
        original = path.read_bytes()
        try:
            if old is not None:
                assert old in original
                modified = original.replace(old, new)
            else:
                modified = original + new
            path.write_bytes(modified)
            edited = snapshot(name + "-edit")
            actual = {target for target in TARGETS if edited["sha256"][target] != result["baseline"]["sha256"][target]}
            assert actual == changed, (name, actual, changed)
            result[name + "_edit"] = edited
        finally:
            path.write_bytes(original)
        restored = snapshot(name + "-restore")
        assert restored["sha256"] == result["baseline"]["sha256"], name
        result[name + "_restore"] = restored
        (OUTPUT / "incremental.json").write_text(json.dumps(result, indent=2) + "\n")
    # A harmless configure-script comment changes the declared foreign action,
    # rebuilding the same native sources in another temporary directory. This
    # catches libtool build-directory paths in deployed ELF files/build IDs.
    configure = Path("CLI/bazel/configure.sh")
    original = configure.read_bytes()
    try:
        configure.write_bytes(original + b"\n# Native CLI reproducibility probe.\n")
        result["cli_rebuild"] = snapshot("cli-rebuild")
        assert result["cli_rebuild"]["sha256"] == result["baseline"]["sha256"]
    finally:
        configure.write_bytes(original)
    result["cli_configure_restore"] = snapshot("cli-configure-restore")
    assert result["cli_configure_restore"]["sha256"] == result["baseline"]["sha256"]
    (OUTPUT / "incremental.json").write_text(json.dumps(result, indent=2) + "\n")
    print(json.dumps(result, indent=2))


if __name__ == "__main__":
    main()

"""Build the public TAR interface from an independent Bazel root."""

import hashlib
import json
import os
from pathlib import Path
import shlex
import shutil
import subprocess
import sys
import xml.etree.ElementTree as ET

# Support workers that exclude the script directory with PYTHONSAFEPATH.
sys.path.insert(0, str(Path(__file__).resolve().parent))
from go_versions import verify as verify_go_versions

BAZEL = shlex.split(os.environ.get("BAZEL_COMMAND", "bazel"))
ROOT = Path.cwd().resolve()
OUT = ROOT / "bazel-evidence" / "downstream"
TARGETS = ["//:image_payload", "@sonic-mgmt-framework//dist:runtime", "@sonic-mgmt-framework//dist:runtime.debug_symbols"]
TESTS = ["dist:package_test", "dist:rest_runtime_test", "CLI:installed_runtime_test",
         "tools/bazel:models_repro_test", "tools/bazel:codegen_config_test"]


def main():
    OUT.mkdir(parents=True, exist_ok=True)
    output_base = Path(subprocess.check_output([*BAZEL, "info", "output_base"], text=True).strip())
    workspace = output_base.with_name(output_base.name + "-consumer-workspace")
    consumer_base = output_base.with_name(output_base.name + "-consumer-output")
    workspace.mkdir(exist_ok=True)
    shutil.copytree(ROOT / "tools/bazel/consumer", workspace, dirs_exist_ok=True)
    shutil.copyfile(ROOT / ".bazelversion", workspace / ".bazelversion")
    if (ROOT / ".bazelrc").exists():
        shutil.copyfile(ROOT / ".bazelrc", workspace / ".bazelrc")
    prefix = [*BAZEL, "--output_base=" + str(consumer_base)]
    flags = ["--override_module=sonic-mgmt-framework=" + str(ROOT)]
    # A consumer's stripping policy must not empty the public debug TAR.
    profile = [*flags, "--compilation_mode=fastbuild", "--strip=always"]

    def run(name, *arguments):
        with (OUT / (name + ".log")).open("w") as log:
            result = subprocess.run([*prefix, *arguments], cwd=workspace,
                                    stdout=log, stderr=subprocess.STDOUT)
        if result.returncode:
            print((OUT / (name + ".log")).read_text()[-16000:])
            raise subprocess.CalledProcessError(result.returncode, arguments)

    def query(*arguments):
        return subprocess.check_output([*prefix, *arguments], cwd=workspace, text=True).strip()

    run("build", "build", *profile, *TARGETS)
    run("tests", "test", *profile, *["@sonic-mgmt-framework//" + test for test in TESTS])
    reports = {}
    # Use the link created by the successful test command. Bazel 8's `info
    # bazel-testlogs` can resolve a platform alias before its repository mapping.
    logs = (workspace / "bazel-testlogs").resolve(strict=True)
    for test in TESTS:
        package, name = test.split(":")
        directory = logs / "external/sonic-mgmt-framework+" / package / name
        parsed = ET.parse(directory / "test.xml")
        cases = parsed.findall(".//testcase")
        assert cases and not parsed.findall(".//failure") and not parsed.findall(".//error") and not parsed.findall(".//skipped"), test
        destination = OUT / "testlogs" / package / name
        destination.mkdir(parents=True, exist_ok=True)
        for filename in ("test.xml", "test.log"):
            shutil.copyfile(directory / filename, destination / filename)
        reports[test] = len(cases)
    artifacts = {}
    for target in TARGETS:
        artifact = workspace / query("cquery", *profile, target, "--output=files")
        destination = OUT / artifact.name
        shutil.copyfile(artifact, destination)
        artifacts[target] = {
            "file": destination.name,
            "sha256": hashlib.sha256(destination.read_bytes()).hexdigest(),
            "size": destination.stat().st_size,
        }
    shutil.copyfile(workspace / "MODULE.bazel.lock", OUT / "MODULE.bazel.lock")
    consumer_graph = query("mod", "graph", *flags, "--output=json")
    (OUT / "module-graph.json").write_text(consumer_graph + "\n")
    release = query("info", *flags, "release")
    assert release == "release 8.5.1", release
    def version_query(command, subcommand, *arguments):
        return query(command, subcommand, *flags, *arguments)
    versions = verify_go_versions(version_query, ROOT / "go.mod", OUT / "go-runtime-versions.json")
    # Each root resolves infra through the registry. Record the consumer copy
    # and verify it matches the module materialized by the source build.
    infra_module = Path("external/sonic-build-infra+/MODULE.bazel")
    consumer_infra = (consumer_base / infra_module).read_bytes()
    assert consumer_infra == (output_base / infra_module).read_bytes(), "source and consumer selected different infra modules"
    evidence = {"complete": True, "bazel_release": release, "consumer_compilation_mode": "fastbuild", "consumer_strip": "always", "source_checkout": str(ROOT), "source_revision": subprocess.check_output(["git", "rev-parse", "HEAD"], text=True).strip(),
                "infra_module_sha256": hashlib.sha256(consumer_infra).hexdigest(),
                "tests": reports, "artifacts": artifacts, "go_runtime_versions": versions}
    (OUT / "validation.json").write_text(json.dumps(evidence, indent=2) + "\n")
    print(json.dumps(evidence, indent=2))


if __name__ == "__main__":
    main()

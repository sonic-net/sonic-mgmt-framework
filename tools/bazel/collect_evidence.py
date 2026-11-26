"""Retain exact public outputs, declared source actions and required test reports."""

import hashlib
import json
import os
from pathlib import Path
import re
import shlex
import shutil
import subprocess
import sys
import xml.etree.ElementTree as ET

sys.path.insert(0, str(Path(__file__).resolve().parent))
from go_versions import verify as verify_go_versions

BAZEL = shlex.split(os.environ.get("BAZEL_COMMAND", "bazel"))
OUT = Path("bazel-evidence")
TARGETS = ["//dist:runtime", "//dist:runtime.debug_symbols"]
TESTS = ["tools/bazel/codegen_config_test", "tools/bazel/models_repro_test", "rest/server/server_test", "tools/openapi_tests/generator_test", "dist/rest_runtime_test",
         "dist/package_test", "CLI/cli_client_test", "CLI/generator_order_test", "CLI/rest_tls_test", "CLI/installed_runtime_test"]


def command(*args):
    return subprocess.check_output([*BAZEL, *args], text=True)


def main():
    OUT.mkdir(exist_ok=True)
    complete = all(os.environ.get(name) == "success" for name in ("CI_TEST_RESULT", "CI_INCREMENTAL_RESULT", "CI_DEBUGGER_RESULT", "CI_CONSUMER_RESULT"))
    reports = {}
    for name in TESTS:
        directory = Path("bazel-testlogs") / name
        xml = directory / "test.xml"
        if not xml.exists():
            if complete:
                raise AssertionError("Missing required test report: " + name)
            continue
        parsed = ET.parse(xml)
        cases = parsed.findall(".//testcase")
        if complete:
            assert cases and not parsed.findall(".//failure") and not parsed.findall(".//error"), name
        destination = OUT / "testlogs" / name
        destination.mkdir(parents=True, exist_ok=True)
        for file in (xml, directory / "test.log"):
            shutil.copyfile(file, destination / file.name)
        reports[name] = len(cases)
    source_log = Path("bazel-testlogs/rest/server/server_test/test.log")
    if complete:
        text = source_log.read_text()
        passed = re.findall(r"^\s*--- PASS: (\S+)", text, re.M)
        skipped = re.findall(r"^\s*--- SKIP: (\S+)", text, re.M)
        assert passed and len(skipped) == 7 and not re.search(r"^\s*--- FAIL:", text, re.M)
        reports["rest_source_cases"] = {"passed": len(passed), "skipped": skipped}
    artifacts = {}
    if complete:
        for target in TARGETS:
            artifact = Path(command("cquery", target, "--output=files").strip())
            destination = OUT / ("runtime.debug_symbols.tar" if "debug_symbols" in target else "runtime.tar")
            shutil.copyfile(artifact, destination)
            artifacts[target] = {"file": destination.name, "sha256": hashlib.sha256(destination.read_bytes()).hexdigest(), "size": destination.stat().st_size}
        debugger = json.loads((OUT / "debugger.json").read_text())
        package_log = Path("bazel-testlogs/dist/package_test/test.log").read_text()
        package_records = [json.loads(line) for line in package_log.splitlines()
                           if line.startswith('{"debug_sha256":')]
        assert len(package_records) == 1, "Missing installed-package evidence"
        package = package_records[0]
        for record in (debugger, package):
            assert record["runtime_sha256"] == artifacts["//dist:runtime"]["sha256"]
            assert record["debug_sha256"] == artifacts["//dist:runtime.debug_symbols"]["sha256"]
        assert package["matching_elf_pairs"] == debugger["matching_symbol_lookups"] == len(debugger["checks"]) == 8
        incremental = json.loads((OUT / "incremental.json").read_text())
        for stage in ("baseline", "noop", "rest_restore", "ui_restore", "cli_restore", "cli_rebuild", "cli_configure_restore"):
            for target in TARGETS:
                assert incremental[stage]["sha256"][target] == artifacts[target]["sha256"], stage
        reports["installed_payload"] = package
        graph = json.loads(command("aquery", "deps(//dist:runtime) + deps(//dist:runtime.debug_symbols)", "--output=jsonproto", "--include_commandline=false"))
        fragments = {fragment["id"]: fragment for fragment in graph.get("pathFragments", [])}
        artifacts_by_id = {artifact["id"]: artifact for artifact in graph.get("artifacts", [])}
        def artifact_path(identifier):
            parts = []
            while identifier:
                fragment = fragments[identifier]
                parts.append(fragment["label"])
                identifier = fragment.get("parentId")
            return "/".join(reversed(parts))
        declared = {artifact_path(artifact["pathFragmentId"]) for artifact in artifacts_by_id.values()}
        rest_sources = {str(path) for path in Path("rest").rglob("*.go") if not path.name.endswith("_test.go")}
        cli_sources = {str(path) for path in Path("CLI/klish/patches/klish-2.1.4").rglob("*") if path.is_file()}
        for name, sources in (("rest", rest_sources), ("patched_klish", cli_sources)):
            assert sources and sources <= declared, (name, sorted(sources - declared))
        (OUT / "source-coverage.json").write_text(json.dumps({
            "rest_sources": sorted(rest_sources), "patched_klish_inputs": sorted(cli_sources),
            "all_declared_in_runtime_actions": True,
        }, indent=2) + "\n")
        # Locked external DEBs are allowed inputs. Only generated action outputs
        # would violate the producer scope.
        outputs = [fragments[artifacts_by_id[identifier]["pathFragmentId"]]["label"]
                   for action in graph.get("actions", []) for identifier in action.get("outputIds", [])]
        assert not any(path.endswith((".deb", ".ddeb", ".udeb")) for path in outputs), "DEB output in source graph"
        counts = {}
        for action in graph.get("actions", []):
            name = action["mnemonic"]
            counts[name] = counts.get(name, 0) + 1
        assert counts.get("GoCompilePkg", 0) and counts.get("FrameworkRestModels", 0), counts
        assert counts.get("CcConfigureMakeRule", 0) and counts.get("CliPatchedSource", 0), counts
        (OUT / "source-actions.json").write_text(json.dumps({"action_count": len(graph.get("actions", [])), "mnemonics": counts, "deb_paths": []}, indent=2) + "\n")
    lock = Path("MODULE.bazel.lock")
    if complete:
        assert lock.exists(), "Missing required owner MODULE.bazel.lock"
    if lock.exists():
        shutil.copyfile(lock, OUT / lock.name)
    revision = subprocess.check_output(["git", "rev-parse", "HEAD"], text=True).strip()
    # checkout's shallow boundary hides parents from revision walking; the
    # immutable commit object still records the actual tested merge parents.
    commit = subprocess.check_output(["git", "cat-file", "-p", "HEAD"], text=True)
    parents = [line.removeprefix("parent ") for line in commit.split("\n\n", 1)[0].splitlines()
               if line.startswith("parent ")]
    head, base = os.environ.get("CI_HEAD_SHA", ""), os.environ.get("CI_BASE_SHA", "")
    if head:
        assert head == revision or head in parents, "Tested checkout does not include the PR head"
    if base and head != revision:
        assert base in parents, "Tested merge does not include the intended base"
    if complete:
        reports["go_runtime_versions"] = verify_go_versions(command, "go.mod", OUT / "go-runtime-versions.json")
        downstream = json.loads((OUT / "downstream/validation.json").read_text())
        assert downstream["complete"] and downstream["source_revision"] == revision
        assert len(downstream["tests"]) == 5 and downstream["bazel_release"] == "release 8.5.1"
        assert downstream["consumer_compilation_mode"] == "fastbuild" and downstream["consumer_strip"] == "always"
        for target, data in downstream.get("artifacts", {}).items():
            destination = OUT / "downstream" / data["file"]
            assert destination.exists() and destination.stat().st_size == data["size"]
            assert hashlib.sha256(destination.read_bytes()).hexdigest() == data["sha256"]
        assert (OUT / "downstream/module-graph.json").exists(), "Missing consumer module-graph.json"
        reports["downstream_consumer"] = downstream
        owner_graph = command("mod", "graph", "--output=json")
        (OUT / "module-graph.json").write_text(owner_graph + "\n")
    evidence = {"tested_revision": revision, "tested_parents": parents,
                "head_sha": head, "base_sha": base, "architecture": os.uname().machine,
                "complete": complete, "reports": reports, "artifacts": artifacts}
    (OUT / "validation.json").write_text(json.dumps(evidence, indent=2) + "\n")
    print(json.dumps(evidence, indent=2))


if __name__ == "__main__":
    main()

"""Apply models/yang_to_openapi.mk's two generator passes in an action directory."""

import argparse
import os
from pathlib import Path
import re
import runpy
import shutil
import subprocess
import sys
import tempfile



def validate_codegen_config(path):
    """Keep the default Make profile explicit; reject unsupported custom codegen."""
    settings = {"YANGAPI_EXCLUDES", "PY_YANGAPI_CLIENTS", "OPENAPI_EXCLUDES", "PY_OPENAPI_CLIENTS"}
    seen = set()
    for number, line in enumerate(path.read_text().splitlines(), 1):
        line = line.split("#", 1)[0].strip()
        if not line:
            continue
        match = re.fullmatch(r"([A-Z_]+)\s*(?:\+=|:=|\?=|=)\s*", line)
        if not match or match[1] not in settings:
            raise ValueError(f"{path}:{number}: Bazel supports the default empty codegen profile; declare nonempty selections and client generators before using them: {line}")
        seen.add(match[1])
    if seen != settings:
        raise ValueError(f"{path}: missing codegen settings: {sorted(settings - seen)}")

def main():
    parser = argparse.ArgumentParser()
    for name in ("models", "pyang", "pyang-wheel", "plugin", "go", "ui", "docs", "config"):
        parser.add_argument("--" + name, required=True, type=Path)
    parser.add_argument("--asset", action="append", default=[], type=Path)
    parser.add_argument("--openapi-spec", action="append", default=[], type=Path)
    args = parser.parse_args()
    validate_codegen_config(args.config)
    if args.openapi_spec:
        raise ValueError("Handwritten OpenAPI specifications require declared upstream openapi_generator targets")
    models = args.models.absolute()
    pyang = args.pyang.absolute()
    plugin = args.plugin.absolute()
    source_root = plugin.parents[3]
    go, ui, docs = (p.absolute() for p in (args.go, args.ui, args.docs))
    # The child py_binary has its own declared dependencies and runfiles.
    # Inheriting the driver's RUNFILES_DIR selects the wrong virtualenv.
    child_env = dict(os.environ)
    for key in ("RUNFILES_DIR", "RUNFILES_MANIFEST_FILE", "PYTHON_RUNFILES"):
        child_env.pop(key, None)
    for output in (go, ui, docs):
        output.mkdir(parents=True, exist_ok=True)
    with tempfile.TemporaryDirectory(prefix="framework-models-") as temporary:
        work = Path(temporary)
        for asset in args.asset:
            source = asset.absolute()
            relative = source.relative_to(source_root)
            dest = work / relative
            dest.parent.mkdir(parents=True, exist_ok=True)
            shutil.copyfile(source, dest)
        staged_ui = work / "build/rest_server/dist/ui"
        shutil.copytree(work / "tools/swagger_codegen/ui-dist", staged_ui)
        shutil.copyfile(work / "tools/swagger_codegen/go-server/src/openapi/routes.go", go / "routes.go")
        for sources, search in (
            (sorted(models.glob("*.yang")) + sorted((models / "extensions").glob("*.yang")),
             [models / "common", models, models / "extensions"]),
            (sorted((models / "sonic").glob("*.yang")),
             [models / "common", models / "sonic/common", models / "sonic"]),
        ):
            if not sources:
                raise ValueError("The management-common model set is incomplete")
            subprocess.run([
                str(pyang), str(args.pyang_wheel.absolute()), "-f", "swaggerapi", "--outdir", str(staged_ui),
                "--plugindir", str(plugin.parent), "--with-md-doc", "--md-outdir", str(docs),
                "--with-serverstub", "--stub-outdir", str(go),
                "--template-dir", str(work / "tools/codegen/go-server/templates-yang"),
                "-p", ":".join(map(str, search)), *map(str, sources),
            ], check=True, env=child_env)
        runpy.run_path(str(work / "tools/ui_gen/src/uigen.py"), run_name="__main__")
        previous = sys.argv
        try:
            sys.argv = ["index.py", "--mdDir", str(docs)]
            runpy.run_path(str(work / "tools/restconf_doc_tools/index.py"), run_name="__main__")
        finally:
            sys.argv = previous
        shutil.copytree(staged_ui, ui, dirs_exist_ok=True)
    print(f"Generated {len(list(go.glob('*.go')))} Go sources and {len(list(ui.glob('*.yaml')))} model APIs")


if __name__ == "__main__":
    main()

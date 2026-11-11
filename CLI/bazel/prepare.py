"""Stage the existing CLI sources using declared upstream archive/XML tools."""
import argparse
import hashlib
import importlib
import json
import os
from pathlib import Path
import shutil
import subprocess
import tempfile

KLISH_SHA256 = "c3e9d65868a6fc07aa118c43dc9c43688d76ac3d2c8780599f3bbf1c2b5a1c33"


def run(argv, **kwargs):
    subprocess.run([str(item) for item in argv], check=True, **kwargs)


def tool_command(root, name):
    executable = root.resolve() / "bin" / name
    if not executable.is_file():
        raise RuntimeError(f"missing shared build tool: {name}")
    return [str(executable)]


def unpack(args):
    args.output.mkdir(parents=True, exist_ok=True)
    for archive in sorted(args.archives):
        run([args.bsdtar, "-xf", archive, "-C", args.output, "--no-same-owner"])
    # Debian absolute aliases refer to the installed root. Relocate them within
    # this declared build-input tree so Bazel never follows a host /etc or /usr.
    for path in sorted(args.output.rglob("*")):
        if path.is_symlink():
            target = os.readlink(path)
            if target.startswith("/"):
                destination = args.output / target.lstrip("/")
                path.unlink()
                path.symlink_to(os.path.relpath(destination, path.parent))
    # tzdata's data archive leaves localtime for a maintainer script. This SDK
    # and execution-tool root is deliberately UTC, with no host timezone input.
    localtime = args.output / "usr/share/zoneinfo/localtime"
    if localtime.is_symlink() and not localtime.exists():
        if os.readlink(localtime) != "../../../etc/localtime" or not (localtime.parent / "Etc/UTC").is_file():
            raise RuntimeError("unexpected declared timezone alias")
        localtime.unlink()
        localtime.symlink_to("Etc/UTC")


def klish(args):
    if hashlib.sha256(args.archive.read_bytes()).hexdigest() != KLISH_SHA256:
        raise RuntimeError("vendored klish archive differs from the reviewed Make source")
    args.output.mkdir(parents=True, exist_ok=True)
    run([args.bsdtar, "-xf", args.archive, "-C", args.output, "--strip-components=1", "--no-same-owner"])
    mapping = json.loads(args.inputs.read_text())
    for relative, source in sorted(mapping.items()):
        source = Path(source)
        destination = args.output / relative.removesuffix(".diff")
        if relative.endswith(".diff"):
            # These owner-maintained files use normal diff, not a repository patch.
            # Name the target explicitly and reject fuzz, reversal and partial apply.
            run(tool_command(args.tools, "patch") + ["--batch", "--forward", "--fuzz=0", "--no-backup-if-mismatch", str(destination), str(source)])
        else:
            destination.parent.mkdir(parents=True, exist_ok=True)
            shutil.copy2(source, destination)
    shutil.copy2(args.configure, args.output / "bazel-configure.sh")



def assets(args):
    # The working copy is action-private: legacy generators may write temporary
    # files beside their Python sources, never into the source checkout.
    mapping = json.loads(args.inputs.read_text())
    with tempfile.TemporaryDirectory(prefix="cli-inputs-") as temporary:
        source_root = Path(temporary).resolve()
        for relative, source in sorted(mapping.items()):
            destination = source_root / relative
            destination.parent.mkdir(parents=True, exist_ok=True)
            shutil.copy2(source, destination)
        out = args.output.resolve()
        out.mkdir(parents=True, exist_ok=True)
        command_tree = out / "command-tree"
        command_tree.mkdir()
        shutil.copytree(source_root / "clitree/cli-xml/include", command_tree / "include")
        for source in sorted((source_root / "clitree/cli-xml").glob("*.xml")):
            shutil.copy2(source, command_tree / source.name)
        macros = out / "cli-macro"
        shutil.copytree(source_root / "clitree/macro", macros)
        environment = dict(os.environ, PATH=str(args.tools.resolve() / "bin"), LC_ALL="C", LANG="C", PYTHONHASHSEED="0")
        scripts = source_root / "clitree/scripts"
        run([args.tools.resolve() / "bin/bash", scripts / "klish_platform_features_process.sh", source_root / "clicfg", out], cwd=scripts, env=environment)
        # Invoke the existing implementations with the same parameters as Make.
        import sys
        sys.path.insert(0, str(scripts))
        importlib.import_module("klish_replace_macro").replace_macros(str(command_tree), str(macros), "6", False)
        importlib.import_module("klish_insert_pipe").insert_pipe(str(command_tree), False)
        importlib.import_module("klish_ins_def_cmd").ins_def_cmd(str(command_tree), str(command_tree), False)
        for source in sorted((source_root / "actioner").glob("*.py")):
            shutil.copy2(source, out / source.name)
        shutil.copytree(source_root / "renderer/scripts", out / "scripts")
        shutil.copytree(source_root / "renderer/templates", out / "render-templates")
        shutil.copy2(scripts / "sonic-clish.xsd", command_tree / "sonic-clish.xsd")
        for directory in (command_tree, command_tree / "include"):
            xml_files = sorted(directory.glob("*.xml"))
            if not xml_files:
                raise RuntimeError(f"empty installed CLI XML set: {directory}")
            run(tool_command(args.tools, "xmllint") + ["--nonet", "--noout", "--schema", str(command_tree / "sonic-clish.xsd"), *map(str, xml_files)], env=environment)
        # Make's installed tree includes these macro inputs (its cleanup uses the
        # literal {TGT_DIR}) and completion marker. Retain that existing contract.
        (command_tree / ".done").touch()
        out.chmod(0o755)
        for path in out.rglob("*"):
            if path.is_dir():
                path.chmod(0o755)
            elif path.is_file():
                path.chmod(0o755 if path.stat().st_mode & 0o111 else 0o644)
                os.utime(path, (0, 0))


def main():
    parser = argparse.ArgumentParser()
    modes = parser.add_subparsers(dest="mode", required=True)
    p = modes.add_parser("unpack")
    p.add_argument("--bsdtar", type=Path, required=True)
    p.add_argument("--archives", type=Path, nargs="+", required=True)
    p.add_argument("--output", type=Path, required=True)
    p = modes.add_parser("klish")
    for name in ("bsdtar", "archive", "inputs", "tools", "configure", "output"):
        p.add_argument("--" + name, type=Path, required=True)
    p = modes.add_parser("assets")
    for name in ("inputs", "tools", "output"):
        p.add_argument("--" + name, type=Path, required=True)
    args = parser.parse_args()
    {"unpack": unpack, "klish": klish, "assets": assets}[args.mode](args)


if __name__ == "__main__":
    main()

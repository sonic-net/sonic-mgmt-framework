"""Compare complete installed payloads and every deployed ELF/debug pair."""

import hashlib
import json
from pathlib import Path
import re
import struct
import subprocess
import sys
import tarfile
import tempfile
import unittest
import zlib

RUNTIME, DEBUG, UI, CLI, READELF, OBJCOPY = [Path(p).absolute() for p in sys.argv[1:7]]
del sys.argv[1:7]
LIBRARIES = ("libclish", "libkonf", "liblub", "libtinyrl")


def files(archive):
    with tarfile.open(archive) as source:
        result = {}
        for member in source.getmembers():
            name = member.name.removeprefix("./")
            if not member.isdir():
                if name in result:
                    raise AssertionError("Duplicate installed path: " + name)
                result[name] = (member, source.extractfile(member).read() if member.isfile() else None)
        return result


def readelf(*args):
    return subprocess.check_output([str(READELF), *map(str, args)], text=True)


class PackageContractTest(unittest.TestCase):
    def test_installed_payload_and_debug(self):
        runtime, debug = files(RUNTIME), files(DEBUG)
        assets = {
            "etc/rest_server/ui/" + str(p.relative_to(UI)): p.read_bytes()
            for p in UI.rglob("*") if p.is_file()
        }
        assets.update({
            "usr/sbin/cli/" + str(p.relative_to(CLI)): p.read_bytes()
            for p in CLI.rglob("*") if p.is_file()
        })
        binaries = {"usr/sbin/rest_server", "usr/sbin/generate_cert", "usr/sbin/cli/clish",
                    "usr/sbin/cli/.libs/clish_plugin_clish.so"}
        binaries.update("usr/sbin/cli/.libs/" + name + ".so.1.0.0" for name in LIBRARIES)
        static = {"usr/sbin/cli/.libs/" + name + ".a" for name in (*LIBRARIES, "clish_plugin_clish")}
        links = {"usr/sbin/cli/.libs/" + name + suffix: name + ".so.1.0.0"
                 for name in LIBRARIES for suffix in (".so", ".so.1")}
        self.assertEqual(set(runtime), set(assets) | binaries | static | set(links) | {"usr/sbin/cli/clish_start"})
        for name, (member, data) in runtime.items():
            self.assertEqual((member.uid, member.gid), (0, 0), name)
            self.assertIn(member.mtime, (0, 1672560000), name)
            if name in assets:
                self.assertEqual(data, assets[name], name)
            if name.startswith("etc/rest_server/ui/") or name in static:
                self.assertEqual(member.mode, 0o644, name)
            if name in binaries or name == "usr/sbin/cli/clish_start":
                self.assertEqual(member.mode, 0o755, name)
            if name in links:
                self.assertTrue(member.issym(), name)
                self.assertEqual(member.linkname, links[name], name)
        expected_symbols = set()
        with tempfile.TemporaryDirectory(prefix="framework-elf-") as temporary:
            root = Path(temporary)
            for number, name in enumerate(sorted(binaries)):
                with self.subTest(binary=name):
                    binary = root / f"binary-{number}"
                    binary.write_bytes(runtime[name][1])
                    self.assertTrue(binary.read_bytes().startswith(b"\x7fELF"))
                    self.assertIn("Advanced Micro Devices X86-64", readelf("-h", binary))
                    match = re.search(r"Build ID: ([0-9a-f]+)", readelf("-n", binary))
                    self.assertIsNotNone(match, name)
                    identity = match[1]
                    path = f"usr/lib/debug/.build-id/{identity[:2]}/{identity[2:]}.debug"
                    expected_symbols.add(path)
                    self.assertIn(path, debug)
                    symbols = root / Path(path).name
                    symbols.write_bytes(debug[path][1])
                    self.assertEqual(re.search(r"Build ID: ([0-9a-f]+)", readelf("-n", symbols))[1], identity)
                    sections = readelf("-SW", symbols)
                    self.assertIn(".debug_info", sections)
                    self.assertIn(".debug_line", sections)
                    self.assertNotRegex(readelf("-SW", binary), r"\.(?:z?debug_)\w+")
                    dynamic = readelf("-d", binary)
                    self.assertNotIn("_solib", dynamic)
                    self.assertNotIn(".runfiles", dynamic)
                    for search_path in re.findall(r"\((?:RPATH|RUNPATH)\).*?\[([^\]]*)\]", dynamic):
                        self.assertLessEqual(set(search_path.split(":")), {
                            "/usr/sbin/cli/.libs", "$ORIGIN", "$ORIGIN/.libs",
                        }, name)
                    link = root / "debuglink"
                    subprocess.run([str(OBJCOPY), "--dump-section", f".gnu_debuglink={link}", str(binary)], check=True)
                    payload = link.read_bytes()
                    end = payload.index(b"\0")
                    self.assertEqual(payload[:end].decode(), symbols.name)
                    crc = struct.unpack_from("<I", payload, (end + 4) & ~3)[0]
                    self.assertEqual(crc, zlib.crc32(symbols.read_bytes()))
        self.assertEqual(set(debug), expected_symbols)
        for name, (member, data) in debug.items():
            self.assertEqual((member.uid, member.gid), (0, 0), name)
            self.assertIn(member.mtime, (0, 1672560000), name)
            self.assertEqual(member.mode, 0o644, name)
        print(json.dumps({"runtime_sha256": hashlib.sha256(RUNTIME.read_bytes()).hexdigest(),
            "debug_sha256": hashlib.sha256(DEBUG.read_bytes()).hexdigest(),
            "installed_files_and_links": len(runtime), "matching_elf_pairs": len(binaries),
            "source_asset_files": len(assets)}, sort_keys=True))


if __name__ == "__main__":
    unittest.main()

"""Regression for shared views being emitted in filesystem-dependent files."""
import contextlib
import io
import os
from pathlib import Path
import sys
import tempfile
import unittest
from unittest import mock

sys.path.insert(0, str(Path(__file__).resolve().parents[1] / "clitree/scripts"))
import klish_ins_def_cmd


class GeneratorOrderTest(unittest.TestCase):
    def test_reversed_directory_enumeration_keeps_identical_command_tree(self):
        xml = '''<?xml version="1.0" encoding="UTF-8"?>
<CLISH_MODULE xmlns="http://www.dellemc.com/sonic/XMLSchema">
<VIEW name="config-view" prompt="config#"><COMMAND name="%s" help="fixture"/></VIEW>
</CLISH_MODULE>'''
        outputs = []
        original_walk = os.walk
        for reverse in (False, True):
            with tempfile.TemporaryDirectory() as temporary:
                root = Path(temporary)
                # The same view spans two source files; exactly one must own the
                # injected exit/end commands, independent of host directory order.
                for name in ("b.xml", "a.xml"):
                    (root / name).write_text(xml % name[0])

                def walk(path):
                    for current, dirs, files in original_walk(path):
                        yield current, dirs, sorted(files, reverse=reverse)

                with mock.patch.object(klish_ins_def_cmd.os, "walk", walk), contextlib.redirect_stdout(io.StringIO()):
                    klish_ins_def_cmd.ins_def_cmd(str(root), str(root), False)
                outputs.append({path.name: path.read_bytes() for path in root.iterdir()})
        self.assertEqual(outputs[0], outputs[1])
        self.assertIn(b'name="exit"', outputs[0]["a.xml"])
        self.assertNotIn(b'name="exit"', outputs[0]["b.xml"])


if __name__ == "__main__":
    unittest.main()

"""Run the four unchanged golden-output comparisons from this directory's Makefile."""

import difflib
import os
from pathlib import Path
import subprocess
import sys
import unittest

PYANG, WHEEL, PLUGIN, MARKER = [Path(p).absolute() for p in sys.argv[1:5]]
del sys.argv[1:5]


class GeneratorTest(unittest.TestCase):
    def compare(self, sources, expected, options=()):
        environment = dict(os.environ)
        for key in ("RUNFILES_DIR", "RUNFILES_MANIFEST_FILE", "PYTHON_RUNFILES"):
            environment.pop(key, None)
        result = subprocess.run(
            [str(PYANG), str(WHEEL), "-f", "swaggerapi", "--plugindir", str(PLUGIN.parent),
             *options, *sources],
            cwd=MARKER.parent, env=environment, check=True, stdout=subprocess.PIPE,
        ).stdout.decode()
        golden = (MARKER.parent / expected).read_text()
        self.assertEqual(golden, result, "".join(difflib.unified_diff(
            golden.splitlines(True), result.splitlines(True), fromfile=expected, tofile="generated")))

    def test_rpc(self):
        self.compare(["test-rpc.yang"], "test-rpc.yang.expect")

    def test_data_nodes(self):
        self.compare(["test-data-nodes.yang"], "test-data-nodes.yang.expect")

    def test_complex_model(self):
        self.compare(["ietf-snmp.yang", "ietf-snmp-community.yang"], "ietf-snmp.yang.expect", ["--with-oneof"])

    def test_complex_model_no_oneof(self):
        self.compare(["ietf-snmp.yang", "ietf-snmp-community.yang"], "ietf-snmp.no-oneof.yang.expect")


if __name__ == "__main__":
    unittest.main()

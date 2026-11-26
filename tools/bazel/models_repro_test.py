"""Regenerate all REST outputs with different process hashes and compare bytes."""

import hashlib
import json
import os
from pathlib import Path
import subprocess
import sys
import tempfile
import unittest

DRIVER, PYANG, WHEEL, MODELS, PLUGIN, CONFIG = [Path(p).absolute() for p in sys.argv[1:7]]
ASSETS = [Path(p).absolute() for p in sys.argv[7:]]
sys.argv[1:] = []


def digest(root):
    return {str(p.relative_to(root)): hashlib.sha256(p.read_bytes()).hexdigest()
            for p in root.rglob("*") if p.is_file()}


class ModelsReproTest(unittest.TestCase):
    def test_independent_generations(self):
        snapshots = []
        with tempfile.TemporaryDirectory(prefix="framework-repro-") as temporary:
            for index in range(5):
                output = Path(temporary) / str(index)
                environment = dict(os.environ, PYTHONHASHSEED=str(index))
                for key in ("RUNFILES_DIR", "RUNFILES_MANIFEST_FILE", "PYTHON_RUNFILES"):
                    environment.pop(key, None)
                subprocess.run([str(DRIVER), "--models", str(MODELS), "--pyang", str(PYANG),
                    "--pyang-wheel", str(WHEEL), "--plugin", str(PLUGIN), "--config", str(CONFIG),
                    "--go", str(output / "go"), "--ui", str(output / "ui"), "--docs", str(output / "docs"),
                    *[arg for asset in ASSETS for arg in ("--asset", str(asset))]],
                    check=True, env=environment)
                snapshots.append(digest(output))
            self.assertTrue(snapshots[0])
            for index, snapshot in enumerate(snapshots[1:], start=1):
                self.assertEqual(snapshots[0], snapshot, f"Independent generator {index} changed bytes")
        print(json.dumps({"generations": 5, "output_files": len(snapshots[0]), "all_sha256_equal": True}))


if __name__ == "__main__":
    unittest.main()

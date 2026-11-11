"""Reject unsupported Make codegen customizations instead of silently ignoring them."""
from pathlib import Path
import sys
import tempfile
import unittest
from tools.bazel.generate_rest_models import validate_codegen_config

CONFIG = Path(sys.argv.pop(1)).read_text()

class CodegenConfigTest(unittest.TestCase):
    def test_current_default(self):
        with tempfile.TemporaryDirectory() as directory:
            config = Path(directory) / "codegen.config"
            config.write_text(CONFIG)
            validate_codegen_config(config)

    def test_customization_requires_declared_support(self):
        for setting in ("YANGAPI_EXCLUDES", "OPENAPI_EXCLUDES", "PY_YANGAPI_CLIENTS", "PY_OPENAPI_CLIENTS"):
            with self.subTest(setting=setting), tempfile.TemporaryDirectory() as directory:
                config = Path(directory) / "codegen.config"
                config.write_text(CONFIG + f"\n{setting} += custom-module\n")
                with self.assertRaisesRegex(ValueError, "default empty codegen profile"):
                    validate_codegen_config(config)

if __name__ == "__main__":
    unittest.main()

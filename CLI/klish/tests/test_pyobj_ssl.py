#!/usr/bin/env python3

import os
from pathlib import Path
import subprocess
import sys
import tempfile


def main():
    clish = Path(sys.argv[1]).resolve()
    library_dir = Path(sys.argv[2]).resolve()

    with tempfile.TemporaryDirectory() as temporary_dir:
        root = Path(temporary_dir)
        (root / "startup.xml").write_text(
            '<CLISH_MODULE xmlns="http://www.dellemc.com/sonic/XMLSchema">'
            '<STARTUP view="test-view"/>'
            '<VIEW name="test-view" prompt="test# ">'
            '<COMMAND name="check-ssl" help="Check SSL">'
            '<ACTION builtin="clish_pyobj">ssl_probe check</ACTION>'
            '</COMMAND></VIEW></CLISH_MODULE>'
        )
        (root / "ssl_probe.py").write_text(
            'def run(func, args):\n'
            '    import os\n'
            '    import ssl\n'
            '    with open(os.environ["SSL_PROBE_MARKER"], "w") as marker:\n'
            '        marker.write(ssl.OPENSSL_VERSION)\n'
            '    return 0\n'
        )

        env = os.environ.copy()
        env.pop("LD_PRELOAD", None)
        env["CLISH_NOAUTH"] = "1"
        env["LD_LIBRARY_PATH"] = str(library_dir)
        env["PYTHONPATH"] = str(root)
        marker = root / "ssl-import-ok"
        env["SSL_PROBE_MARKER"] = str(marker)
        result = subprocess.run(
            [str(clish), "-l", "-e", "-b", "-f", "/dev/null", "-x", str(root)],
            input="check-ssl\n", env=env, stdout=subprocess.PIPE,
            stderr=subprocess.STDOUT, text=True, timeout=30,
        )
        if result.returncode or not marker.is_file():
            print("KLISH exited with {}: {}".format(result.returncode, result.stdout),
                  file=sys.stderr)
            return 1
        print("Embedded Python SSL import succeeded")
    return 0


if __name__ == "__main__":
    sys.exit(main())

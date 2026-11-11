"""Entry point for the pinned upstream pyang implementation."""

import sys
import zipfile

if __name__ == "__main__":
    # pyang 2.4 ships a wheel script rather than a Python module entry point.
    # Execute those exact pinned bytes with the declared Python dependencies.
    wheel = sys.argv.pop(1)
    with zipfile.ZipFile(wheel) as archive:
        scripts = [name for name in archive.namelist() if name.endswith(".data/scripts/pyang")]
        if len(scripts) != 1:
            raise ValueError("Expected one upstream pyang wheel script")
        source = archive.read(scripts[0])
    exec(compile(source, scripts[0], "exec"), {"__name__": "__main__", "__file__": scripts[0]})

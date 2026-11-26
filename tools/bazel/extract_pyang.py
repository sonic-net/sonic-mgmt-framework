"""Extract the packaged pyang 2.4 CLI as a declared Python source."""

from pathlib import Path
import sys
import zipfile


def main():
    wheel, output = map(Path, sys.argv[1:])
    with zipfile.ZipFile(wheel) as archive:
        [entry] = [name for name in archive.namelist() if name.endswith(".data/scripts/pyang")]
        output.write_bytes(archive.read(entry))


if __name__ == "__main__":
    main()

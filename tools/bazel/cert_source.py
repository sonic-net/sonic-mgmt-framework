"""Select the Go SDK's standalone certificate program as a normal Go source."""
from pathlib import Path
import sys
source, output = map(Path, sys.argv[1:])
text = source.read_text()
assert text.count("//go:build ignore\n") == 1
output.write_text(text.replace("//go:build ignore\n", "", 1))

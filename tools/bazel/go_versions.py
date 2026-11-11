"""Check the deployed Go libraries affected by Bazel build-tool requirements."""

import json
from pathlib import Path
import re


def verify(query, source, output):
    native = dict(re.findall(
        r"^\s*(golang.org/x/(?:crypto|net|sys|text))\s+(v\S+)",
        Path(source).read_text(), re.M))
    assert len(native) == 4, "Missing native REST dependency versions"
    repositories = ["@@gazelle++go_deps+org_golang_x_" + name.rsplit("/", 1)[1]
                    for name in sorted(native)]
    declarations = query("mod", "show_repo", *repositories)
    selected = {}
    for record in re.findall(r"go_repository\(\n(.*?)\n\)", declarations, re.S):
        fields = dict(re.findall(r'^  (\w+) = "([^"\n]*)",', record, re.M))
        if fields.get("importpath") in native:
            selected[fields["importpath"]] = fields["version"]
    assert selected == native, ("Bazel changed deployed Go library versions", native, selected)
    Path(output).write_text(json.dumps({"native_go_mod": native, "bazel": selected}, indent=2) + "\n")
    return selected

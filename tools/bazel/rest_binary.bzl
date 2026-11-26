"""Preserve REST's legacy Go flags independently of the consuming workspace."""

_GO_OPTIONS = str(Label("@io_bazel_rules_go//go/config:gc_goopts"))

def _rest_flags_impl(_settings, _attr):
    # rest/Makefile applies these flags to all Go packages, leaving target C/C++
    # settings and the certificate tool's normal optimization unchanged.
    return {_GO_OPTIONS: ["-N", "-l"]}

_rest_flags = transition(
    implementation = _rest_flags_impl,
    inputs = [],
    outputs = [_GO_OPTIONS],
)

def _rest_binary_impl(ctx):
    binary = ctx.attr.binary[0][DefaultInfo]
    output = ctx.actions.declare_file(ctx.label.name)
    ctx.actions.symlink(output = output, target_file = binary.files_to_run.executable, is_executable = True)
    return [DefaultInfo(files = depset([output]), executable = output, runfiles = binary.default_runfiles)]

rest_binary = rule(
    implementation = _rest_binary_impl,
    executable = True,
    attrs = {
        "binary": attr.label(mandatory = True, cfg = _rest_flags),
        "_allowlist_function_transition": attr.label(default = "@bazel_tools//tools/allowlists/function_transition_allowlist"),
    },
)

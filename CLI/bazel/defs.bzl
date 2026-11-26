"""CLI-specific orchestration; archive work uses the upstream BSD tar toolchain."""

load("@bazel_skylib//lib:paths.bzl", "paths")

# pkg_tar tree-artifact paths include the external repository when consumed.
ASSETS_STRIP_PREFIX = "/" + paths.join(Label("//CLI:installed_assets").workspace_root, "CLI/installed_assets")

_TAR = "@tar.bzl//tar/toolchain:type"

def _unpack_impl(ctx):
    output = ctx.actions.declare_directory(ctx.label.name)
    bsdtar = ctx.toolchains[_TAR]
    args = ctx.actions.args()
    args.add("unpack")
    args.add("--bsdtar", bsdtar.tarinfo.binary)
    args.add("--output", output.path)
    args.add("--archives")
    args.add_all(ctx.files.archives)
    ctx.actions.run(
        executable = ctx.executable._prepare,
        arguments = [args],
        inputs = ctx.files.archives,
        tools = bsdtar.default.files,
        outputs = [output],
        mnemonic = "CliDeclaredInputs",
    )
    return [DefaultInfo(files = depset([output]))]

cli_unpack = rule(
    implementation = _unpack_impl,
    attrs = {
        "archives": attr.label_list(allow_files = True, mandatory = True),
        "_prepare": attr.label(default = Label(":prepare"), executable = True, cfg = "exec"),
    },
    toolchains = [_TAR],
)

def _prepare_impl(ctx):
    output = ctx.actions.declare_directory(ctx.label.name)
    manifest = ctx.actions.declare_file(ctx.label.name + ".inputs.json")
    mappings = {}
    for file in ctx.files.srcs:
        marker = "/" + ctx.attr.source_prefix + "/"
        path = "/" + file.short_path
        if marker not in path:
            fail("CLI input is outside %s: %s" % (ctx.attr.source_prefix, file.short_path))
        relative = path.split(marker, 1)[1]
        if relative in mappings:
            fail("duplicate CLI source: " + relative)
        mappings[relative] = file.path
    ctx.actions.write(manifest, json.encode(mappings))
    args = ctx.actions.args()
    args.add("assets")
    args.add("--inputs", manifest)
    args.add("--output", output.path)
    args.add("--tools", ctx.file.tools.path)
    inputs = ctx.files.srcs + [manifest, ctx.file.tools]
    tools = []
    ctx.actions.run(
        executable = ctx.executable._prepare,
        arguments = [args],
        inputs = inputs,
        tools = tools,
        outputs = [output],
        env = {"LC_ALL": "C", "LANG": "C", "PYTHONHASHSEED": "0"},
        mnemonic = "CliCommandTree",
    )
    return [DefaultInfo(files = depset([output]))]

cli_prepare = rule(
    implementation = _prepare_impl,
    attrs = {
        "srcs": attr.label_list(allow_files = True, mandatory = True),
        "source_prefix": attr.string(mandatory = True),
        "tools": attr.label(allow_single_file = True, cfg = "exec", mandatory = True),
        "_prepare": attr.label(default = Label(":prepare"), executable = True, cfg = "exec"),
    },
)

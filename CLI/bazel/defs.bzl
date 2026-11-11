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
    # foreign_cc derives a source root from the tree artifact's parent. Give
    # it a dedicated parent so sibling generated CLI outputs are not inputs.
    output = ctx.actions.declare_directory(ctx.label.name + "/source" if ctx.attr.mode == "klish" else ctx.label.name)
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
    args.add(ctx.attr.mode)
    args.add("--inputs", manifest)
    args.add("--output", output.path)
    args.add("--tools", ctx.file.tools.path)
    inputs = ctx.files.srcs + [manifest, ctx.file.tools]
    tools = []
    if ctx.attr.mode == "klish":
        bsdtar = ctx.toolchains[_TAR]
        args.add("--bsdtar", bsdtar.tarinfo.binary)
        args.add("--archive", ctx.file.archive)
        args.add("--configure", ctx.file.configure)
        inputs += [ctx.file.archive, ctx.file.configure]
        tools += bsdtar.default.files.to_list()
    ctx.actions.run(
        executable = ctx.executable._prepare,
        arguments = [args],
        inputs = inputs,
        tools = tools,
        outputs = [output],
        env = {"LC_ALL": "C", "LANG": "C", "PYTHONHASHSEED": "0"},
        mnemonic = "CliPatchedSource" if ctx.attr.mode == "klish" else "CliCommandTree",
    )
    return [DefaultInfo(files = depset([output]))]

cli_prepare = rule(
    implementation = _prepare_impl,
    attrs = {
        "mode": attr.string(values = ["klish", "assets"], mandatory = True),
        "srcs": attr.label_list(allow_files = True, mandatory = True),
        "source_prefix": attr.string(mandatory = True),
        "archive": attr.label(allow_single_file = True),
        "configure": attr.label(allow_single_file = True),
        "tools": attr.label(allow_single_file = True, cfg = "exec", mandatory = True),
        "_prepare": attr.label(default = Label(":prepare"), executable = True, cfg = "exec"),
    },
    toolchains = [_TAR],
)

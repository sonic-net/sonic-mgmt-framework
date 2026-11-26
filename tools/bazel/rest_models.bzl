"""Run the existing YANG/OpenAPI generators with declared owner inputs."""

def _rest_models_impl(ctx):
    generated_go = ctx.actions.declare_directory(ctx.label.name + "_go")
    ui = ctx.actions.declare_directory(ctx.label.name + "_ui")
    docs = ctx.actions.declare_directory(ctx.label.name + "_docs")
    args = ctx.actions.args()
    args.add("--models", ctx.file.models.path)
    args.add("--config", ctx.file.config)
    args.add("--pyang", ctx.executable._pyang)
    args.add("--pyang-wheel", ctx.file._pyang_wheel)
    args.add("--plugin", ctx.file.plugin)
    args.add("--go", generated_go.path)
    args.add("--ui", ui.path)
    args.add("--docs", docs.path)
    args.add_all(ctx.files.assets, before_each = "--asset")
    args.add_all(ctx.files.openapi_specs, before_each = "--openapi-spec")
    ctx.actions.run(
        executable = ctx.executable._driver,
        arguments = [args],
        inputs = [ctx.file.models, ctx.file.plugin, ctx.file._pyang_wheel, ctx.file.config] + ctx.files.assets + ctx.files.openapi_specs,
        tools = [ctx.attr._driver[DefaultInfo].files_to_run, ctx.attr._pyang[DefaultInfo].files_to_run],
        outputs = [generated_go, ui, docs],
        env = {"PYTHONHASHSEED": "0", "PYTHONNOUSERSITE": "1"},
        execution_requirements = {"block-network": "1"},
        mnemonic = "FrameworkRestModels",
    )
    return [
        DefaultInfo(files = depset([generated_go])),
        OutputGroupInfo(ui = depset([ui]), docs = depset([docs])),
    ]

rest_models = rule(
    implementation = _rest_models_impl,
    attrs = {
        "models": attr.label(allow_single_file = True, mandatory = True),
        "config": attr.label(allow_single_file = True, mandatory = True),
        "plugin": attr.label(allow_single_file = True, mandatory = True),
        "assets": attr.label_list(allow_files = True),
        "openapi_specs": attr.label_list(allow_files = [".yaml"]),
        "_driver": attr.label(default = "//tools/bazel:generate_rest_models", executable = True, cfg = "exec"),
        "_pyang": attr.label(default = "//tools/bazel:pyang", executable = True, cfg = "exec"),
        "_pyang_wheel": attr.label(default = "@framework_pip//pyang:whl", allow_single_file = True, cfg = "exec"),
    },
)

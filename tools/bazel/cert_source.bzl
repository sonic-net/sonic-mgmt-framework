"""Prepare Make's certificate tool from the selected public Go SDK provider."""

_GO_TOOLCHAIN = "@io_bazel_rules_go//go:toolchain"

def _cert_source_impl(ctx):
    sdk = ctx.toolchains[_GO_TOOLCHAIN].sdk
    if sdk.version != "1.25.9":
        fail("The certificate source contract requires Go 1.25.9, got " + sdk.version)
    source_path = sdk.root_file.dirname + "/src/crypto/tls/generate_cert.go"
    sources = [source for source in sdk.srcs.to_list() if source.path == source_path]
    if len(sources) != 1:
        fail("The selected Go SDK must contain crypto/tls/generate_cert.go")
    output = ctx.actions.declare_file("generate_cert.go")
    ctx.actions.run(
        executable = ctx.executable._prepare,
        arguments = [sources[0].path, output.path],
        inputs = sources,
        outputs = [output],
        mnemonic = "FrameworkCertificateSource",
    )
    return [DefaultInfo(files = depset([output]))]

cert_source = rule(
    implementation = _cert_source_impl,
    attrs = {
        "_prepare": attr.label(
            default = "//tools/bazel:cert_source",
            executable = True,
            cfg = "exec",
        ),
    },
    toolchains = [_GO_TOOLCHAIN],
)

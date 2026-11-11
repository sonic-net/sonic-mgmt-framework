# Native AMD64 CLI source outputs

`//CLI:runtime` installs the existing Framework CLI under `/usr/sbin/cli`.
`//CLI:runtime.debug_symbols` contains the matching split debug files. Neither
output generates a Debian package. The repository's existing Make entry points
remain available.

The build has separate actions for:

- `//CLI:klish_sources`: unpack the checked-in klish 2.1.4 archive with the
  registered BSD tar toolchain, apply the existing normal-diff patches with a
  declared GNU patch, and copy the existing added source files.
- `//CLI:klish`: build those sources with `rules_foreign_cc.configure_make`, the
  shared GCC Trixie toolchain and shared declared Autotools runtime.
- `//CLI:installed_assets`: run the existing platform-feature script, macro
  expansion, pipe insertion and default-command insertion, then validate every
  installed command XML against the existing XSD with declared xmllint.
- `//CLI:assets` and `//CLI:klish_runtime`: package the assets and native files;
  the final runtime TAR combines them without compiling either again.

The native contract includes `clish`, `clish_start`, the five shared libraries
and the static archives copied by `CLI/klish/Makefile`. The asset contract
includes the actioner Python files, renderer scripts/templates, command XML/XSD,
and the macro inputs and completion marker currently retained by Make. Existing
script executable bits are preserved. Runtime TAR timestamps are fixed, and
compiler prefix maps remove action-directory paths from debugging information.

The CLI selects libxml2, libcurl, Python 3.13 embedding and cJSON development
inputs from the published build-infra APT set `framework_cli_target_debian`.
Its shared foreign runtime supplies GNU patch, xmllint and xsltproc for the
generators in the execution configuration. The isolated build-input tree
relocates Debian aliases inside the tree and uses UTC during the build.

The CLI's runtime Python imports use the root `framework_pip` lock (requests,
six, natsort and Jinja2); lxml is a generator dependency. A containing image must
supply those runtime Python packages and the native external library closure.

To update CLI system packages, update the Framework set in the build-infra APT
lock and select its published module in both the source and consumer roots.

The original `//CLI:cli_client_test` remains intact. The generator-order
regression verifies that a view shared by two XML files receives its default
commands in the same file even when directory enumeration is reversed.

`//CLI:rest_tls_test` preserves the original C++ TLS URL/CA assertions.
`//CLI:installed_runtime_test` extracts the produced runtime TAR and executes
its `clish` with the declared Trixie loader, native libraries and Python standard
library. It checks the version, parses a real installed `show tech-support`
command and rejects an unknown command. This is command-tree validation, not
an assertion that the REST service or switch hardware is running.

CI separately runs GDB against all eight Framework native executables/libraries:

```sh
python3 CLI/bazel/verify_gdb.py --runtime bazel-bin/dist/runtime.tar \
  --debug bazel-bin/dist/runtime.debug_symbols.tar --gdb /usr/bin/gdb \
  --output bazel-evidence/gdb.json
```

The verifier requires a source-line lookup for each binary with the supplied
debug TAR, and repeats every lookup without it to prove no host symbols satisfy
the check. Pass `--cli-only` with the CLI runtime/debug TARs to check its six
ELFs alone. `//dist:package_test` verifies the separate byte/ownership/mode,
AMD64, build-ID, debuglink CRC and debug-section contracts.

# Native CLI source outputs

`//CLI:runtime` installs the existing Framework CLI under `/usr/sbin/cli`.
`//CLI:runtime.debug_symbols` contains the matching split debug files. Neither
output generates a Debian package. The repository's existing Make entry points
remain available.

The build has separate actions for:

- `@framework_klish`: use the shared source-preparation rule to extract the
  checked-in klish 2.1.4 archive, apply the existing replacement files and
  unified patches, and expose individual source files.
- `//CLI:klish`: build those sources with native `cc_library`,
  `cc_shared_library` and `cc_binary` targets using the GCC Trixie toolchain.
  The configuration template preserves the original options and the reviewed
  target-toolchain probe; normal builds do not invoke Autotools or Make.
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
Compilation consumes the APT library targets directly through `CcInfo`.
Declared execution tools supply xmllint and xsltproc for the asset generators.
The separate installed-runtime test input relocates Debian aliases within its
declared library tree and uses UTC.

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
selected target architecture, build-ID, debuglink CRC and debug-section contracts.

## ARM64 compilation

The native CLI rules accept AMD64 and ARM64 Linux/Trixie target platforms. The
reviewed target-toolchain probes yield the same configuration macros on both
architectures, including the LP64 type sizes. Use
`--platforms=@sonic_build_infra//platforms:aarch64_trixie` for ARM64 output.

The complete CLI also requires the matching ARM64 view of the production
`framework_cli_target_debian` set. The currently selected Infra .44 lock provides
that set only for AMD64 and lacks ARM64 curl/cJSON development inputs. These
producer inputs must be available before the complete ARM64 CLI can build.
ARM64 validation can use the shared updater to add same-version ARM64 SDK
requests in an isolated Infra copy. Verify that existing package records remain
unchanged, then supply that validation copy through an explicit module override.
The registry entry documents the recorded validation inputs. The production
lock is not changed by this Framework adaptation.
The installed CLI runtime test retains its AMD64 loader and platform constraint.

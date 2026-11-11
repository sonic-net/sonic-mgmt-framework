# Native AMD64 Bazel build

Bazel builds this repository's REST server, certificate tool and patched klish
CLI from source. Management-common owns CVL/translib, prepared YANG models and
generated Go bindings; Framework consumes those targets and runs its own
YANG-to-REST generator. Runtime and debug TARs feed Docker composition directly.
No target generates a DEB, and the existing Make entry points remain available.

| Target | Output |
| --- | --- |
| `//rest/main:rest_server` | REST executable and generated API routes |
| `//rest/cert:generate_cert` | Certificate tool from the pinned Go SDK source |
| `//models:rest_models`, `:ui`, `:docs` | Generated Go API, UI and API documentation |
| `//CLI:runtime` | Patched klish, matching libraries, command tree and Python assets |
| `//dist:runtime` | Complete legacy install layout under `/usr/sbin` and `/etc/rest_server/ui` |
| `//dist:runtime.debug_symbols` | Symbols split from the same eight deployed ELF files |

The selected configuration is native AMD64 Debian Trixie with Bazel 8.5.1,
Go 1.25.9, Python 3.13 and the shared GCC toolchain. REST compilation preserves
Make's `-gcflags="all=-N -l"` on the target itself, including when another
Bazel root consumes it; the certificate tool keeps normal Go optimization. External native libraries and execution tools use
reviewed package locks. CLI source preparation and compilation use the shared
build-tools runtime; the component lock contains target libraries only.

## Review dependencies

Management Common `1.0.1-05f83dcab8fe2e1a0e0fc37ac172b289d6968e2b` and the
public goyang/ygot owners are available in the maintained SONiC registry after
[Management Common #3](https://github.com/securely1g/sonic-mgmt-common/pull/3),
[registry #214](https://github.com/securely1g/sonic-bazel-registry/pull/214) and
[registry #216](https://github.com/securely1g/sonic-bazel-registry/pull/216).
Source CI and the separate consumer resolve the published build-infra 0.0.28
module from registry main. Its foreign runtime supplies GNU patch, xmllint and
xsltproc for the CLI generators. [Shared CLI execution tools](https://github.com/securely1g/sonic-build-infra/pull/72).

```sh
bazel build //dist:runtime //dist:runtime.debug_symbols
```

`MODULE.bazel.lock` is generated and ignored; CI retains it as evidence. Package
locks stay checked in. The maintained BUILD targets use the selected Common
module directly, preserving the legacy neighboring-checkout `go.mod` declaration.
`go.bazel.work` preserves the four deployed `golang.org/x` library versions
from the native `go.mod`; Bazel build-tool requirements must not silently upgrade
the REST runtime. Downstream roots load this exported work file, as shown in
`tools/bazel/consumer/MODULE.bazel`. The native `go.mod` and `go.sum` are unchanged.
The public goyang/ygot modules own their shared Go types. Remaining Go dependency
patches are root-only settings, and consuming roots select the owner-exported
patches explicitly. The certificate generator reads its source
from the selected public Go toolchain provider, without assuming an SDK repository
name that only works in this checkout.

## Validation

CI starts from a checkout without a generated module lock. It builds the public
TARs, measures a no-op, independently edits REST source/UI templates/CLI assets,
and checks that only the corresponding outputs change. Restoring each input
must reproduce every original hash.

Klish uses libtool's install/relink step before packaging so its library search
paths name `/usr/sbin/cli/.libs`. A harmless configure-script edit also rebuilds
the same native sources in a new temporary directory and must preserve every
TAR hash; package checks reject temporary build paths in ELF search paths.

`tools/bazel/check_consumer.py` also builds the public runtime/debug TARs from an
independent root, composing them with Common data as a Docker consumer does. It
runs five owner package/generator/installed-service tests and retains their actual
reports and generated resolution lock. The test supplies this Framework checkout
as the module under test. It also verifies that the independent root resolved the
same published build-infra module as the source build.

The original REST suite keeps its seven opt-in SSH/TACACS authentication skips;
those require configured user accounts and are not silently enabled against the
CI host. The legacy `rest/main/main_test.go` is a long-running coverage daemon,
so `//rest/main:coverage_server` is manual. The installed-service test instead
starts the actual packaged executable with private Redis, generated certificates
and the installed models/UI, then checks TLS and RESTCONF responses.

The four original OpenAPI golden comparisons, CLI client assertions, XML
preparation, TLS URL/CA checks and package contracts run separately. The full
package check compares generated assets and install paths, ownership, modes,
symlinks, architecture and every ELF/debug pair. These source/TAR checks do not
claim Docker integration, PAM/TACACS authentication or VS image validation.

`models/codegen.config` is a declared input. The current four empty settings are
supported; nonempty exclusions/client selections and handwritten OpenAPI specs
fail clearly until their corresponding generator targets are declared.

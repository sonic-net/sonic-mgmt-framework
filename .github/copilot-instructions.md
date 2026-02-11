# Copilot Instructions for sonic-mgmt-framework

## Project Overview

sonic-mgmt-framework provides the SONiC management framework, including the REST server, CLI infrastructure, and OpenAPI/YANG code generation tooling. It serves as the northbound management interface for SONiC switches, offering REST APIs and a KLISH-based CLI that translates user requests into Redis DB operations via the translib layer from `sonic-mgmt-common`. This is the user-facing management layer of the SONiC management stack.

## Architecture

```
sonic-mgmt-framework/
├── rest/                         # REST server (Go)
│   ├── main/                     # REST server entry point
│   │   └── main.go              # Server main()
│   ├── server/                   # HTTP server implementation
│   └── Makefile
├── CLI/                          # KLISH-based CLI framework
│   ├── clitree/                  # CLI command tree definitions (XML)
│   ├── actioner/                 # CLI action handlers (Python scripts)
│   ├── renderer/                 # CLI output rendering templates (J2)
│   ├── clicfg/                   # CLI configuration files
│   ├── klish/                    # KLISH shell framework
│   └── Makefile
├── models/                       # API model definitions
│   ├── openapi/                  # OpenAPI (Swagger) specifications
│   ├── .gitkeep
│   ├── Makefile
│   ├── codegen.config            # Code generation configuration
│   ├── openapi_codegen.mk       # OpenAPI code generation rules
│   └── yang_to_openapi.mk       # YANG-to-OpenAPI conversion rules
├── tools/                        # Code generation and development tools
│   ├── codegen/                  # API server code generator
│   ├── pyang/                    # pyang YANG processing plugins
│   ├── swagger_codegen/          # Swagger code generation
│   ├── restconf_doc_tools/       # RESTCONF documentation generators
│   ├── openapi_tests/            # OpenAPI validation tests
│   ├── ui_gen/                   # UI generation tools
│   └── test/                     # Test utilities
├── debian/                       # Debian packaging
├── Makefile                      # Top-level build orchestration
└── .github/                      # CI workflows
```

### Key Concepts
- **REST server**: Go-based HTTP server that exposes RESTCONF/OpenAPI endpoints and delegates to translib for DB operations
- **KLISH CLI**: XML-defined command tree with Python actioners that call REST APIs or directly interact with SONiC
- **OpenAPI code generation**: YANG models are converted to OpenAPI specs, which then generate Go server stubs
- **Actioner pattern**: CLI commands trigger Python scripts in `CLI/actioner/` that format requests and call REST endpoints
- **Renderer pattern**: CLI output is formatted using Jinja2 templates in `CLI/renderer/`

## Language & Style

- **Primary languages**: Go (REST server), Python (CLI actioners, tools), XML (CLI tree definitions)
- **Secondary**: Jinja2 (renderers), YANG (models), Shell scripts
- **Go conventions**:
  - Standard `gofmt` formatting
  - Types/Functions: `PascalCase` for exported, `camelCase` for unexported
  - File names: `snake_case.go`
- **Python conventions**:
  - PEP 8 compliant
  - 4 spaces indentation
  - CLI actioner scripts are standalone executables
- **XML CLI definitions**: Follow KLISH XML schema for command tree structure
- **Jinja2 templates**: Use `.j2` extension; follow existing template patterns in `CLI/renderer/`

## Build Instructions

```bash
# Full build requires sonic-buildimage environment
# For standalone development:

# Build REST server
cd rest
make

# Build CLI
cd CLI
make

# Build models / code generation
cd models
make

# Build Debian package (in sonic-buildimage context)
BLDENV=stretch make target/debs/stretch/sonic-mgmt-framework_1.0-01_amd64.deb

# Build Docker container
BLDENV=stretch make target/docker-sonic-mgmt-framework.gz

# Incremental rebuild
BLDENV=stretch make target/debs/stretch/sonic-mgmt-framework_1.0-01_amd64.deb-clean
BLDENV=stretch make target/debs/stretch/sonic-mgmt-framework_1.0-01_amd64.deb
```

## Testing

```bash
# REST server tests
cd rest
go test -v ./...

# OpenAPI validation tests
cd tools/openapi_tests
# Run validation against generated specs

# CLI testing is typically done in integration environments
# Manual testing: Load CLI tree in KLISH and verify commands

# Tool tests
cd tools/test
# Run tool-specific tests
```

## PR Guidelines

- **Signed-off-by**: REQUIRED on all commits (`git commit -s`)
- **CLA**: Sign the Linux Foundation EasyCLA
- **Single commit per PR**: Squash commits before merge
- **CLI changes**: Include XML tree definitions, actioner scripts, and renderer templates together
- **REST API changes**: Update OpenAPI specs and verify generated code
- **YANG model changes**: Coordinate with `sonic-mgmt-common` for model updates
- **Reference**: Link SONiC management framework design documents in PR

## Dependencies

- **sonic-mgmt-common**: Provides translib, CVL, and YANG models (core dependency)
- **sonic-swss-common**: Redis DB access (via translib)
- **KLISH**: CLI shell framework (embedded in `CLI/klish/`)
- **pyang**: YANG model processing
- **swagger-codegen**: OpenAPI code generation
- **Go standard library**: HTTP server, JSON, etc.
- **Jinja2**: Python template engine for CLI rendering
- **sonic-buildimage**: Full build environment

## Gotchas

- **Build environment**: Full builds require the `sonic-buildimage` Docker environment; standalone builds only work for individual components
- **Code generation**: Much of the REST server code is auto-generated from OpenAPI specs — don't manually edit generated files
- **YANG ↔ OpenAPI sync**: Changes to YANG models in `sonic-mgmt-common` must be reflected in OpenAPI specs here
- **CLI tree ordering**: KLISH XML tree files have include dependencies; incorrect ordering causes CLI load failures
- **Actioner path**: CLI actioners must be executable and have correct shebang lines
- **Renderer templates**: Jinja2 templates must handle both empty results and error cases
- **Python path**: CLI actioners may need `PYTHONPATH` set to find SONiC Python libraries
- **REST authentication**: The REST server supports certificate-based auth; test with proper certs
- **Incremental builds**: Use `-clean` make targets before rebuilding specific components
- **Model interdependencies**: OpenAPI specs may import types from other specs; verify cross-references

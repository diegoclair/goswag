# Changelog

Notable changes from v1.1.0 onward are documented in this file; for earlier
releases see the git history.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v2.2.0] - 2026-09-23

Regenerating after this release rewrites every route that takes a path
parameter, so expect a large diff on `goswag.go` and on the spec. The API being
described does not change; where it was already written down correctly, nothing
moves.

### Fixed

- Route paths reach the spec in OpenAPI's own syntax. Echo and Gin spell a path
  parameter `:id` and a catch-all `*filepath`, and goswag passed that straight
  into `@Router`, so every parameterised route produced a document no validator
  would accept — on a 124-route API, half the paths. They are now written
  `{id}` / `{filepath}`; a path already in that shape is left alone.
- A path parameter the route declares but no `PathParam` call describes is now
  written out instead of being dropped. OpenAPI rejects a path whose parameters
  are not all described, and a path parameter can only ever be a required string,
  so the only thing that was missing is a name — it takes the one the path uses.

### Added

- `goswag docs --dedupe` factors the responses and parameters that repeat across
  operations into the reusable objects OpenAPI 2.0 defines (`#/responses`,
  `#/parameters`) and points each operation at them. swag restates both inline on
  every operation, so a default response set is copied once per route; on a
  124-route API this took `swagger.json` from 16,770 to 13,025 lines without
  changing the API it describes. Off by default. It reads `swagger.json` and
  rewrites `swagger.yaml` and `docs.go` from it, so the outputs stay in agreement,
  and running it twice over one spec changes nothing the second time.

## [v2.1.0] - 2026-08-30

### Added

- `goswag docs --output-types` forwards swag's `--ot`. Defaults to `go,json,yaml`
  (swag's own default), so generation is unchanged unless you pass it. Use
  `--output-types json,yaml` to skip `docs.go` and keep `github.com/swaggo/swag`
  out of your `go.mod`. Values other than `go`, `json`, `yaml` and `yml` are
  rejected before swag runs.

## [v2.0.1] - 2026-08-25

### Fixed

- Generated `goswag.go` is stable under `gofmt`: the `//nolint:unused` stub line
  no longer carries a trailing space, and the `goswag` CLI runs `gofmt` after
  `swag fmt`, so saving the file in an editor no longer produces a diff.

## [v2.0.0] - 2026-08-25

### Changed

- **BREAKING** Module path is now `github.com/diegoclair/goswag/v2`. Update imports
  and run `go get github.com/diegoclair/goswag/v2`.
- **BREAKING** Echo support targets Echo v5 (`github.com/labstack/echo/v5`).
  Handlers take `*echo.Context` instead of `echo.Context`.
- Gin updated to v1.12.0. Public Gin API of goswag is unchanged.

## [v1.3.0]

### Added

- Fluent MCP builder: `NewMCP(...).Tool(...)/.AddTools(...).Build()`, with `Tool`
  as a generic method so In/Out are inferred per call.

### Changed

- Minimum Go version is 1.27.0.

## [v1.2.2]

### Fixed

- Generated annotations disambiguate types that share a short name across
  packages, so a project with two same-named packages no longer gets half its
  spec pointing at the other package's types.

### Changed

- Generator writes into a string builder instead of printing incrementally.
- Minimum Go version is 1.26.6.

## [v1.2.1]

### Fixed

- Generator imports packages named inside generics and containers.

## [v1.2.0]

### Added

- Runtime MCP server builder: `Tool(handler, ...)` plus `NewMCPServer()` returning
  a mountable `*mcp.Server`, backed by `modelcontextprotocol/go-sdk`.

### Changed

- Minimum Go version is 1.26.5, plus dependency security bumps.

## [v1.1.0]

### Fixed

- Handler names derived from the route handler are disambiguated with a short
  hash of the package qualifier, so handlers sharing a short name across
  packages no longer collide in the generated `goswag.go`.

[v2.2.0]: https://github.com/diegoclair/goswag/releases/tag/v2.2.0
[v2.1.0]: https://github.com/diegoclair/goswag/releases/tag/v2.1.0
[v2.0.1]: https://github.com/diegoclair/goswag/releases/tag/v2.0.1
[v2.0.0]: https://github.com/diegoclair/goswag/releases/tag/v2.0.0
[v1.3.0]: https://github.com/diegoclair/goswag/releases/tag/v1.3.0
[v1.2.2]: https://github.com/diegoclair/goswag/releases/tag/v1.2.2
[v1.2.1]: https://github.com/diegoclair/goswag/releases/tag/v1.2.1
[v1.2.0]: https://github.com/diegoclair/goswag/releases/tag/v1.2.0
[v1.1.0]: https://github.com/diegoclair/goswag/releases/tag/v1.1.0

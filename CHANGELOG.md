# Changelog

Notable changes from v1.1.0 onward are documented in this file; for earlier
releases see the git history.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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

[v2.0.0]: https://github.com/diegoclair/goswag/releases/tag/v2.0.0
[v1.3.0]: https://github.com/diegoclair/goswag/releases/tag/v1.3.0
[v1.2.2]: https://github.com/diegoclair/goswag/releases/tag/v1.2.2
[v1.2.1]: https://github.com/diegoclair/goswag/releases/tag/v1.2.1
[v1.2.0]: https://github.com/diegoclair/goswag/releases/tag/v1.2.0
[v1.1.0]: https://github.com/diegoclair/goswag/releases/tag/v1.1.0

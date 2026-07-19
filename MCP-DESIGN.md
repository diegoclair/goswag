# MCP Generation — Design Proposal

> **Status:** Draft — design direction revised 2026-06-02 (see §0). To be validated during the Phase 1 spike.
> **Target release:** v2.0.0.
> **Last updated:** 2026-06-02.

This document proposes adding Model Context Protocol (MCP) server generation
to **goswag**. The goal is letting the *same* route declarations that already
produce a Swagger spec **also** expose those routes as MCP tools that LLMs
can invoke — without duplicating definitions and without surprising the user
with magic.

This doc is a contract for the implementation. The open architectural choices
were revisited against a real embedding scenario; the resulting direction is
in §0 and is to be validated during implementation.

---

## 0. Design direction (revised 2026-06-02 — to be validated in Phase 1)

The pre-analysis below (§2.1 onward) leaned **codegen-only**. Revisiting it
against a real scenario — a host that wants to expose its domain operations as
agent tools **embedded in its existing HTTP server** — flipped the primary
surface. Treat §0 as authoritative where it conflicts with later sections.

- **Runtime builder first (hybrid), not codegen-only.** goswag already
  collects routes at runtime (`GenerateSwagger()` runs inside the user's
  `main.go`) and the route handler is already a func value in memory. So the
  primary surface is a **runtime builder**: `.AsTool(handler, opts...)` marks a
  route + stores its pure handler, and a new `NewMCPServer(opts...)` walks the
  collected routes, reflects each input struct → JSON Schema, wires the handler
  + `PreInvoke`/`PostInvoke` + safety, and returns an `*mcp.Server` **and an
  `http.Handler`**. This reverses the "code generation, not runtime"
  non-negotiable of §2.1 — justified because (a) wiring a func value at runtime
  is robust where codegen of a func reference (closures, methods) is fragile,
  and (b) a host that mounts the MCP server **embedded in its existing HTTP
  server**, sharing DI, finds a separate generated binary awkward. Codegen of a
  standalone binary stays a v2.x *option*, not the default.
- **Embedded Streamable HTTP is the priority transport** (NOT plain SSE — the
  2024 HTTP+SSE transport is deprecated/superseded by Streamable HTTP). With the
  official Go SDK, `NewMCPServer()` builds an `*mcp.Server`; the host gets a
  `http.Handler` via the SDK's `mcp.NewStreamableHTTPHandler(getServer, opts)`
  and mounts it on its existing router (e.g. echo at `/mcp`), reusing the host's
  services, DB and request-context machinery. `ServeStdio` (and a standalone
  Streamable-HTTP runner) remain offered (§5.1) for non-embedded users. The
  embedded handler MUST follow the HTTP-transport security baseline (§5.4).
- **`PreInvoke` / `PostInvoke` are generic, content-agnostic hooks** — goswag
  takes NO position on what they do. `PreInvoke(ctx, toolName, *input) error`
  is an arbitrary pre-invocation gate (return non-nil to reject); `PostInvoke`
  is an arbitrary post-invocation observer. Auth, rate limiting, tenant scoping,
  audit logging are all the **consumer's** concern, wired through these hooks —
  none of that vocabulary belongs in the lib. goswag's only opinions: it never
  bundles auth, and auth-style headers (`Authorization`/`X-Token`/…) are
  **omitted from the tool inputSchema** (§4.3) so the LLM never sees them; the
  hook reads them from the request context. (For a worked multi-tenant
  token-auth example, see the consumer cookbook, not this design doc.)
- **The host composes multi-server configs; goswag owns one server.** A host's
  agent may need several MCP servers of different transports (the goswag
  Streamable-HTTP one + external stdio-only servers like a filesystem/Slack/
  GitHub MCP). goswag generates one server; the host aggregates the client-side
  config. goswag is not a client or an aggregator.
- **Curated tool surface over auto-exposing all CRUD.** `.AsTool()` stays
  strictly opt-in per route (§2.3); hosts are expected to expose a small,
  agent-ergonomic set (good ACI) rather than mirroring every endpoint.
- **Targets:** official SDK `github.com/modelcontextprotocol/go-sdk` (pin v1.x,
  ≥ the version that supports Streamable HTTP); spec revision **2025-11-25**.
  Design **stateless-forward** for the 2026-07-28 RC — do NOT build
  `Mcp-Session-Id` session-store machinery the next spec removes.

---

## 1. Why this matters

Conversational interfaces over existing REST APIs are a fast-growing
real-world need. Today, exposing a Go API to an LLM via MCP requires a
hand-written MCP server that mirrors every endpoint. Definitions drift,
schemas get out of sync, and security defaults are improvised per project.

goswag already owns the structured definition of every route — method,
path, parameters, body type, return type, description. Generating MCP
tools from that same source is a natural fit, and avoids the duplication
that motivated goswag in the first place (vs. raw `swaggo/swag` comments).

---

## 2. Design principles

These are non-negotiable and should drive every decision below.

### 2.1 Code generation, not runtime — ⚠️ REVISED, see §0
> **Superseded by §0 (2026-06-02): runtime builder is now the primary surface;
> codegen is a v2.x option.** The original reasoning is kept below for history.

goswag stays a **generator**. It emits a `mcp_server.go` file the user
compiles and ships. No hidden runtime, no goroutines started behind the
scenes, no MCP server embedded in the lib. The user reads the generated
file, debugs it, audits it. Consistent with how Swagger generation works
today.

### 2.2 Explicit handlers (no automatic context fakes)
Each MCP-eligible route requires the user to pass a **pure function** —
no `echo.Context` / `*gin.Context`. This:

- Decouples the MCP handler from the HTTP framework
- Makes auth and validation responsibilities explicit
- Avoids the entire class of bugs caused by synthetic HTTP contexts
- Forces (in a healthy way) separation of transport and business logic

An "automatic" mode that reuses the HTTP handler via synthetic context
was considered and rejected. It looks convenient but accumulates edge
cases (middleware that reads from `request.RemoteAddr`, response
streaming, multipart forms…). The explicit handler is one extra line for
the user and removes a whole category of failure modes.

### 2.3 Safe by default
MCP tools that mutate state are dangerous when wired to an autonomous
LLM. The default behaviour must be cautious:

- **Read-only by default.** `.AsTool()` on a `GET` is allowed implicitly;
  on `POST/PUT/PATCH/DELETE` requires `.AsTool(goswag.AllowWrite())` or
  similar opt-in.
- **No tools without `.AsTool()`.** Adding goswag to a project that
  already uses goswag for Swagger must not retroactively expose every
  route as an MCP tool.
- **No `interface{}` / `any` inputs.** If the input type cannot be
  reduced to a concrete JSON Schema, generation fails loudly. We do not
  emit a tool the LLM cannot reason about.

### 2.4 JSON Schema from Go types
Tool input schemas come from reflecting the input struct using a
well-maintained library (e.g. `invopop/jsonschema`). goswag does **not**
re-implement Go-to-JSON-Schema conversion — that is a deep, lived-in
problem solved better elsewhere.

### 2.5 Auth, rate limiting, and observability are the user's job
goswag does not embed auth. It produces a server with a clear extension
point (middleware hook on `NewMCPServer(...)`), and the project plugs in
whatever it already uses for the REST API. Same model as `gqlgen`.

---

## 3. Proposed public API

### 3.1 Marking a route as an MCP tool

```go
ge.GET("/orders/:id", h.GetOrder).
    Summary("Get order by ID").
    PathParam("id", true).
    Returns([]goswag.ReturnType{{StatusCode: 200, Body: Order{}}}).
    AsTool(func(ctx context.Context, in GetOrderInput) (Order, error) {
        return orderService.Get(ctx, in.ID)
    })
```

For writes, the user must explicitly opt in:

```go
ge.POST("/orders", h.CreateOrder).
    Summary("Create a new order").
    Read(CreateOrderInput{}).
    AsTool(
        func(ctx context.Context, in CreateOrderInput) (Order, error) {
            return orderService.Create(ctx, in)
        },
        goswag.AllowWrite(),
    )
```

### 3.2 Input struct conventions

A single struct unifies everything the LLM sends:

```go
type GetOrderInput struct {
    ID string `path:"id" jsonschema:"required,description=The order ID"`
}

type CreateOrderInput struct {
    OrgID    string         `path:"orgId" jsonschema:"required"`
    Page     int            `query:"page"`
    AuthTok  string         `header:"X-Token" jsonschema:"required"`
    Body     CreateOrderDTO `body:"" jsonschema:"required"`
}
```

Why a single struct (not multiple positional args)?

- Maps 1:1 to a single JSON Schema (MCP's contract).
- The LLM sees one input shape, not "where does this field go".
- Tags reuse familiar Go conventions (`query:`, `path:`, `header:`,
  `body:`).

### 3.3 What the builder produces (runtime)

In the runtime-first model (§0), `NewMCPServer(opts...)` walks the
`.AsTool()`-marked routes and registers each with the official SDK's
**generic, type-safe** registration. `mcp.AddTool[In, Out]` derives the JSON
Schema from the input struct **and validates incoming arguments against it
before the handler runs** — so there is no hand-rolled `json.RawMessage`
unmarshal, and the input-validation work of §4.8 is mostly inherited from the
SDK rather than re-implemented:

```go
import "github.com/modelcontextprotocol/go-sdk/mcp"

// built in-process; the host mounts the result (see §5.1), not a generated file.
func NewMCPServer(opts ...goswag.MCPOption) *mcp.Server {
    cfg := goswag.NewMCPConfig(opts...) // PreInvoke/PostInvoke, size caps, hidden-field policy

    s := mcp.NewServer(&mcp.Implementation{Name: "your-api", Version: cfg.Version}, nil)

    // one registration per .AsTool() route — typed handler, schema auto-derived + auto-validated:
    mcp.AddTool(s, &mcp.Tool{
        Name:        "get_order_by_id",
        Description: "Get an order by its ID.",
    }, func(ctx context.Context, req *mcp.CallToolRequest, in GetOrderInput) (*mcp.CallToolResult, Order, error) {
        if err := cfg.PreInvoke(ctx, "get_order_by_id", &in); err != nil {
            return nil, Order{}, err // generic gate: auth / rate-limit / etc. (consumer-defined)
        }
        out, err := userHandlers.getOrderByID(ctx, in) // the pure func passed to .AsTool()
        cfg.PostInvoke(ctx, "get_order_by_id", &in, out, err)
        return nil, out, err
    })

    return s
}
```

> Exact SDK signatures track the pinned SDK version (§0 targets); the shape
> above is the generic `AddTool[In, Out]` path, not the lower-level
> `server.AddTool(json.RawMessage)` one (which would discard the free schema
> validation). `PreInvoke`/`PostInvoke` are the generic hooks from §0 — goswag
> does not define their contents.

### 3.4 CLI (optional, v2.x — standalone codegen)

The runtime builder above needs no CLI. For users who want a **standalone**
MCP binary (not embedded), a later v2.x adds codegen that emits an equivalent
`mcp_server.go`:

```sh
goswag mcp generate         # optional: writes ./goswag/mcp_server.go for a standalone binary
goswag mcp generate -o ./mcp
```

---

## 4. Security considerations

This section is the most important part of this document. As an
open-source library, anything we ship gets dropped into other people's
production systems. The defaults need to be conservative.

### 4.1 Write operations must be opt-in
`.AsTool()` on a write method without `goswag.AllowWrite()` is a
**generation error**, not a warning. Forces the user to make a conscious
decision.

### 4.2 Sensitive fields must be markable
Provide a struct tag to keep fields out of the LLM-facing schema:

```go
type CreateOrderInput struct {
    CustomerID string `json:"customer_id" jsonschema:"required"`
    InternalID string `goswag:"mcp:hidden"` // not exposed to LLM
}
```

Generated code rejects requests where hidden fields are populated by
the LLM. This protects against an LLM filling internal-only fields it
shouldn't know about.

### 4.3 No leaking auth tokens to the schema
Header parameters used for authentication must not be exposed in the
tool's `inputSchema`. If the project's REST API takes `Authorization:
Bearer …`, the MCP server gets the token from its own auth middleware
(via `cfg.PreInvoke`), not from the LLM. Default behaviour: any header
named `Authorization`, `Cookie`, `X-API-Key`, `X-Token`, or matching a
configurable list is **omitted from the schema** and pulled from
context instead.

### 4.4 Response filtering
Some response fields shouldn't reach the LLM (PII, audit metadata,
internal IDs). Same tag mechanism:

```go
type Order struct {
    ID         string `json:"id"`
    CustomerID string `json:"customer_id"`
    AdminNotes string `json:"admin_notes,omitempty" goswag:"mcp:hidden"`
}
```

Generated code strips hidden fields before returning to the LLM.

### 4.5 Payload size limits
Large tool responses blow the client's context. Express the cap in the
**client's terms**, not a lib-invented byte size: Claude Code warns around
~10k tokens and caps at ~25k (`MAX_MCP_OUTPUT_TOKENS`). goswag exposes a
configurable cap (token-oriented, byte fallback) that truncates with a warning
to the LLM. Stops the "tool returns 50k rows and blows the context" failure mode.

### 4.6 Rate limiting hook
`cfg.PreInvoke` is called before every tool invocation with the tool
name and decoded input. Projects implement rate limiting here, scoped
however they want (per user, per tool, per session).

### 4.7 Audit log hook
A second hook, `cfg.PostInvoke(ctx, tool, in, out, err)`, gives
projects a place to write structured audit logs. Especially important
for write tools, where regulatory requirements (SOC2, GDPR) typically
require traceable decisions.

### 4.8 Untrusted input is parsed defensively
- JSON decoding uses `DisallowUnknownFields` so the LLM cannot smuggle
  extra fields that bypass schema validation.
- Numeric fields are range-checked against their Go types (no
  `int64(1e18)` into an `int32`).
- String fields with `maxLength` in the schema are enforced server-side
  too (schema is documentation, not enforcement).

### 4.9 No filesystem / process side effects in generated code
The generated MCP server never reads from disk, never shells out,
never reaches outside the handlers the user provided. Reviewers can
audit the generated file in seconds.

---

## 5. Operations & lifecycle

### 5.1 Transport
Two transports per the current spec — **plain SSE is deprecated, do not offer
it**. goswag hands you the `*mcp.Server`; you use the SDK's entrypoints:

```go
// stdio — local subprocess MCP (a client spawns it)
goswag.ServeStdio(server)

// Streamable HTTP — SDK returns an http.Handler you mount on your own mux
h := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return server }, nil)
echoMux.Any("/mcp", echo.WrapHandler(h)) // embedded (recommended, §5.2)
```

### 5.2 Embedded vs standalone
**Recommended: embedded.** Mount the Streamable-HTTP handler on the host's
existing router so the tools share the app's DI, services, DB, and request
context — the reason §0 picks the runtime builder. A **standalone** binary (its
own listener; the v2.x codegen path) is for hosts wanting process isolation; it
then owns its own auth/observability:

```go
func main() {
    srv := NewMCPServer(goswag.WithPreInvoke(gate), goswag.WithPostInvoke(audit))
    http.ListenAndServe("127.0.0.1:9000", mcp.NewStreamableHTTPHandler(
        func(*http.Request) *mcp.Server { return srv }, nil))
}
```

### 5.3 Hot reload / dev loop
The runtime builder rebuilds with the app — no separate step. (The optional
v2.x `goswag mcp generate` codegen is idempotent; wire it into `go generate`.)

### 5.4 HTTP-transport security baseline (MUST)
Independent of goswag, a Streamable-HTTP MCP server MUST:
- **Validate the `Origin` header** on every request (DNS-rebinding defense).
- **Bind to `127.0.0.1`** for local servers (not `0.0.0.0`). An embedded handler
  sharing a host bound to `0.0.0.0` MUST auth-gate the `/mcp` route.
- **Handle `MCP-Protocol-Version`** — reject unsupported/missing with 400.

Document these so consumers don't skip them; goswag can offer a wrapping
middleware that enforces Origin + protocol-version by default.

---

## 6. What goswag does *not* do

Saying "no" up front avoids scope creep later.

- **No bundled auth.** Hooks are exposed; the user wires them.
- **No MCP client.** Only server generation.
- **No vector search, retrieval, embeddings.** Out of scope.
- **No multi-API aggregation.** One project, one MCP server.
- **No prompt templates / resource URIs (yet).** v2 focuses on tools.
  Resources and prompts can come in v2.x.
- **No automatic handler reuse.** As argued in §2.2, the user provides a
  pure function. We do *not* synthesize an `echo.Context` to call the
  HTTP handler.
- **No partial / streaming tool responses (yet).** MCP supports this;
  the first cut returns full responses. Streaming is a v2.x addition.

---

## 7. Open questions

These are not yet resolved. Feedback welcome.

### 7.1 Tool naming
`Summary("Get order by ID")` → `get_order_by_id`? Or require an
explicit `.ToolName("get_order")`? Auto-derivation is convenient but
silently rebrands a tool when the summary changes. Explicit is safer.
**Lean:** require explicit `.ToolName(...)`, fall back to a slugged
summary with a warning logged at generation time.

### 7.2 Versioning of the MCP server
The MCP protocol carries a server version. Should it default to the
goswag library version, the user's module version, or be required as a
flag? **Lean:** required flag in `NewMCPServer(opts...)`.

### 7.3 Error mapping
When the user's handler returns an error, how is it surfaced to the
LLM? The full error message risks leaking internals. **Lean:** by
default, only the error type is exposed; the user can opt into
`goswag.WithVerboseErrors()` for development.

### 7.4 Input schema customization
`invopop/jsonschema` supports tag-driven customization but is not the
only option. If we hit a wall (recursive types, generics), we may want
to allow `.WithCustomSchema(json.RawMessage)` as an escape hatch per
tool. **Lean:** ship without this, add only if a real use case demands.

---

## 8. Implementation roadmap

### Phase 1 — Spike (1 week)
Goal: prove end-to-end on a single route. Throwaway code allowed.

- Add `.AsTool(handler)` to one framework (echo) only.
- Reflect input struct → JSON Schema using `invopop/jsonschema`.
- Generate `mcp_server.go` for one hand-picked route.
- Compile, run with stdio transport, call from Claude Desktop, see
  result.
- Write findings in this doc's §10 (lessons learned).

### Phase 2 — Core public API (2-3 weeks)
- `.AsTool()` on both echo and gin, GET only.
- `goswag mcp generate` CLI command.
- Tag conventions (`path:`, `query:`, `header:`, `body:`, `goswag:`).
- `cfg.PreInvoke` / `PostInvoke` hooks.
- Documentation: README section + a `examples/mcp/` directory.

### Phase 3 — Write operations (1-2 weeks)
- `goswag.AllowWrite()` option.
- Sensitive field filtering (`goswag:"mcp:hidden"`).
- Auth-header omission defaults.
- Response size limits.

### Phase 4 — Hardening (ongoing)
- Audit log hook.
- Verbose error mode.
- Streaming (if/when demand appears).
- Resources & prompts primitives.

### Phase 5 — v2.0.0 release
After Phase 3 is battle-tested in at least one real project, cut a
release. The library version bumps because v2 is a meaningful surface
addition (not a breaking change to existing Swagger generation).

---

## 9. Backwards compatibility

Every change proposed here is additive. Projects already using goswag
for Swagger generation see no behavioural change unless they call
`.AsTool()` or run `goswag mcp generate`. v1.x code continues to work
under v2.x without modification.

The only "soft" surface change is the new struct tags
(`goswag:"mcp:hidden"`). They are ignored by Swagger generation, so
adding them does not affect existing users.

---

## 10. Lessons learned (to be filled during Phase 1)

*Populated during the spike. Captures what surprised us, what we got
wrong in this design, and what we changed.*

---

## 11. References

- Model Context Protocol specification:
  https://modelcontextprotocol.io
- Anthropic Go SDK (preferred): https://github.com/modelcontextprotocol/go-sdk
- `invopop/jsonschema` for Go → JSON Schema reflection.
- `gqlgen` as the design analogue for "code-gen, auth-is-yours".

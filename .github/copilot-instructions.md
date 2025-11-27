## TypedHandler — current status & developer notes

This file documents the real, up-to-date status of the `typedhandler` package in this
repository. `typedhandler` is a compact, high-performance Go HTTP request parser and
handler generator that composes a request parser, optional instance pooling, business
logic (service function) and consistent response/error writing into an `http.HandlerFunc`.

The implementation uses Go generics and reflection (metadata is cached) and ships with
examples and tests. The code is intentionally small and focuses on clarity and
low-allocation paths; performance-sensitive JSON parsing can optionally be swapped in
(see `examples/sample/sample_easyjson.go`).

Core files and responsibilities
- `typedhandler/handler.go` — handler creation (`CreateHandler`, `CreateSimpleHandler`),
    response and error writing logic.
- `typedhandler/parser.go` — request parsing entrypoints and orchestration.
- `typedhandler/schema_helper.go`, `parser_infos.go` (if present) — reflection-based
    metadata extraction for struct tags and body parsing strategies.
- `typedhandler/data_conversion.go` — conversion helpers for path/query/header values
    (strings → ints, bools, time.Time, etc).
- `examples/` — runnable examples (`examples/simple/main.go` shows a minimal server).
- Tests and benchmarks are under `typedhandler/` (e.g. `handler_test.go`,
    `handler_create_handler_test.go`).

Quick usage summary
- Create a parser for your request struct and then create a handler.
  See examples/simple/main.go for concrete usage with CreateParser and CreateHandler.
- Alternatively, use CreateSimpleHandler, which is a convenience wrapper that builds
  the parser for you internally.

Service function contract
- Signature: func(ctx context.Context, req RIn) (ROut, int, error)
- The returned int is the HTTP status code; if <=0 it defaults to 200 OK.

Key conventions & constraints
- Request schema types are expected to be pointer types (e.g. `*MyRequest`). The
    parser and some helpers will panic if a non-pointer type is supplied.
- Struct tags supported: `path`, `query`, `header`, and `json` (for body fields).
- Header-tagged fields must be `string` types (the parser will panic on invalid
    header field types during initialization).
- The library does not perform validation (the `validate` tag may be recognised but
    is not enforced). Add your own validation in the service function or integrate a
    validator in your project.

Pooling and lifecycle
- The package provides optional instance pooling (controlled by the global
    `typedhandler.PoolEnabled` flag). When pooling is enabled, request instances are
    reused via `sync.Pool` to reduce allocations.
- Implement `Resettable` on request types to customise reset behavior. The package
    exposes `CreatedInstances()` and `ResetCreatedInstances()` for bookkeeping in
    benchmarks/tests.

Body parsing modes
- JsonBody (default): entire body is unmarshaled into fields tagged with `json:`.
- NoBody: skip body parsing (useful for GET/DELETE endpoints).
- JsonField: unmarshal only into a specific field. Implement `GetBodyField() any`
    on the request type to return a pointer to the target field.

Error handling
- Support for custom error types: implement `HttpError` (Status() int) or
    `HttpJsonError` (Status() int, Json() []byte) to control status code and body.

Development & testing
- Run tests: `make test` (Makefile delegates to `go test` for the module).
- Benchmarks: `make bench` (writes output to `benchmarks/benchmark.txt`).
- Linting: `make deps` will install development tools; CI is configured to run
    `golangci-lint` in the project.

Compatibility
- The project uses Go generics and targets a recent Go version (see `go.mod`).
    The repository has been developed against Go 1.25+ but will likely work with
    other recent releases that support generics.

Examples & tests to check for usage
- `examples/simple/main.go` — demonstrates `CreateParser` + `CreateHandler` and
    a working HTTP server.
- `README.md` — quick-start snippets and more guidance.
- `typedhandler/handler_test.go`, `typedhandler/handler_create_handler_test.go` —
    unit tests and benchmarks that exercise the public handler creation APIs.

Notes and TODOs
- The library focuses on parsing and handler composition; adding a small built-in
    validation hook (pluggable validator) could be a useful enhancement.
- Consider documenting the expected route syntax for path parameters (the project
    assumes a router that supports `{param}` style path parameters when used with
    `http.HandleFunc` from Go 1.22+).

If you need the copilot instructions to be even more condensed or to include
direct links to specific test function names, tell me which targets to reference
and I'll update the file accordingly.

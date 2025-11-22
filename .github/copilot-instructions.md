# Router - HTTP Handler Helper

This is a high-performance Go HTTP request parser and handler generator that uses generics, reflection, and object pooling to minimize allocations.

## Architecture Overview

### Core Components

**Request Flow**: HTTP Request → Parser → Instance Pool → Service Function → Response Writer
- `router.go`: Main handler creation using generics (`CreateHandler`, `CreateSimpleHandler`)
- `parser.go`: Request parsing logic extracting data from path/query/headers/body into structs
- `instance_pool.go`: sync.Pool-based memory reuse to reduce allocations (toggle with `PoolEnabled`)
- `parser_infos.go`: Reflection-based metadata extraction from struct tags at initialization
- `data_conversion.go`: Type conversion for path/query/header values (string → int/bool/time.Time/etc)

### Key Design Patterns

**Generic Handler Creation**: Handlers are created with type parameters for request and response schemas:
```go
CreateHandler[RIn RequestSchema, ROut ResponseSchema](
    parseRequestFunc ParseRequestFunc[RIn],
    serviceFunc ServiceFunc[RIn, ROut],
) HandlerFunc
```

**Struct Tag-Based Parsing**: Request structs use tags to define data sources:
```go
type LoginRequest struct {
    UserName string `json:"username"`           // from body
    City     string `query:"city"`              // from query params
    State    string `header:"state"`            // from headers
    Country  string `path:"country"`            // from path params
}
```

**Object Pooling**: `InstancePool` reuses request struct instances via `sync.Pool`:
- Controlled by global `PoolEnabled` flag
- Custom reset via `Resettable` interface or auto-generated reset functions
- Track allocations with `CreatedInstances()` / `ResetCreatedInstances()`

**Body Parsing Strategies** (see `parser_infos.go`):
- `NoBody`: No body parsing
- `JsonBody`: Unmarshal entire body into struct (fields with `json:` tags)
- `JsonField`: Unmarshal body into specific field via `BodyFieldGetter` interface

### Error Handling

Error responses support custom status codes via interfaces:
- `HttpError`: Custom status code (`Status() int`)
- `HttpJsonError`: Custom status + JSON body (`Status() int`, `Json() []byte`)

## Development Workflows

**Run Tests**:
```bash
make test                  # Run all tests
make coverage              # Run with coverage report (requires go-test-coverage)
```

**Benchmarking**:
```bash
make bench                 # Run benchmarks, save to benchmarks/benchmark.txt
make bench-stat            # Compare new benchmarks against previous run using benchstat
```

**Dependencies**:
```bash
make deps                  # Install all dev tools (pre-commit, golangci-lint, etc)
```

**Key Dependencies**:
- `github.com/mailru/easyjson`: Optional for performance-critical JSON unmarshaling (see `sample_easyjson.go`)
- `github.com/stretchr/testify`: Testing assertions
- Go 1.25.3+ (uses generics heavily)

## Project-Specific Conventions

**Pointer vs Value Types**: Request schemas MUST be pointer types:
```go
// Correct
CreateParser[*LoginRequest]()

// Will panic
CreateParser[LoginRequest]()  // Error: "CreateHandler must receive *LoginRequest"
```

**Parser Initialization**: `getParserInfos[T]()` uses reflection once per type, then caches metadata. Panics early if struct tags are invalid.

**Testing Style**:
- Use `t.Parallel()` for all tests
- Test helper types defined inline (see `router_test.go` for `httpError`, `jsonError`)
- Benchmark both pool-enabled and pool-disabled scenarios

**Reset Pattern**: Implement `Resettable` interface for custom cleanup:
```go
func (r *Request) Reset() {
    r.Name = ""
    r.Age = 0
    // ...
}
```

**Type Conversion**: Supports string → `int/uint/float/bool/time.Time/time.Duration` for path/query/header params (see `data_conversion.go`)

## Integration Points

**Example Usage** (`examples/simple/main.go`):
```go
requestParser := router.CreateParser[*LoginRequest]()
handler := router.CreateHandler(requestParser, serviceFunc)
http.HandleFunc("POST /login", handler)
```

**Service Function Signature**:
```go
func(ctx context.Context, request RIn) (response ROut, status int, err error)
```

## Critical Notes

- **No validation built-in**: The `validate` tag is detected but not enforced—integrate your own validator
- **Header fields must be strings**: Parser will panic during initialization if header-tagged fields aren't strings
- **Body field interface**: For `JsonField` body type, implement `GetBodyField() any` returning a pointer to the field
- **Benchmark files**: `benchmarks/` contains historical benchmark data for comparison

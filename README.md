# TypedHandler

[![golangci lint](https://github.com/guionardo/typedhandler/actions/workflows/golangci.yml/badge.svg)](https://github.com/guionardo/typedhandler/actions/workflows/golangci.yml)
![coverage](https://raw.githubusercontent.com/guionardo/typehandler/badges/.badges/main/coverage.svg)

A high-performance, zero-allocation HTTP request parser for Go that automatically maps HTTP requests to strongly-typed structs using generics and object pooling.

## Features

- 🚀 **Zero Allocations**: Uses `sync.Pool` for request struct reuse
- 🎯 **Type-Safe**: Leverages Go generics for compile-time type checking
- 🏷️ **Struct Tag-Based**: Parse path params, query strings, headers, and body with simple tags
- ⚡ **High Performance**: Reflection done once at initialization, cached for reuse
- 🔧 **Flexible**: Supports custom error types, body parsing strategies, and reset patterns

## Quick Start

```go
package main

import (
    "context"
    "net/http"

    "github.com/guionardo/typedhandler"
)

// Define your request schema
type LoginRequest struct {
    UserName string `json:"username"`      // from body
    City     string `query:"city"`         // from query params
    State    string `header:"X-State"`     // from headers
    Country  string `path:"country"`       // from path params
}

type LoginResponse struct {
    Token string `json:"token"`
}

// Implement your business logic
func loginService(ctx context.Context, req *LoginRequest) (*LoginResponse, int, error) {
    // Your logic here
    return &LoginResponse{Token: "abc123"}, http.StatusOK, nil
}

func main() {
    // Create parser and handler
    parser := typedhandler.CreateParser[*LoginRequest]()
    handler := typedhandler.CreateHandler(parser, loginService)

    // Register with http.ServeMux (Go 1.22+ routing)
    http.HandleFunc("POST /login/{country}", handler)
    http.ListenAndServe(":8080", nil)
}
```

## Installation

```bash
go get github.com/guionardo/typedhandler
```

## Struct Tags

TypedHandler supports multiple data sources via struct tags:

| Tag      | Source              | Example                       |
| -------- | ------------------- | ----------------------------- |
| `path`   | URL path parameters | `{id}` in route `/users/{id}` |
| `query`  | Query string        | `?page=1&limit=10`            |
| `header` | HTTP headers        | `X-API-Key`, `Authorization`  |
| `json`   | JSON request body   | `{"username": "john"}`        |

### Supported Types

Path, query, and header parameters support automatic conversion to:
- `string`
- `int`, `int8`, `int16`, `int32`, `int64`
- `uint`, `uint8`, `uint16`, `uint32`, `uint64`
- `float32`, `float64`
- `bool`
- `time.Time` ([multiple formats](#timetime-parsing))
- `time.Duration`

### time.Time parsing

By default, the parser will use a set of layouts from the standard lib:

| Layout      | Template                              |
| ----------- | ------------------------------------- |
| DateTime    | "2006-01-02 15:04:05"                 |
| RFC3339     | "2006-01-02T15:04:05Z07:00"           |
| RFC3339Nano | "2006-01-02T15:04:05.999999999Z07:00" |
| RFC1123     | "Mon, 02 Jan 2006 15:04:05 MST"       |
| RFC1123Z    | "Mon, 02 Jan 2006 15:04:05 -0700"     |
| ANSIC       | "Mon Jan _2 15:04:05 2006"            |
| DateOnly    | "2006-01-02"                          |
| TimeOnly    | "15:04:05"                            |

The parser will priorize the layouts with successful parsing.

If you need to use another formats, call the `typedhandler.SetTimeLayouts` func

## Body Parsing Strategies

TypedHandler supports three body parsing modes:

### 1. JsonBody (Default)
Unmarshals entire request body into struct fields with `json:` tags:

```go
type Request struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}
```

### 2. NoBody
Skips body parsing entirely (for GET, DELETE, etc.):

```go
type Request struct {
    ID string `path:"id"`
    // No body fields
}
```

### 3. JsonField
Unmarshals body into a specific field. Implement `BodyFieldGetter`:

```go
type Request struct {
    ID   string `path:"id"`
    Data MyData
}

func (r *Request) GetBodyField() any {
    return &r.Data
}
```

## Object Pooling

Enable/disable pooling globally:

```go
typedhandler.PoolEnabled = true  // default: true
```

Track allocations:

```go
count := typedhandler.CreatedInstances()
typedhandler.ResetCreatedInstances()
```

### Custom Reset Logic

Implement the `Resettable` interface for custom cleanup:

```go
type Request struct {
    Name  string
    Items []string
}

func (r *Request) Reset() {
    r.Name = ""
    r.Items = r.Items[:0] // reuse slice capacity
}
```

## Error Handling

Return custom HTTP status codes via interfaces:

```go
type MyError struct {
    message string
}

func (e MyError) Error() string { return e.message }
func (e MyError) Status() int   { return http.StatusBadRequest }
```

For JSON error responses:

```go
type MyJSONError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
}

func (e MyJSONError) Error() string { return e.Message }
func (e MyJSONError) Status() int   { return http.StatusBadRequest }
func (e MyJSONError) Json() []byte {
    b, _ := json.Marshal(e)
    return b
}
```

## Performance

TypedHandler is designed for high-throughput APIs:

- **Zero allocations** per request when pooling is enabled
- **Reflection cached** at parser creation time
- **Optimized JSON parsing** with optional [easyjson](https://github.com/mailru/easyjson) support

Run benchmarks:

```bash
make bench           # Run benchmarks
make bench-stat      # Compare with previous run
```

## Examples

See [examples/simple/main.go](examples/simple/main.go) for a complete working example.

## Requirements

- Go 1.22+ (for enhanced routing patterns)
- Go 1.23.3+ recommended (for latest generics features)

## License

MIT License - see LICENSE file for details

## Contributing

Contributions welcome! Please open an issue or PR.

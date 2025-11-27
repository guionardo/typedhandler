# 🏷️ Struct Tag-Based

## Request Struct

Parsing of the request data uses a special struct with tags that describes from where it will get the information.

We will call this struct as `RequestStruct`.

## HTTP Request Anatomy

```
> curl -X METHOD http://apihost/path/{pid}/resource/{rid}?name=John&age=20 -H "Authorization: Bearer" -d '{"message":"Great news!", "success": true}'
```

First step: define your request data using a struct with special tags:

| Part | Tag    | Description                                                   | Type  |
| ---- | ------ | --------------------------------------------------------------- | ----|
| path | [path](#path)   | Value from request path parameter                               | [multiple](#type-conversion)|
| query | [query](#query)  | Value from request query (alias = `form`)                       |[multiple](#type-conversion)
| header | [header](#header) | Value from request header                                       |only string |
| body | [json](#json)   | Request body will be unmarshaled into struct                    | struct |
| body | [body](#body)   | Request body will be unmarshaled into inner field of the struct | struct |

```go
type LoginRequest struct {
    UserName string `json:"username"`
    Password string `json:"password"`
}
```

In this example, the `json` tags will be interpreted as a request with a JSON body that will be unmarshaled into a `LoginRequest` struct.

```
> curl -X POST http://localhost/login -d '{"username":"john","password":"mary"}'
```

You can use these tags to parse the request:

## `path`

For route path parameters.

```go
type SampleRequest struct {
    Id int `path:"id"`
}
```

In [net/http](https://pkg.go.dev/net/http#hdr-Patterns-ServeMux), when you define a route like this:

```go
http.HandleFunc("/item/{id}", handler)
```

```shell
> curl http://localhost/item/10
```

The request will populate the field `Id` with the value 10 (converted to int) provided in the path.

## `query`

For query parameters.

```go
type SampleRequest struct {
    Name string `query:"name"`
}
```

```shell
> curl http://localhost/?name=John
```

The request will populate the field `Name` with the value "John"

## `header`

For header parameters.

```go
type SampleRequest struct {
    Authorization string `header:"authorization"`
}
```

```shell
> curl http://localhost -H 'Authorization: token'
```

## `json`

For JSON body.

## `body`

## Type conversion

The automatic parsing of the fields will convert to:

* String (default)
* Bool (parsed by [strconv.ParseBool](https://pkg.go.dev/strconv#ParseBool))
* Int (8, 16, 32, 64) (parsed by [strconv.ParseInt](https://pkg.go.dev/strconv#ParseInt))
* UInt (8, 16, 32, 64) (parsed by [strconv.ParseUint](https://pkg.go.dev/strconv#ParseUInt))
* Float (32, 64) (parsed by [strconv.ParseFloat](https://pkg.go.dev/strconv#ParseFloat))
* time.Duration (parsed by [time.ParseDuration](https://pkg.go.dev/time#ParseDuration))
* time.Time (smart parsing with multiple layouts)

### ⏱️ time.Time parsing

Time can be represented by different layouts. To handle this, I choose a dynamic approach.

Using a slice of layouts, the parser will try to parse a time string throught all the layouts. When it finds a match, the layout will be the first to be used in the next parsing, optimizing the search by a valid layout on each iteration.

The default [layouts](https://pkg.go.dev/time#pkg-constants) are:

* time.DateTime
* time.RFC3339
* time.RFC3339Nano
* time.RFC1123
* time.RFC1123Z
* time.ANSIC
* time.DateOnly
* time.TimeOnly

You can change it using the func `typedhandler.SetTimeLayouts(layouts []string)`

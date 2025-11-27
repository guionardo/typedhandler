# 🏷️ Struct Tag-Based

First step: define your request data using a struct with special tags:

| Tag    | Description                                                     |
| ------ | --------------------------------------------------------------- |
| json   | Request body will be unmarshaled into struct                    |
| path   | Value from request path parameter                               |
| query  | Value from request query (alias = `form`)                       |
| header | Value from request header                                       |
| body   | Request body will be unmarshaled into inner field of the struct |

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

## `json`

For JSON body.

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

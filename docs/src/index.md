# Typed Handler

Typed Handler is a high-performance, zero-allocation HTTP request parser for Go that automatically maps HTTP requests to strongly-typed structs using generics and object pooling.

<div class="grid cards" markdown>

- 🚀 **Zero Allocations**

    ---

    Uses `sync.Pool` for request struct reuse

- 🎯 **Type-Safe**

    ---

    Leverages Go generics for compile-time type checking

- 🏷️ **Struct Tag-Based**

    ---

    Parse path params, query strings, headers, and body with simple tags

    <!-- [Learn more about the dependency container :material-arrow-right:](container.md) -->

- ⚡ **High Performance**

    ---

    Reflection done once at initialization, cached for reuse

    [Learn more about modules :material-arrow-right:](modules.md)

- 🔧 **Flexible**

    Supports custom error types, body parsing strategies, and reset patterns

</div>

[Get started :material-arrow-right-bold:](get-started/index.md){ .md-button .md-button--primary }

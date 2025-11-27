# Typed Handler

Typed Handler is a high-performance, minimal-allocation HTTP request parser for Go that automatically maps HTTP requests to strongly-typed structs using generics and object pooling.

<div class="grid cards" markdown>

- 🚀 **Minimal Allocations**

    ---

    Uses `sync.Pool` for request struct reuse

- 🎯 **Type-Safe**

    ---

    Leverages Go generics for compile-time type checking

- 🏷️ **Struct Tag-Based**

    ---

    Parse path params, query strings, headers, and body with simple tags

    [Learn more about the control tags :material-arrow-right:](concepts/struct-tag-based.md)

- ⚡ **High Performance**

    ---

    Reflection done once at initialization, cached for reuse

- 🔧 **Flexible**

    ---

    Supports custom error types, body parsing strategies, and reset patterns

- ✨ **Future features**

    ---

    Generated code, to increase performance with zero reflection

</div>

[Get started :material-arrow-right-bold:](get-started/index.md){ .md-button .md-button--primary }

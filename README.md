# Kore

Kore is a lightweight, clean, and highly extensible Dependency Injection (DI) framework for Go. It focuses on the clean separation of **Discovery** (locating dependencies) and **Identity** (component metadata), providing a predictable and type-safe way to manage complex component graphs.

## Key Features

- **Clean Architecture**: Separates navigation (`Locator`) from instantiation context (`Handle`).
- **Discovery vs Identity**: Components can find others via locators without leaking their own internal state.
- **Multidimensional Isolation**: Support for Categories, Scopes, and Capability Tags.
- **Lazy Initialization**: Components are instantiated only when first requested.
- **Circular Dependency Detection**: Built-in safety against recursive dependency chains.
- **Zero External Dependencies**: Pure Go implementation.

## Installation

```bash
go get github.com/origadmin/kore
```

## Quick Start

### 1. Define Your Components

```go
type Database struct {
    DSN string
}

type Logger struct {
    Prefix string
}

type Service struct {
    db     *Database
    logger *Logger
}
```

### 2. Register Providers

```go
import "github.com/origadmin/kore"

reg := kore.New()

// Register a Logger
reg.Register("logger", func(ctx context.Context, h kore.Handle) (any, error) {
    return &Logger{Prefix: "[APP]"}, nil
})

// Register a Database with dependency on Logger
reg.Register("database", func(ctx context.Context, h kore.Handle) (any, error) {
    // Find Logger using Locator
    logger, _ := kore.GetDefault[*Logger](ctx, h.Locator().In("logger"))
    return &Database{DSN: "localhost:5432"}, nil
})
```

### 3. Load and Use

```go
ctx := context.Background()
reg.Load(ctx, nil) // Load configurations

// Retrieve the database
db, err := kore.GetDefault[*Database](ctx, reg.In("database"))
```

## Core Concepts: The Multi-dimensional DI

Kore organizes components into a three-dimensional coordinate system, allowing for precise control over component visibility and lifecycle.

### 1. Horizontal Dimension: Categories
Categories define **what** a component is (e.g., `database`, `cache`, `logger`). 
- **Purpose**: Grouping identical types of infrastructure.
- **Usage**: You can iterate over all components within a category or retrieve a specific one by name.

### 2. Vertical Dimension: Scopes
Scopes define **where** a component exists (e.g., `_global`, `server`, `client`).
- **Hierarchy**: Components in a specific scope are isolated from others.
- **Fallback**: Kore supports intelligent fallback mechanisms, allowing local scopes to inherit or override global defaults.

### 3. Specialization: Capability Tags
Tags define **which variant** or **what capability** a provider offers (e.g., `gateway`, `mock`, `production`).
- **Isolation**: A `Locator` created with the `gateway` tag can only "see" providers registered with that same tag (or common providers).
- **Perspective**: This allows you to create different "perspectives" of the same dependency graph for different parts of your application.

## Dual-Axis Isolation Model

Kore provides a robust isolation model that separates components along two orthogonal axes: **Vertical (Contextual)** and **Horizontal (Functional)**.

### Vertical Isolation: S-Single (Scope-based)
Vertical isolation is achieved through **Scopes**, partitioning components based on their **Runtime Environment**.
- **Mechanism**: Each Scope (e.g., `_global`, `server`, `client`) is a physically distinct instance container.
- **S-Single**: For a given `(Category, Scope, Name)`, there is exactly one instance. This ensures that a component in the `server` scope cannot accidentally reference or interfere with a component in the `client` scope, even if they share the same name.
- **Use Case**: Environment-specific overrides and lifecycle management.

### Horizontal Isolation: T-Multi (Tag-based)
Horizontal isolation is achieved through **Tags**, partitioning components based on their **Responsibility and Capability**.
- **Mechanism**: Providers register themselves with specific Capability Tags. When you enter a category via `reg.In(cat, WithInTags("gateway"))`, you create a **Perspective**.
- **T-Multi**: A single Category can house multiple providers for the same name, but only those matching the requested "Perspective Tags" (the Capability Set) are visible.
- **Responsibility Filtering**: This allows you to hide internal-only components from public-facing locators or switch between different implementation variants (e.g., `production` vs `mock`) without changing the component's name or code.
- **Inheritance**: "Common" providers (those with no tags) are visible across all horizontal perspectives, acting as a shared foundation.

## Native Configuration Binding

Unlike other DI frameworks that treat configuration as an afterthought, Kore is a **Config-First** engine.
- **The Resolver**: Every category can have a `Resolver` that maps raw input (YAML, JSON, or Map) to structured `ModuleConfig`.
- **Automatic Instantiation**: Once a config entry is mapped (e.g., a database config named `mysql`), Kore automatically knows how to use the registered `Provider` to turn that config into a live object.
- **Active Instance**: Supports an `Active` flag in configuration to designate which named instance should be the `_default` for that category.

## Comparison with Mainstream Frameworks

| Feature | Kore | Google Wire | Uber Fx / Dig |
| :--- | :--- | :--- | :--- |
| **Mechanism** | Run-time (Provider-based) | Compile-time (Code Gen) | Run-time (Reflection) |
| **Dependency Resolution** | Explicit via `Locator` | Implicit via Type Matching | Implicit via Type Matching |
| **Identity vs Search** | **Absolute Separation** | None (Single Graph) | None (Single Graph) |
| **Multi-instance Support** | Native (Category/Name/Tag) | Complex (using tags/alias) | Complex (using groups/names) |
| **Configuration Binding** | Built-in `Resolver` | Manual | Manual / External |
| **Error Detection** | First-use / Initialization | Compile-time | App Start-up |

## Why Kore? The Competitive Edge

Kore is not just another object wirer; it is a **Structured Discovery Engine**. Compared to mainstream frameworks like Google Wire or Uber Fx, Kore provides a multi-dimensional approach to dependency management.

### Feature Comparison

| Feature | Kore | Google Wire | Uber Fx / Dig |
| :--- | :--- | :--- | :--- |
| **Vertical Isolation (S-Single)** | **Native (Scopes)**: Physical isolation of instances by environment. | **None**: Single global graph; naming conflicts are manual. | **Weak**: Limited lifecycle scoping; lacks multi-layer isolation. |
| **Horizontal Filtering (T-Multi)** | **Native (Tags)**: Logical filtering of responsibilities via Perspectives. | **None**: Relies on static types or manual struct naming. | **Complex**: Requires annotations or groups; hard to maintain. |
| **Config Binding (Config-First)** | **Deeply Integrated**: Automatic mapping from raw config to instances. | **None**: Pure wiring; config handling is entirely external. | **None**: Requires external libraries; mapping is scattered. |
| **Separation (Identity/Locator)** | **Absolute (Plan A)**: Identity (Handle) is isolated from Discovery. | **Mixed**: Components have excessive visibility into the graph. | **Obscure**: Reflection leads to implicit container coupling. |
| **Multi-instance Management** | **Seamless**: Native support for N-instances via Category/Name/Tag. | **Verbose**: Requires unique types/providers for each instance. | **Heavy**: Complexity grows exponentially with instance counts. |

### The Kore Advantage

#### 1. Precision via S-Single & T-Multi
- **Vertical (S-Single)**: Ensures that your `server` database and `monitoring` database never touch, even if they share the same provider. 
- **Horizontal (T-Multi)**: Allows you to define implementation variants (e.g., `public-api` vs `internal-tool`) in the same category and switch between them simply by changing the locator's perspective.

#### 2. Native Configuration-to-Instance Pipeline
Kore bridges the gap between static code and dynamic configuration. By defining a `Resolver`, you transform your DI from a static object graph into a dynamic engine that responds to your environment settings.

#### 3. Predictability over Magic
By enforcing an explicit `h.Locator().Get()` pattern, Kore removes the "magic" of reflection and the "clutter" of code generation. You gain absolute traceability of every dependency in your system.

## Architecture: Plan A (Composition)

Kore implements the **Plan A** composition pattern, where the `Handle` (Identity) provided during instantiation does not inherit from the `Locator` (Navigation). Instead, it provides a `Locator()` method. This ensures that:

1.  **Identity is Local**: A component's name and config are only accessible via its own `Handle`.
2.  **Navigation is Explicit**: To find another component, you must explicitly use a `Locator`.
3.  **Automatic State Stripping**: Calling `In()` on a locator returns a fresh locator without any residual identity from the caller, preventing accidental state leakage.

## License

Kore is released under the MIT License.

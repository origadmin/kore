/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package kore

import (
	"context"
	"iter"
)

// Locator defines the interface for outward navigation and component discovery.
// It is a pure search and positioning tool that does not carry identity or configuration.
type Locator interface {
	// Get retrieves a component by name from the current perspective (category/scope/tags).
	Get(ctx context.Context, name string) (any, error)
	// Iter returns a type-safe iterator for all registered components in the current perspective.
	Iter(ctx context.Context) iter.Seq2[string, any]
	// In switches to a different perspective (category/scope/tags) and returns a new Locator.
	// This operation is a "coordinate shift" and results in a clean locator without previous identity.
	In(cat Category, opts ...InOption) Locator

	// Category returns the current category of the locator.
	Category() Category
	// Scope returns the current scope of the locator.
	Scope() Scope
}

// Handle represents the execution context and identity of a component during its instantiation.
// It is the "passport" passed to the Provider function.
type Handle interface {
	// Name returns the unique name assigned to this component instance.
	Name() string
	// Config returns the configuration object associated with this component instance.
	Config() any
	// Locator provides access to the search capability, allowing this component to find its dependencies.
	// The locator's coordinates (Category/Scope/Tags) will match the component's original registration.
	Locator() Locator
}

// Provider is a function that creates a component instance based on its handle and locator.
type Provider func(ctx context.Context, h Handle) (any, error)

// Registry defines the core DI container interface for registration, configuration loading, and discovery.
type Registry interface {
	// Register stores a component provider with its metadata into the registry.
	Register(cat Category, p Provider, opts ...RegisterOption)
	// Has checks if a specific category has any registered providers.
	Has(cat Category, opts ...RegisterOption) bool
	// Load binds external configuration sources to registered categories and instances.
	Load(ctx context.Context, source any, opts ...LoadOption) error
	// In enters a specific category and scope perspective, returning a Locator for component retrieval.
	In(cat Category, opts ...InOption) Locator
}

// Resolver is a function that transforms a raw source into a structured module configuration.
type Resolver func(source any, cat Category) (*ModuleConfig, error)

// ModuleConfig defines the structured mapping of configuration entries for a specific category.
type ModuleConfig struct {
	// Entries is a list of named configuration values.
	Entries []ConfigEntry
	// Active specifies which entry should be considered the default/primary instance.
	Active string
}

// ConfigEntry is a single named configuration value.
type ConfigEntry struct {
	Name  string
	Value any
}

// New creates a new DI container instance (Registry) with the provided options.
func New(opts ...RegistryOption) Registry {
	o := &RegistryOptions{
		CategoryResolvers: make(map[Category]Resolver),
	}
	for _, opt := range opts {
		if opt != nil {
			opt(o)
		}
	}

	return &containerImpl{
		modules:           make(map[moduleKey]*moduleState),
		providers:         make(map[Category][]*providerEntry),
		categoryResolvers: o.CategoryResolvers,
	}
}

// --- Generic Helpers ---

// Get retrieves a component by name from a locator and asserts its type.
func Get[T any](ctx context.Context, l Locator, name string) (T, error) {
	var zero T
	inst, err := l.Get(ctx, name)
	if err != nil {
		return zero, err
	}
	if t, ok := inst.(T); ok {
		return t, nil
	}
	return zero, nil
}

// GetDefault retrieves the default component from a locator and asserts its type.
func GetDefault[T any](ctx context.Context, l Locator) (T, error) {
	return Get[T](ctx, l, DefaultName)
}

// AsConfig extracts and asserts the configuration from a handle.
func AsConfig[T any](h Handle) (*T, error) {
	cfg := h.Config()
	if cfg == nil {
		return nil, nil
	}
	if t, ok := cfg.(*T); ok {
		return t, nil
	}
	return nil, nil
}

// Iter returns a type-safe iterator for components in a locator.
func Iter[T any](ctx context.Context, l Locator) iter.Seq2[string, T] {
	return func(yield func(string, T) bool) {
		for name, inst := range l.Iter(ctx) {
			if t, ok := inst.(T); ok {
				if !yield(name, t) {
					return
				}
			}
		}
	}
}

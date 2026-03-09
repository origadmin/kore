/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package kore

// Common Types for DI management
type (
	Category string
	Scope    string
	Priority int
)

const (
	// GlobalScope is the default system fallback scope.
	GlobalScope Scope = "_global"

	// DefaultName is the system key for the active/default instance.
	DefaultName = "_default"
)

// --- Registration Options (Providers) ---

type RegistrationOptions struct {
	Resolver Resolver
	Scopes   []Scope
	Priority Priority
	Tag      string
}

type RegisterOption func(*RegistrationOptions)

func WithScope(s Scope) RegisterOption {
	return func(o *RegistrationOptions) { o.Scopes = append(o.Scopes, s) }
}

func WithPriority(p Priority) RegisterOption {
	return func(o *RegistrationOptions) { o.Priority = p }
}

func WithTag(tag string) RegisterOption {
	return func(o *RegistrationOptions) { o.Tag = tag }
}

func WithResolver(res Resolver) RegisterOption {
	return func(o *RegistrationOptions) { o.Resolver = res }
}

// --- Perspective Options (In/Query) ---

type InOptions struct {
	Scope Scope
	Tags  []string
}

type InOption func(*InOptions)

func WithInScope(s Scope) InOption {
	return func(o *InOptions) { o.Scope = s }
}

func WithInTags(tags ...string) InOption {
	return func(o *InOptions) { o.Tags = append(o.Tags, tags...) }
}

// --- Loading Options ---

type LoadOptions struct {
	Category Category
	Scope    Scope
	Name     string
	Resolver Resolver
	Tags     []string
}

type LoadOption func(*LoadOptions)

func ForCategory(cat Category) LoadOption {
	return func(o *LoadOptions) { o.Category = cat }
}

func ForName(name string) LoadOption {
	return func(o *LoadOptions) { o.Name = name }
}

func ForScope(s Scope) LoadOption {
	return func(o *LoadOptions) { o.Scope = s }
}

func WithLoadResolver(res Resolver) LoadOption {
	return func(o *LoadOptions) { o.Resolver = res }
}

// --- Container Options ---

type RegistryOptions struct {
	CategoryResolvers map[Category]Resolver
}

type RegistryOption func(*RegistryOptions)

func WithCategoryResolvers(res map[Category]Resolver) RegistryOption {
	return func(o *RegistryOptions) {
		if o.CategoryResolvers == nil {
			o.CategoryResolvers = make(map[Category]Resolver)
		}
		for k, v := range res {
			o.CategoryResolvers[k] = v
		}
	}
}

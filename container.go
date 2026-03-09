/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package kore

import (
	"context"
	"fmt"
	"iter"
	"sync"
)

type status int

const (
	statusNone status = iota
	statusInstantiating
	statusReady
	statusError
)

type moduleKey struct {
	category Category
	scope    Scope
}

type componentMeta struct {
	config any
	status status
	inst   any
	err    error
	tag    string
}

type moduleState struct {
	mu          sync.RWMutex
	instances   map[string]*componentMeta
	order       []string
	defaultName string
	bound       bool
}

func makeInstanceKey(name, tag string) string {
	if tag == "" {
		return name
	}
	return name + "@" + tag
}

func configKey(name string) string {
	return makeInstanceKey(name, "_config")
}

type providerEntry struct {
	provider Provider
	resolver Resolver
	scopes   []Scope
	priority Priority
	tag      string
}

type containerImpl struct {
	mu                sync.RWMutex
	modules           map[moduleKey]*moduleState
	providers         map[Category][]*providerEntry
	categoryResolvers map[Category]Resolver
	isLoaded          bool
}

func (c *containerImpl) Register(cat Category, p Provider, opts ...RegisterOption) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.isLoaded {
		panic(fmt.Sprintf("kore: cannot register category %s after Load() has been called", cat))
	}

	cfg := &RegistrationOptions{}
	for _, opt := range opts {
		opt(cfg)
	}

	entry := &providerEntry{
		provider: p,
		resolver: cfg.Resolver,
		scopes:   cfg.Scopes,
		priority: cfg.Priority,
		tag:      cfg.Tag,
	}

	entries := c.providers[cat]
	inserted := false
	for i, e := range entries {
		if entry.priority >= e.priority {
			entries = append(entries[:i], append([]*providerEntry{entry}, entries[i:]...)...)
			inserted = true
			break
		}
	}
	if !inserted {
		entries = append(entries, entry)
	}
	c.providers[cat] = entries
}

func (c *containerImpl) Has(cat Category, opts ...RegisterOption) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, ok := c.providers[cat]
	return ok
}

func (c *containerImpl) Load(ctx context.Context, source any, opts ...LoadOption) error {
	c.mu.Lock()
	c.isLoaded = true
	c.mu.Unlock()

	loadOpts := &LoadOptions{}
	for _, opt := range opts {
		opt(loadOpts)
	}

	c.mu.RLock()
	var cats []Category
	if loadOpts.Category != "" {
		if _, ok := c.providers[loadOpts.Category]; ok {
			cats = append(cats, loadOpts.Category)
		}
	} else {
		for cat := range c.providers {
			cats = append(cats, cat)
		}
	}
	c.mu.RUnlock()

	for _, cat := range cats {
		entries := c.getProviderEntries(cat)
		if len(entries) == 0 {
			continue
		}

		primaryEntry := entries[0]
		registeredScopes := make(map[Scope]bool)
		for _, entry := range entries {
			if len(entry.scopes) == 0 {
				registeredScopes[GlobalScope] = true
			} else {
				for _, s := range entry.scopes {
					registeredScopes[s] = true
				}
			}
		}

		for s := range registeredScopes {
			if loadOpts.Scope != "" && s != loadOpts.Scope {
				continue
			}
			if err := c.bindWithSource(cat, s, primaryEntry, source, loadOpts.Resolver, loadOpts.Name); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *containerImpl) getProviderEntries(cat Category) []*providerEntry {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.providers[cat]
}

func (c *containerImpl) bindWithSource(cat Category, scope Scope, entry *providerEntry, source any, resolver Resolver, filterName string) error {
	mKey := moduleKey{category: cat, scope: scope}
	s := c.getModuleState(mKey)
	s.mu.Lock()
	defer s.mu.Unlock()

	var mc *ModuleConfig
	var err error

	if resolver != nil {
		mc, err = resolver(source, cat)
	} else if entry.resolver != nil {
		mc, err = entry.resolver(source, cat)
	} else {
		c.mu.RLock()
		r := c.categoryResolvers[cat]
		c.mu.RUnlock()
		if r != nil {
			mc, err = r(source, cat)
		}
	}

	if err == nil && mc == nil {
		mc = &ModuleConfig{
			Entries: []ConfigEntry{{Name: DefaultName, Value: source}},
			Active:  DefaultName,
		}
	}

	if err != nil {
		return err
	}

	for _, cfgEntry := range mc.Entries {
		if filterName != "" && cfgEntry.Name != filterName {
			continue
		}

		key := configKey(cfgEntry.Name)
		if _, exists := s.instances[key]; !exists {
			s.instances[key] = &componentMeta{config: cfgEntry.Value, status: statusNone}
			s.order = append(s.order, cfgEntry.Name)
		}
	}

	if mc.Active != "" && (filterName == "" || mc.Active == filterName) {
		s.defaultName = mc.Active
	} else if s.defaultName == "" {
		foundDefault := false
		for _, e := range mc.Entries {
			if e.Name == DefaultName {
				s.defaultName = e.Name
				foundDefault = true
				break
			}
		}
		if !foundDefault && len(mc.Entries) == 1 {
			s.defaultName = mc.Entries[0].Name
		}
	}

	s.bound = true
	return nil
}

func (c *containerImpl) getModuleState(key moduleKey) *moduleState {
	c.mu.Lock()
	defer c.mu.Unlock()
	if s, ok := c.modules[key]; ok {
		return s
	}
	s := &moduleState{
		instances: make(map[string]*componentMeta),
	}
	c.modules[key] = s
	return s
}

func (c *containerImpl) In(cat Category, opts ...InOption) Locator {
	inOpts := &InOptions{Scope: GlobalScope}
	for _, opt := range opts {
		opt(inOpts)
	}
	return &locatorHandle{
		c:        c,
		category: cat,
		scope:    inOpts.Scope,
		tags:     inOpts.Tags,
	}
}

func (c *containerImpl) instantiate(ctx context.Context, cat Category, scope Scope, name string, requestedTags []string) (any, error) {
	if name == "" {
		name = DefaultName
	}

	mKey := moduleKey{category: cat, scope: scope}
	c.mu.RLock()
	s, exists := c.modules[mKey]
	c.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("kore: scope %s not initialized for category %s", scope, cat)
	}

	// Name Resolution: Handle DefaultName redirection
	actualName := name
	if name == DefaultName && s.defaultName != "" {
		actualName = s.defaultName
	}

	s.mu.RLock()
	configMeta, ok := s.instances[configKey(actualName)]
	s.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("kore: component %s/%s not found in scope %s", cat, actualName, scope)
	}

	c.mu.RLock()
	entries := c.providers[cat]
	c.mu.RUnlock()

	var lastErr error
	for _, entry := range entries {
		scopeMatch := false
		if len(entry.scopes) == 0 {
			scopeMatch = true
		} else {
			for _, es := range entry.scopes {
				if es == scope {
					scopeMatch = true
					break
				}
			}
		}
		if !scopeMatch {
			continue
		}

		if !isProviderVisible(entry.tag, requestedTags) {
			continue
		}

		instanceKey := makeInstanceKey(name, entry.tag)
		s.mu.Lock()
		meta, exists := s.instances[instanceKey]
		if !exists {
			meta = &componentMeta{config: configMeta.config, status: statusNone, tag: entry.tag}
			s.instances[instanceKey] = meta
		}

		if meta.status == statusReady {
			inst := meta.inst
			s.mu.Unlock()
			return inst, nil
		}
		if meta.status == statusInstantiating {
			s.mu.Unlock()
			return nil, fmt.Errorf("kore: circular dependency %s/%s", cat, instanceKey)
		}

		meta.status = statusInstantiating
		s.mu.Unlock()

		h := &entryHandle{
			name: actualName,
			meta: meta,
			l: &locatorHandle{
				c:        c,
				category: cat,
				scope:    scope,
				tags:     requestedTags,
			},
		}
		inst, err := entry.provider(ctx, h)

		s.mu.Lock()
		if err == nil && inst != nil {
			meta.inst = inst
			meta.status = statusReady
			s.mu.Unlock()
			return inst, nil
		}

		meta.status = statusNone
		if err != nil {
			lastErr = err
		}
		s.mu.Unlock()
	}

	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("kore: no compatible provider found for %s/%s in scope %s with capabilities %v", cat, name, scope, requestedTags)
}

func (c *containerImpl) iterInternal(ctx context.Context, cat Category, scope Scope, tags []string) iter.Seq2[string, any] {
	return func(yield func(string, any) bool) {
		mKey := moduleKey{category: cat, scope: scope}
		s := c.getModuleState(mKey)
		s.mu.RLock()
		order := make([]string, len(s.order))
		copy(order, s.order)
		s.mu.RUnlock()

		for _, name := range order {
			inst, err := c.instantiate(ctx, cat, scope, name, tags)
			if err == nil {
				if !yield(name, inst) {
					return
				}
			}
		}
	}
}

func isProviderVisible(providerTag string, requestedTags []string) bool {
	if providerTag == "" {
		return true
	}
	if len(requestedTags) == 0 {
		return true
	}
	for _, rt := range requestedTags {
		if providerTag == rt {
			return true
		}
	}
	return false
}

type locatorHandle struct {
	c        *containerImpl
	category Category
	scope    Scope
	tags     []string
}

func (l *locatorHandle) Get(ctx context.Context, name string) (any, error) {
	return l.c.instantiate(ctx, l.category, l.scope, name, l.tags)
}

func (l *locatorHandle) Iter(ctx context.Context) iter.Seq2[string, any] {
	return l.c.iterInternal(ctx, l.category, l.scope, l.tags)
}

func (l *locatorHandle) In(cat Category, opts ...InOption) Locator {
	inOpts := &InOptions{
		Scope: GlobalScope,
		Tags:  l.tags,
	}
	for _, opt := range opts {
		opt(inOpts)
	}
	return &locatorHandle{
		c:        l.c,
		category: cat,
		scope:    inOpts.Scope,
		tags:     inOpts.Tags,
	}
}

func (l *locatorHandle) Scope() Scope       { return l.scope }
func (l *locatorHandle) Category() Category { return l.category }

type entryHandle struct {
	name string
	meta *componentMeta
	l    *locatorHandle
}

func (e *entryHandle) Name() string { return e.name }
func (e *entryHandle) Config() any {
	if e.meta == nil {
		return nil
	}
	return e.meta.config
}
func (e *entryHandle) Locator() Locator { return e.l }

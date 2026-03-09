/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package kore_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/origadmin/kore"
)

// --- Mocks ---

type Database struct {
	DSN    string
	Logger *Logger
}

type Logger struct {
	Prefix string
}

func TestKore_FullLifecycle(t *testing.T) {
	ctx := context.Background()
	reg := kore.New()

	// 1. Register Logger
	reg.Register("logger", func(ctx context.Context, h kore.Handle) (any, error) {
		return &Logger{Prefix: "[KORE]"}, nil
	})

	// 2. Register Database with dependency on Logger
	reg.Register("database", func(ctx context.Context, h kore.Handle) (any, error) {
		// Verify local identity
		assert.Equal(t, "main-db", h.Name())
		assert.Equal(t, "postgres://localhost", h.Config())

		// Explicit discovery using Locator
		l := h.Locator().In("logger")
		log, err := kore.GetDefault[*Logger](ctx, l)
		if err != nil {
			return nil, err
		}

		return &Database{
			DSN:    h.Config().(string),
			Logger: log,
		}, nil
	})

	// 3. Load configurations
	err := reg.Load(ctx, nil, kore.WithLoadResolver(func(source any, cat kore.Category) (*kore.ModuleConfig, error) {
		if cat == "database" {
			return &kore.ModuleConfig{
				Entries: []kore.ConfigEntry{{Name: "main-db", Value: "postgres://localhost"}},
				Active:  "main-db",
			}, nil
		}
		return nil, nil
	}))
	assert.NoError(t, err)

	// 4. Retrieve Database
	db, err := kore.GetDefault[*Database](ctx, reg.In("database"))
	assert.NoError(t, err)
	assert.Equal(t, "postgres://localhost", db.DSN)
	assert.Equal(t, "[KORE]", db.Logger.Prefix)
}

func TestKore_CircularDependency(t *testing.T) {
	ctx := context.Background()
	reg := kore.New()

	// A depends on B
	reg.Register("A", func(ctx context.Context, h kore.Handle) (any, error) {
		_, err := h.Locator().In("B").Get(ctx, kore.DefaultName)
		return "A-instance", err
	})

	// B depends on A
	reg.Register("B", func(ctx context.Context, h kore.Handle) (any, error) {
		_, err := h.Locator().In("A").Get(ctx, kore.DefaultName)
		return "B-instance", err
	})

	_ = reg.Load(ctx, nil)

	_, err := reg.In("A").Get(ctx, kore.DefaultName)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "circular dependency")
}

func TestKore_PerspectiveIsolation(t *testing.T) {
	ctx := context.Background()
	reg := kore.New()

	reg.Register("service", func(ctx context.Context, h kore.Handle) (any, error) {
		return "GatewayService", nil
	}, kore.WithTag("gateway"))

	reg.Register("service", func(ctx context.Context, h kore.Handle) (any, error) {
		return "FeatureService", nil
	}, kore.WithTag("feature"))

	_ = reg.Load(ctx, nil)

	// Gateway perspective should only see GatewayService
	hGateway := reg.In("service", kore.WithInTags("gateway"))
	inst, err := hGateway.Get(ctx, kore.DefaultName)
	assert.NoError(t, err)
	assert.Equal(t, "GatewayService", inst)

	// Feature perspective should only see FeatureService
	hFeature := reg.In("service", kore.WithInTags("feature"))
	inst2, err := hFeature.Get(ctx, kore.DefaultName)
	assert.NoError(t, err)
	assert.Equal(t, "FeatureService", inst2)
}

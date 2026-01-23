// Package telemetry provides OpenTelemetry tracing for the application.
package telemetry

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

type contextKey string

const (
	gormSpanKey        contextKey = "otel:span"
	callBackBeforeName string     = "otel:before"
	callBackAfterName  string     = "otel:after"
)

// GORMTelemetryPlugin is a GORM plugin that adds OpenTelemetry tracing
type GORMTelemetryPlugin struct {
	tracer trace.Tracer
}

// NewGORMTelemetryPlugin creates a new GORM telemetry plugin
func NewGORMTelemetryPlugin(serviceName string) *GORMTelemetryPlugin {
	return &GORMTelemetryPlugin{
		tracer: otel.Tracer(serviceName),
	}
}

// Name returns the plugin name
func (p *GORMTelemetryPlugin) Name() string {
	return "otel:gorm"
}

// Initialize registers the callbacks for tracing
func (p *GORMTelemetryPlugin) Initialize(db *gorm.DB) error {
	// Register before callbacks
	if err := db.Callback().Create().Before("gorm:create").Register(callBackBeforeName, p.before("create")); err != nil {
		return err
	}
	if err := db.Callback().Query().Before("gorm:query").Register(callBackBeforeName, p.before("query")); err != nil {
		return err
	}
	if err := db.Callback().Update().Before("gorm:update").Register(callBackBeforeName, p.before("update")); err != nil {
		return err
	}
	if err := db.Callback().Delete().Before("gorm:delete").Register(callBackBeforeName, p.before("delete")); err != nil {
		return err
	}
	if err := db.Callback().Row().Before("gorm:row").Register(callBackBeforeName, p.before("row")); err != nil {
		return err
	}
	if err := db.Callback().Raw().Before("gorm:raw").Register(callBackBeforeName, p.before("raw")); err != nil {
		return err
	}

	// Register after callbacks
	if err := db.Callback().Create().After("gorm:create").Register(callBackAfterName, p.after()); err != nil {
		return err
	}
	if err := db.Callback().Query().After("gorm:query").Register(callBackAfterName, p.after()); err != nil {
		return err
	}
	if err := db.Callback().Update().After("gorm:update").Register(callBackAfterName, p.after()); err != nil {
		return err
	}
	if err := db.Callback().Delete().After("gorm:delete").Register(callBackAfterName, p.after()); err != nil {
		return err
	}
	if err := db.Callback().Row().After("gorm:row").Register(callBackAfterName, p.after()); err != nil {
		return err
	}
	if err := db.Callback().Raw().After("gorm:raw").Register(callBackAfterName, p.after()); err != nil {
		return err
	}

	return nil
}

func (p *GORMTelemetryPlugin) before(operation string) func(*gorm.DB) {
	return func(db *gorm.DB) {
		ctx := db.Statement.Context
		if ctx == nil {
			ctx = context.Background()
		}

		spanName := fmt.Sprintf("gorm:%s", operation)
		ctx, span := p.tracer.Start(ctx, spanName)

		// Store span in statement context
		db.Statement.Context = context.WithValue(ctx, gormSpanKey, span)

		// Add attributes
		span.SetAttributes(
			attribute.String("db.system", "postgresql"),
			attribute.String("db.operation", operation),
		)

		// Add table name if available
		if db.Statement.Table != "" {
			span.SetAttributes(attribute.String("db.table", db.Statement.Table))
		}
	}
}

func (p *GORMTelemetryPlugin) after() func(*gorm.DB) {
	return func(db *gorm.DB) {
		// Retrieve span from context
		spanValue := db.Statement.Context.Value(gormSpanKey)
		if spanValue == nil {
			return
		}

		span, ok := spanValue.(trace.Span)
		if !ok {
			return
		}
		defer span.End()

		// Add SQL query as attribute
		if db.Statement.SQL.String() != "" {
			span.SetAttributes(attribute.String("db.statement", db.Statement.SQL.String()))
		}

		// Add rows affected
		if db.Statement.RowsAffected != 0 {
			span.SetAttributes(attribute.Int64("db.rows_affected", db.Statement.RowsAffected))
		}

		// Record error if any
		if db.Error != nil {
			span.RecordError(db.Error)
			span.SetStatus(codes.Error, db.Error.Error())
		} else {
			span.SetStatus(codes.Ok, "")
		}
	}
}

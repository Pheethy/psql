package psql

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/opentracing/opentracing-go"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/spf13/cast"
)

type TracingHook struct {
	tracer opentracing.Tracer
}

func NewTracingHook(tracing opentracing.Tracer) *TracingHook {
	return &TracingHook{
		tracer: tracing,
	}
}

func (h *TracingHook) getOperationName(query string) string {
	defaultOperationName := "database"
	selectReg := regexp.MustCompile(`SELECT`)
	insertReg := regexp.MustCompile(`INSERT\s+INTO`)
	updateReg := regexp.MustCompile(`UPDATE\s+.+\s+SET`)
	deleteReg := regexp.MustCompile(`DELETE\s+FROM`)

	query = strings.ToUpper(query)
	selectIndex := selectReg.FindStringIndex(query)
	insertIndex := insertReg.FindStringIndex(query)
	updateIndex := updateReg.FindStringIndex(query)
	deleteIndex := deleteReg.FindStringIndex(query)

	if selectIndex == nil && insertIndex == nil && updateIndex == nil && deleteIndex == nil {
		return defaultOperationName
	}

	if deleteIndex != nil {
		return "DELETE"
	}
	if updateIndex != nil {
		return "UPDATE"
	}
	if insertIndex != nil {
		return "INSERT"
	}
	if selectIndex != nil {
		return "SELECT"
	}

	return "NONE"
}

// Before hook will print the query with it's args and return the context with the timestamp
func (h *TracingHook) Before(ctx context.Context, query string, args ...interface{}) (context.Context, error) {
	if ctx != nil {
		span := opentracing.SpanFromContext(ctx)
		if span != nil {
			span, ctx = opentracing.StartSpanFromContext(ctx, "database", opentracing.ChildOf(span.Context()))
			span.SetTag("operation", h.getOperationName(query))
			span.LogFields(
				otlog.String("statement", query),
			)

			if args != nil && len(args) > 0 {
				var argsString = []string{}
				for index, arg := range args {
					argsString = append(argsString, fmt.Sprintf(`$$%s:%s`, cast.ToString(index+1), cast.ToString(arg)))
				}
				span.LogFields(
					otlog.String("args", strings.Join(argsString, ",")),
				)
			}
		}
	}
	return ctx, nil
}

// After hook will get the timestamp registered on the Before hook and print the elapsed time
func (h *TracingHook) After(ctx context.Context, query string, args ...interface{}) (context.Context, error) {
	if ctx != nil {
		span := opentracing.SpanFromContext(ctx)
		if span != nil {
			defer span.Finish()

			span.SetTag("error", false)
		}
	}
	return ctx, nil
}

// Hook OnError
func (h *TracingHook) OnError(ctx context.Context, err error, query string, args ...interface{}) error {
	if ctx != nil {
		span := opentracing.SpanFromContext(ctx)
		if span != nil {
			defer span.Finish()

			span.SetTag("error", true)
			span.LogFields(
				otlog.Message(err.Error()),
			)
		}
	}

	return err
}

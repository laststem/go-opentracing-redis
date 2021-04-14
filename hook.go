package opentracing_redis

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-redis/redis/v8"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

func Wrap(client redis.UniversalClient) {
	client.AddHook(&tracingHook{})
}

type tracingHook struct{}

func (h *tracingHook) BeforeProcess(ctx context.Context, cmd redis.Cmder) (context.Context, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, command(cmd))
	tagging(span)
	Customize(span, cmd)
	return ctx, nil
}

func (h *tracingHook) AfterProcess(ctx context.Context, cmd redis.Cmder) error {
	span := opentracing.SpanFromContext(ctx)
	span.Finish()
	return nil
}

func (h *tracingHook) BeforeProcessPipeline(ctx context.Context, cmds []redis.Cmder) (context.Context, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, command(cmds...))
	tagging(span)
	for _, cmd := range cmds {
		Customize(span, cmd)
	}
	return ctx, nil
}

func (h *tracingHook) AfterProcessPipeline(ctx context.Context, cmds []redis.Cmder) error {
	span := opentracing.SpanFromContext(ctx)
	span.Finish()
	return nil
}

func command(cmds ...redis.Cmder) string {
	var commands []string
	for _, cmd := range cmds {
		commands = append(commands, fmt.Sprintf("%v", cmd.String()))
	}
	return strings.Join(commands, ", ")
}

func tagging(span opentracing.Span) {
	ext.Component.Set(span, "redis")
	ext.DBType.Set(span, "redis")
	ext.SpanKind.Set(span, "client")
}

type customizer func(opentracing.Span, redis.Cmder)

var Customize customizer = func(opentracing.Span, redis.Cmder) {}

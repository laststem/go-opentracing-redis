package example

import (
	"context"
	"io"
	"log"
	"net/http"
	"time"

	ot_redis "github.com/laststem/go-opentracing-redis"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-lib/metrics"

	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegerlog "github.com/uber/jaeger-client-go/log"

	"github.com/go-redis/redis/v8"
)

func Example() {
	closer := initTracer()
	defer closer.Close()

	client := redis.NewClient(&redis.Options{})
	ot_redis.Wrap(client)

	http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		span, ctx := opentracing.StartSpanFromContext(context.Background(), "hello")
		defer span.Finish()

		business := func(ctx context.Context) {
			client.Set(ctx, "test", 1, time.Millisecond*10)
			client.Get(ctx, "test")

			client.Pipelined(ctx, func(pipeliner redis.Pipeliner) error {
				pipeliner.Set(ctx, "test2", 7, time.Second)
				pipeliner.Get(ctx, "test2")
				return nil
			})
		}
		business(ctx)
	})
	log.Fatal(http.ListenAndServe(":8082", nil))
}

func initTracer() io.Closer {
	cfg := jaegercfg.Configuration{
		ServiceName: "example-redis-tracing",
		Sampler: &jaegercfg.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1,
		},
		Reporter: &jaegercfg.ReporterConfig{
			LogSpans: true,
		},
	}

	jLogger := jaegerlog.StdLogger
	jMetricsFactory := metrics.NullFactory

	tracer, closer, _ := cfg.NewTracer(
		jaegercfg.Logger(jLogger),
		jaegercfg.Metrics(jMetricsFactory),
	)
	opentracing.SetGlobalTracer(tracer)
	return closer
}

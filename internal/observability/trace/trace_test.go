package trace

import (
	"context"
	"net"
	"regexp"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/notfixingit3/echostate/internal/config"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	oteltrace "go.opentelemetry.io/otel/trace"
	collectortrace "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	"google.golang.org/grpc"
)

func TestSetupOTelWithoutOTLPEndpoint(t *testing.T) {
	clearOTelEnv(t)
	restoreOTelGlobals(t)

	tp, err := SetupOTel(context.Background(), &config.Config{Env: "development"})
	require.NoError(t, err)
	require.NotNil(t, tp)
	t.Cleanup(func() { require.NoError(t, tp.Shutdown(context.Background())) })

	_, span := Tracer().Start(context.Background(), "no-exporter")
	span.End()
}

func TestSetupOTelWithOTLPEndpointExportsByGRPC(t *testing.T) {
	clearOTelEnv(t)
	restoreOTelGlobals(t)

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	server := grpc.NewServer()
	service := &recordingTraceService{}
	collectortrace.RegisterTraceServiceServer(server, service)
	t.Cleanup(server.Stop)

	serveErr := make(chan error, 1)
	go func() {
		serveErr <- server.Serve(lis)
	}()

	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", lis.Addr().String())
	t.Setenv("OTEL_TRACES_SAMPLER", "always_on")

	tp, err := SetupOTel(context.Background(), &config.Config{Env: "production"})
	require.NoError(t, err)
	require.NotNil(t, tp)
	t.Cleanup(func() { require.NoError(t, tp.Shutdown(context.Background())) })

	_, span := Tracer().Start(context.Background(), "grpc-export")
	span.End()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	require.NoError(t, tp.ForceFlush(ctx))
	require.Eventually(t, func() bool {
		return service.exportCount.Load() > 0
	}, 2*time.Second, 20*time.Millisecond)
}

func TestExtractTraceparent(t *testing.T) {
	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	t.Cleanup(func() { require.NoError(t, tp.Shutdown(context.Background())) })

	ctx, span := tp.Tracer("test").Start(context.Background(), "root")
	defer span.End()

	traceparent := ExtractTraceparent(ctx)
	require.Regexp(t, regexp.MustCompile(`^00-[0-9a-f]{32}-[0-9a-f]{16}-01$`), traceparent)
	require.Contains(t, traceparent, span.SpanContext().TraceID().String())
	require.Contains(t, traceparent, span.SpanContext().SpanID().String())
	require.Empty(t, ExtractTraceparent(context.Background()))
}

func TestContextFromTraceparent(t *testing.T) {
	const traceparent = "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"

	ctx := ContextFromTraceparent(traceparent)
	sc := oteltrace.SpanContextFromContext(ctx)

	traceID, err := oteltrace.TraceIDFromHex("4bf92f3577b34da6a3ce929d0e0e4736")
	require.NoError(t, err)
	spanID, err := oteltrace.SpanIDFromHex("00f067aa0ba902b7")
	require.NoError(t, err)

	require.True(t, sc.IsValid())
	require.Equal(t, traceID, sc.TraceID())
	require.Equal(t, spanID, sc.SpanID())
	require.True(t, sc.IsSampled())
}

func TestTracerStartWithContextFromTraceparentUsesParent(t *testing.T) {
	clearOTelEnv(t)
	restoreOTelGlobals(t)

	exporter := &memoryExporter{}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithSyncer(exporter),
	)
	t.Cleanup(func() { require.NoError(t, tp.Shutdown(context.Background())) })
	otel.SetTracerProvider(tp)

	const parentTraceparent = "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"
	ctx := ContextFromTraceparent(parentTraceparent)
	_, span := Tracer().Start(ctx, "child")
	span.End()

	spans := exporter.spansSnapshot()
	require.Len(t, spans, 1)

	parent := spans[0].Parent()
	child := spans[0].SpanContext()
	require.Equal(t, oteltrace.SpanContextFromContext(ctx).TraceID(), parent.TraceID())
	require.Equal(t, oteltrace.SpanContextFromContext(ctx).SpanID(), parent.SpanID())
	require.Equal(t, parent.TraceID(), child.TraceID())
	require.NotEqual(t, parent.SpanID(), child.SpanID())
}

func clearOTelEnv(t *testing.T) {
	t.Helper()
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "")
	t.Setenv("OTEL_TRACES_EXPORTER", "")
	t.Setenv("OTEL_TRACES_SAMPLER", "")
	t.Setenv("OTEL_TRACES_SAMPLER_ARG", "")
}

func restoreOTelGlobals(t *testing.T) {
	t.Helper()
	previousProvider := otel.GetTracerProvider()
	previousPropagator := otel.GetTextMapPropagator()
	t.Cleanup(func() {
		otel.SetTracerProvider(previousProvider)
		otel.SetTextMapPropagator(previousPropagator)
	})
}

type recordingTraceService struct {
	collectortrace.UnimplementedTraceServiceServer
	exportCount atomic.Int64
}

func (s *recordingTraceService) Export(context.Context, *collectortrace.ExportTraceServiceRequest) (*collectortrace.ExportTraceServiceResponse, error) {
	s.exportCount.Add(1)
	return &collectortrace.ExportTraceServiceResponse{}, nil
}

type memoryExporter struct {
	mu    sync.Mutex
	spans []sdktrace.ReadOnlySpan
}

func (e *memoryExporter) ExportSpans(_ context.Context, spans []sdktrace.ReadOnlySpan) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.spans = append(e.spans, spans...)
	return nil
}

func (e *memoryExporter) Shutdown(context.Context) error {
	return nil
}

func (e *memoryExporter) spansSnapshot() []sdktrace.ReadOnlySpan {
	e.mu.Lock()
	defer e.mu.Unlock()
	return append([]sdktrace.ReadOnlySpan(nil), e.spans...)
}

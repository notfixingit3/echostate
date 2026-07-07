// Package trace configures OpenTelemetry tracing for EchoState and provides
// W3C traceparent helpers for queue-backed work.
package trace

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/notfixingit3/echostate/internal/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	oteltrace "go.opentelemetry.io/otel/trace"
)

const (
	tracerName        = "echostate"
	connectionTimeout = 5 * time.Second
	exportTimeout     = 10 * time.Second

	defaultProductionSampleRate  = 0.1
	defaultDevelopmentSampleRate = 1.0
)

// SetupOTel configures the global OpenTelemetry tracer provider and W3C Trace
// Context propagator. If no OTLP endpoint is configured, spans are not exported
// unless OTEL_TRACES_EXPORTER=stdout is explicitly set.
func SetupOTel(ctx context.Context, cfg *config.Config) (*sdktrace.TracerProvider, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	sampler, err := samplerFromConfig(cfg)
	if err != nil {
		return nil, err
	}

	options := []sdktrace.TracerProviderOption{
		sdktrace.WithSampler(sampler),
	}

	endpoint := configString(cfg, "OTEL_EXPORTER_OTLP_ENDPOINT",
		"OTelEndpoint",
		"OTELExporterOTLPEndpoint",
		"OTelExporterOTLPEndpoint",
		"OTLPExporterEndpoint",
	)
	exporter := strings.ToLower(configString(cfg, "OTEL_TRACES_EXPORTER",
		"OTELTracesExporter",
		"OTelTracesExporter",
	))

	switch {
	case endpoint != "":
		exp, err := newOTLPExporter(ctx, endpoint)
		if err != nil {
			return nil, err
		}
		options = append(options, sdktrace.WithBatcher(exp, sdktrace.WithExportTimeout(exportTimeout)))
	case exporter == "stdout":
		exp, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
		if err != nil {
			return nil, err
		}
		options = append(options, sdktrace.WithBatcher(exp, sdktrace.WithExportTimeout(exportTimeout)))
	}

	tp := sdktrace.NewTracerProvider(options...)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return tp, nil
}

// ExtractTraceparent serializes the current span context as a W3C traceparent
// header value. It intentionally ignores tracestate.
func ExtractTraceparent(ctx context.Context) string {
	if !oteltrace.SpanContextFromContext(ctx).IsValid() {
		return ""
	}

	propagator := propagation.TraceContext{}
	carrier := propagation.MapCarrier{}
	propagator.Inject(ctx, carrier)
	if !oteltrace.SpanContextFromContext(propagator.Extract(context.Background(), carrier)).IsValid() {
		return ""
	}
	return carrier.Get("traceparent")
}

// ContextFromTraceparent parses a W3C traceparent header value and returns a
// context containing that span context. Invalid inputs return context.Background().
func ContextFromTraceparent(traceparent string) context.Context {
	sc, ok := parseTraceparent(traceparent)
	if !ok {
		return context.Background()
	}
	return oteltrace.ContextWithSpanContext(context.Background(), sc)
}

// Tracer returns EchoState's application tracer from the global provider.
func Tracer() oteltrace.Tracer {
	return otel.Tracer(tracerName)
}

func newOTLPExporter(ctx context.Context, endpoint string) (sdktrace.SpanExporter, error) {
	grpcEndpoint, insecure, err := normalizeGRPCEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	connectCtx, cancel := context.WithTimeout(ctx, connectionTimeout)
	defer cancel()

	options := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(grpcEndpoint),
		otlptracegrpc.WithTimeout(exportTimeout),
	}
	if insecure {
		options = append(options, otlptracegrpc.WithInsecure())
	}

	return otlptracegrpc.New(connectCtx, options...)
}

func normalizeGRPCEndpoint(endpoint string) (string, bool, error) {
	endpoint = strings.TrimSpace(endpoint)
	if endpoint == "" {
		return "", false, fmt.Errorf("OTEL_EXPORTER_OTLP_ENDPOINT is empty")
	}

	if strings.Contains(endpoint, "://") {
		u, err := url.Parse(endpoint)
		if err != nil {
			return "", false, fmt.Errorf("parse OTEL_EXPORTER_OTLP_ENDPOINT: %w", err)
		}
		if u.Host == "" {
			return "", false, fmt.Errorf("OTEL_EXPORTER_OTLP_ENDPOINT must include host")
		}
		return u.Host, u.Scheme == "http", nil
	}

	return endpoint, true, nil
}

func samplerFromConfig(cfg *config.Config) (sdktrace.Sampler, error) {
	samplerName := strings.ToLower(configString(cfg, "OTEL_TRACES_SAMPLER",
		"OTelSampler",
		"OTELTracesSampler",
		"OTelTracesSampler",
	))
	samplerArg := configString(cfg, "OTEL_TRACES_SAMPLER_ARG",
		"OTelSamplerArg",
		"OTELTracesSamplerArg",
		"OTelTracesSamplerArg",
	)

	rate := defaultSampleRate(cfg)
	var err error

	switch samplerName {
	case "", "traceidratio", "parentbased_traceidratio":
		if samplerArg != "" {
			rate, err = parseSampleRate(samplerArg)
		}
	case "always_on", "parentbased_always_on":
		rate = 1.0
	case "always_off", "parentbased_always_off":
		rate = 0.0
	default:
		return nil, fmt.Errorf("unsupported OTEL_TRACES_SAMPLER %q", samplerName)
	}
	if err != nil {
		return nil, err
	}

	return sdktrace.ParentBased(sdktrace.TraceIDRatioBased(rate)), nil
}

func defaultSampleRate(cfg *config.Config) float64 {
	if cfg != nil && strings.EqualFold(cfg.Env, "production") {
		return defaultProductionSampleRate
	}
	return defaultDevelopmentSampleRate
}

func parseSampleRate(value string) (float64, error) {
	rate, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	if err != nil {
		return 0, fmt.Errorf("parse OTEL_TRACES_SAMPLER_ARG: %w", err)
	}
	if rate < 0 || rate > 1 {
		return 0, fmt.Errorf("OTEL_TRACES_SAMPLER_ARG must be between 0 and 1")
	}
	return rate, nil
}

func configString(cfg *config.Config, env string, fields ...string) string {
	if cfg != nil {
		v := reflect.ValueOf(cfg)
		if v.Kind() == reflect.Pointer && !v.IsNil() {
			v = v.Elem()
		}
		if v.IsValid() && v.Kind() == reflect.Struct {
			for _, name := range fields {
				field := v.FieldByName(name)
				if field.IsValid() && field.Kind() == reflect.String {
					if value := strings.TrimSpace(field.String()); value != "" {
						return value
					}
				}
			}
		}
	}
	return strings.TrimSpace(os.Getenv(env))
}

func parseTraceparent(value string) (oteltrace.SpanContext, bool) {
	parts := strings.Split(strings.TrimSpace(value), "-")
	if len(parts) != 4 || parts[0] != "00" {
		return oteltrace.SpanContext{}, false
	}

	traceID, err := oteltrace.TraceIDFromHex(parts[1])
	if err != nil {
		return oteltrace.SpanContext{}, false
	}
	spanID, err := oteltrace.SpanIDFromHex(parts[2])
	if err != nil {
		return oteltrace.SpanContext{}, false
	}
	flags, err := strconv.ParseUint(parts[3], 16, 8)
	if err != nil || len(parts[3]) != 2 {
		return oteltrace.SpanContext{}, false
	}

	sc := oteltrace.NewSpanContext(oteltrace.SpanContextConfig{
		TraceID:    traceID,
		SpanID:     spanID,
		TraceFlags: oteltrace.TraceFlags(flags),
	})
	if !sc.IsValid() {
		return oteltrace.SpanContext{}, false
	}

	return sc, true
}

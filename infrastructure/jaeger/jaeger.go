package jaeger

import (
	"context"
	"net"
	"strings"
	"time"

	"github.com/funxdofficial/golang-module-jaeger/config"
	"github.com/funxdofficial/golang-module-jaeger/domain"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// Tracer mengimplementasikan domain.Tracer dengan Jaeger (OTLP).
type Tracer struct {
	provider *sdktrace.TracerProvider
	serverIP string
}

// Ensure Tracer implements domain.Tracer.
var _ domain.Tracer = (*Tracer)(nil)

// NewTracer membuat tracer baru dari config.
// Panggil sekali saat inisialisasi aplikasi; gunakan Shutdown saat aplikasi selesai.
func NewTracer(cfg config.Config) (*Tracer, error) {
	serverIP := cfg.ServerIP
	if serverIP == "" {
		serverIP = getOutboundIP()
		if serverIP == "" {
			serverIP = "-"
		}
	}

	var opts []sdktrace.TracerProviderOption
	opts = append(opts, sdktrace.WithResource(resource.NewWithAttributes(
		"", // schema URL (optional)
		attribute.String("service.name", cfg.ServiceName),
		attribute.String("server.ip", serverIP),
	)))

	if cfg.Endpoint != "" {
		expOpts := []otlptracehttp.Option{
			otlptracehttp.WithEndpoint(strings.TrimPrefix(strings.TrimPrefix(cfg.Endpoint, "https://"), "http://")),
		}
		if cfg.Insecure {
			expOpts = append(expOpts, otlptracehttp.WithInsecure())
		}
		exp, err := otlptracehttp.New(context.Background(), expOpts...)
		if err != nil {
			return nil, err
		}
		opts = append(opts, sdktrace.WithBatcher(exp))
	}

	provider := sdktrace.NewTracerProvider(opts...)
	otel.SetTracerProvider(provider)

	return &Tracer{
		provider: provider,
		serverIP: serverIP,
	}, nil
}

// Operation membuat span baru (domain.Span) untuk operasi dengan nama dan tipe interaksi.
func (t *Tracer) Operation(ctx context.Context, interactionName domain.InteractionName, interactionType domain.InteractionTypeName) domain.Span {
	tr := otel.Tracer("github.com/funxdofficial/golang-module-jaeger")
	ctx, otelSpan := tr.Start(ctx, interactionName.Value,
		trace.WithAttributes(
			attribute.String("interaction.type", interactionType.Value),
		))
	s := &span{
		otelSpan:        otelSpan,
		ctx:             ctx,
		interactionName: interactionName.Value,
		interactionType: interactionType.Value,
		startTime:       time.Now(),
	}
	return s
}

// Shutdown mematikan tracer provider. Panggil saat aplikasi exit.
func (t *Tracer) Shutdown(ctx context.Context) error {
	if t.provider != nil {
		return t.provider.Shutdown(ctx)
	}
	return nil
}

func getOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return ""
	}
	defer conn.Close()
	addr := conn.LocalAddr().(*net.UDPAddr)
	return addr.IP.String()
}

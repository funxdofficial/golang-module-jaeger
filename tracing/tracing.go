// Package tracing menyediakan API publik modul Jaeger dengan clean architecture.
// Setiap log menyertakan: Timestamp, LevelId (Info/Error/Warning/Debug), TraceId,
// TransactionId, ServerIP, Message, InteractionName, InteractionTypeName.
package tracing

import (
	"context"
	"sync"

	"github.com/funxdofficial/golang-module-jaeger/config"
	"github.com/funxdofficial/golang-module-jaeger/domain"
	"github.com/funxdofficial/golang-module-jaeger/infrastructure/jaeger"
)

var (
	globalTracer   domain.Tracer
	globalCfg      config.Config
	tracersByService map[string]*jaeger.Tracer
	globalMu       sync.RWMutex
)

// Init menginisialisasi tracer global dengan config. Panggil sekali saat startup (misal di main).
// Jika Endpoint kosong, tracer akan no-op (tidak mengirim ke Jaeger).
// Jika ServiceName kosong dan ServiceNames tidak kosong, dipakai ServiceNames[0] sebagai default.
func Init(cfg config.Config) error {
	globalMu.Lock()
	defer globalMu.Unlock()
	if cfg.ServiceName == "" && len(cfg.ServiceNames) > 0 {
		cfg.ServiceName = cfg.ServiceNames[0]
	}
	globalCfg = cfg
	tracersByService = make(map[string]*jaeger.Tracer)
	t, err := jaeger.NewTracer(cfg)
	if err != nil {
		return err
	}
	globalTracer = t
	return nil
}

// GetTracing mengembalikan tracer global. Aman dipanggil sebelum Init; akan mengembalikan no-op tracer.
// Contoh penggunaan:
//
//	trace := libs.GetTracing().Operation(ctx,
//	    tracing.NewInteractionName("Controller InquiryAccount"),
//	    tracing.NewInteractionTypeType("controllers"))
//	defer trace.Finish()
//	trace.Info("Request", "message")
func GetTracing() domain.Tracer {
	globalMu.RLock()
	t := globalTracer
	globalMu.RUnlock()
	if t == nil {
		return jaeger.NoopTracer()
	}
	return t
}

// GetTracingForService mengembalikan tracer untuk serviceName (muncul sebagai service terpisah di Jaeger).
// Tracer dibuat lazy dan di-cache. Jika ServiceNames di config tidak kosong, hanya nama yang ada di daftar yang diizinkan.
func GetTracingForService(serviceName string) domain.Tracer {
	if serviceName == "" {
		return GetTracing()
	}
	globalMu.RLock()
	cfg := globalCfg
	tracers := tracersByService
	globalMu.RUnlock()
	if tracers == nil {
		return jaeger.NoopTracer()
	}
	globalMu.RLock()
	if t, ok := tracers[serviceName]; ok {
		globalMu.RUnlock()
		return t
	}
	if len(cfg.ServiceNames) > 0 {
		allowed := false
		for _, n := range cfg.ServiceNames {
			if n == serviceName {
				allowed = true
				break
			}
		}
		if !allowed {
			globalMu.RUnlock()
			return GetTracing()
		}
	}
	globalMu.RUnlock()

	globalMu.Lock()
	defer globalMu.Unlock()
	if t, ok := tracersByService[serviceName]; ok {
		return t
	}
	t, err := jaeger.NewTracerForService(globalCfg, serviceName)
	if err != nil {
		return jaeger.NoopTracer()
	}
	tracersByService[serviceName] = t
	return t
}

// NewInteractionName membuat nama interaksi untuk Operation (misal: "Controller InquiryAccount").
func NewInteractionName(value string) domain.InteractionName {
	return domain.NewInteractionName(value)
}

// NewInteractionTypeType membuat tipe interaksi untuk Operation (misal: "controllers").
func NewInteractionTypeType(value string) domain.InteractionTypeName {
	return domain.NewInteractionTypeType(value)
}

// NewInteractionTypeName alias untuk NewInteractionTypeType.
func NewInteractionTypeName(value string) domain.InteractionTypeName {
	return domain.NewInteractionTypeName(value)
}

// Shutdown mematikan tracer global dan semua tracer dari GetTracingForService. Panggil saat aplikasi exit (misal di main defer).
func Shutdown(ctx context.Context) error {
	globalMu.Lock()
	defer globalMu.Unlock()
	var firstErr error
	if t, ok := globalTracer.(*jaeger.Tracer); ok {
		firstErr = t.Shutdown(ctx)
	}
	for _, t := range tracersByService {
		if err := t.Shutdown(ctx); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	tracersByService = nil
	return firstErr
}

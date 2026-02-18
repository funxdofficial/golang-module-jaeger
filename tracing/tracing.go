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
	globalTracer domain.Tracer
	globalMu     sync.RWMutex
)

// Init menginisialisasi tracer global dengan config. Panggil sekali saat startup (misal di main).
// Jika Endpoint kosong, tracer akan no-op (tidak mengirim ke Jaeger).
func Init(cfg config.Config) error {
	globalMu.Lock()
	defer globalMu.Unlock()
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

// Shutdown mematikan tracer global. Panggil saat aplikasi exit (misal di main defer).
func Shutdown(ctx context.Context) error {
	globalMu.Lock()
	defer globalMu.Unlock()
	if t, ok := globalTracer.(*jaeger.Tracer); ok {
		return t.Shutdown(ctx)
	}
	return nil
}

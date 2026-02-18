package domain

import "context"

// Span merepresentasikan satu operasi tracing.
// trace_id dan transaction_id sudah otomatis di tag Jaeger.
// server_ip dan exectime ada di process/span. Setiap log (Info, Error, Warning, Debug) menyertakan
// timestamp, level_id, message, scope, interaction_name, interaction_type_name.
type Span interface {
	// Tag menambah tag kustom pada span (muncul di Jaeger). Contoh: span.Tag("field", "value").
	Tag(key, value string)
	// Info mencatat log level Info. Parameter: scope/label, message.
	Info(scope, message string)
	// Error mencatat log level Error.
	Error(scope, message string)
	// Warning mencatat log level Warning.
	Warning(scope, message string)
	// Debug mencatat log level Debug.
	Debug(scope, message string)
	// Finish mengakhiri span dan mengirim ke Jaeger. Selalu panggil dengan defer.
	Finish()
}

// Tracer interface untuk membuat span (layer use case / port).
type Tracer interface {
	// Operation membuat span baru untuk operasi dengan nama dan tipe interaksi.
	// ctx wajib berisi context yang bisa membawa trace id (otel/jaeger).
	Operation(ctx context.Context, interactionName InteractionName, interactionType InteractionTypeName) Span
}

package jaeger

import (
	"context"

	"github.com/funxdofficial/golang-module-jaeger/domain"
)

// noopSpan span yang tidak melakukan apa-apa (jika tracer belum di-init).
type noopSpan struct{}

var _ domain.Span = (*noopSpan)(nil)

func (n *noopSpan) Tag(key, value string)        {}
func (n *noopSpan) Info(scope, message string)   {}
func (n *noopSpan) Error(scope, message string)  {}
func (n *noopSpan) Warning(scope, message string) {}
func (n *noopSpan) Debug(scope, message string)  {}
func (n *noopSpan) Finish()                      {}

// noopTracer tracer yang mengembalikan noop span.
type noopTracer struct{}

var _ domain.Tracer = (*noopTracer)(nil)

func (n *noopTracer) Operation(ctx context.Context, interactionName domain.InteractionName, interactionType domain.InteractionTypeName) domain.Span {
	return &noopSpan{}
}

// NoopTracer mengembalikan tracer no-op.
func NoopTracer() domain.Tracer {
	return &noopTracer{}
}

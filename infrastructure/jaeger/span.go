package jaeger

import (
	"context"
	"time"

	"github.com/funxdofficial/golang-module-jaeger/domain"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type span struct {
	otelSpan        trace.Span
	ctx             context.Context
	interactionName string
	interactionType string
	startTime       time.Time
	finished        bool
}

// ensure span implements domain.Span.
var _ domain.Span = (*span)(nil)

var jakartaTZ *time.Location

func init() {
	jakartaTZ, _ = time.LoadLocation("Asia/Jakarta")
	if jakartaTZ == nil {
		jakartaTZ = time.UTC
	}
}

func (s *span) addEvent(level domain.LevelId, scope, message string) {
	if s.finished {
		return
	}
	// Format: DD/Mon/YYYY HHMMSS.nanosecond (e.g. 27/Jun/2028 143052.123456789), Asia/Jakarta
	timestamp := time.Now().In(jakartaTZ).Format("02/Jan/2006 150405.999999999")
	s.otelSpan.AddEvent(
		"log",
		trace.WithAttributes(
			attribute.String("timestamp", timestamp),
			attribute.String("level_id", string(level)),
			attribute.String("message", message),
			attribute.String("scope", scope),
			attribute.String("interaction_name", s.interactionName),
			attribute.String("interaction_type_name", s.interactionType),
		),
	)
}

// Tag menambah tag kustom pada span (muncul di Jaeger).
func (s *span) Tag(key, value string) {
	if s.finished {
		return
	}
	s.otelSpan.SetAttributes(attribute.String(key, value))
}

func (s *span) Info(scope, message string) {
	s.addEvent(domain.LevelInfo, scope, message)
}

func (s *span) Error(scope, message string) {
	s.addEvent(domain.LevelError, scope, message)
	// Error already recorded in event with level_id "Error"; SetStatus would require go.opentelemetry.io/otel/codes
}

func (s *span) Warning(scope, message string) {
	s.addEvent(domain.LevelWarning, scope, message)
}

func (s *span) Debug(scope, message string) {
	s.addEvent(domain.LevelDebug, scope, message)
}

func (s *span) Finish() {
	if s.finished {
		return
	}
	s.finished = true
	exectimeMs := time.Since(s.startTime).Milliseconds()
	s.otelSpan.SetAttributes(attribute.Int64("exectime", exectimeMs))
	s.otelSpan.End()
}

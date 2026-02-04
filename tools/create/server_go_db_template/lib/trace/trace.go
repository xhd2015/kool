package trace

import (
	"context"

	"github.com/google/uuid"
)

type contextKey string

const traceIDKey contextKey = "trace_id"

// TraceID represents a unique identifier for a request trace
type TraceID string

// NewTraceID generates a new unique trace ID
func NewTraceID() TraceID {
	return TraceID(uuid.New().String())
}

// WithTraceID adds a trace ID to the context
func WithTraceID(ctx context.Context, traceID TraceID) context.Context {
	return context.WithValue(ctx, traceIDKey, traceID)
}

// GetTraceID retrieves the trace ID from context
// Returns empty string if not found
func GetTraceID(ctx context.Context) TraceID {
	if ctx == nil {
		return ""
	}
	if traceID, ok := ctx.Value(traceIDKey).(TraceID); ok {
		return traceID
	}
	return ""
}

// String returns the string representation of TraceID
func (t TraceID) String() string {
	return string(t)
}

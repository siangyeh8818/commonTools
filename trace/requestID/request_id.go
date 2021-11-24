package requestID

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc/metadata"
)

type ctxKey string

const (
	ctxKeyXRequestID      ctxKey = "x-request-id"
	metadataXRequestIDKey        = "x-request-id"
)

// FromContext get x-request-id from context
func FromContext(ctx context.Context) string {
	v, ok := ctx.Value(ctxKeyXRequestID).(string)
	if !ok {
		v = NewRequestID()
	}
	return v
}

// ContextWithXRequestID returns a context.Context with given X-Request-Id value.
func ContextWithXRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, ctxKeyXRequestID, requestID)
}

// MetadataXRequestID returns a context.Context with given X-Request-Id value.
func MetadataXRequestID(ctx context.Context, requestID string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, metadataXRequestIDKey, requestID)
}

// FromContext get x-request-id from context
func MetadataFromContext(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return NewRequestID()
	}
	requestIDMeta, ok := md[metadataXRequestIDKey]
	if !ok || len(requestIDMeta) == 0 {
		return NewRequestID()
	}

	return requestIDMeta[0]
}

func NewRequestID() string {
	return uuid.New().String()
}

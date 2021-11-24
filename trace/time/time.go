package time

import (
	"context"
	"strconv"
	"time"

	"google.golang.org/grpc/metadata"
)

type ctxKey string

const (
	ctxKeyXTime     ctxKey = "x-time"
	metadataTimeKey        = "x-time"
)

// SubTimeFromContext get time sub from context
func SubTimeFromContext(ctx context.Context) int64 {
	v, ok := ctx.Value(ctxKeyXTime).(int64)
	if !ok {
		v = NewTime()
		ctx = context.WithValue(ctx, ctxKeyXTime, v)
	}
	milliTime := time.Now().UTC().UnixNano()/1e6 - v
	return milliTime
}

// GetFromContext get time sub from context
func GetFromContext(ctx context.Context) int64 {
	v, ok := ctx.Value(ctxKeyXTime).(int64)
	if !ok {
		v = NewTime()
	}
	return v
}

// ContextWithTime returns a context.Context with given time value.
func ContextWithTime(ctx context.Context, milliTime int64) context.Context {
	return context.WithValue(ctx, ctxKeyXTime, milliTime)
}

// MetadataTime returns a context.Context with given Time value.
func MetadataTime(ctx context.Context, milliTime int64) context.Context {
	return metadata.AppendToOutgoingContext(ctx, metadataTimeKey, strconv.FormatInt(milliTime, 10))
}

// MetadataTimeFromContext get time sub from meta
func MetadataTimeFromContext(ctx context.Context) int64 {
	var milliTime int64
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return NewTime()
	}
	milliTimeMeta, ok := md[metadataTimeKey]
	if !ok || len(milliTimeMeta) == 0 {
		return NewTime()
	} else {
		n, err := strconv.ParseInt(milliTimeMeta[0], 10, 64)
		if err != nil {
			return NewTime()
		}
		milliTime = n
	}
	return milliTime
}

// SubMetadataTimeFromContext get time sub from meta
func SubMetadataTimeFromContext(ctx context.Context) int64 {
	var milliTime int64
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		milliTime = NewTime()
	}
	milliTimeMeta, ok := md[metadataTimeKey]
	if !ok || len(milliTimeMeta) == 0 {
		milliTime = NewTime()
	} else {
		n, err := strconv.ParseInt(milliTimeMeta[0], 10, 64)
		if err != nil {
			milliTime = NewTime()
		}
		milliTime = n
	}
	subTime := time.Now().UTC().UnixNano()/1e6 - milliTime

	return subTime
}

func NewTime() int64 {
	return time.Now().UTC().UnixNano() / 1e6
}

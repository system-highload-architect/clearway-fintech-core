package context

import (
	"context"

	"google.golang.org/grpc/metadata"
)

const (
	HeaderCorrelationID = "x-correlation-id"
	HeaderIdempotency   = "x-idempotency-key"
)

// AppendFintechMetadata внедряет Correlation-ID и Idempotency-Key в исходящий gRPC-контекст.
// FIXED: Standardized cross-cutting header propagation without breaking domain interface isolation
func AppendFintechMetadata(ctx context.Context, correlationID, idempotencyKey string) context.Context {
	md := metadata.Pairs(
		HeaderCorrelationID, correlationID,
		HeaderIdempotency, idempotencyKey,
	)
	return metadata.NewOutgoingContext(ctx, md)
}

// ExtractFintechMetadata за O(1) вытаскивает метаданные из входящего gRPC-вызова на стороне сервера
func ExtractFintechMetadata(ctx context.Context) (correlationID, idempotencyKey string) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", ""
	}

	if id := md.Get(HeaderCorrelationID); len(id) > 0 {
		correlationID = id[0]
	}
	if key := md.Get(HeaderIdempotency); len(key) > 0 {
		idempotencyKey = key[0]
	}
	return correlationID, idempotencyKey
}

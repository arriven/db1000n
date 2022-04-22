//go:build encrypted
// +build encrypted

package metrics

import (
	"context"

	"go.uber.org/zap"
)

func serveMetrics(ctx context.Context, logger *zap.Logger, listen string) {
	<-ctx.Done()
}

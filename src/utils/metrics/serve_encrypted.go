//go:build encrypted
// +build encrypted

package metrics

import (
	"context"
)

func serveMetrics(ctx context.Context) {
	<-ctx.Done()
}

//go:build !encrypted
// +build !encrypted

package metrics

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

func serveMetrics(ctx context.Context, logger *zap.Logger, listen string) {
	// We don't expect that rendering metrics should take a lot of time and needs long timeout
	const timeout = 30 * time.Second

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(
		prometheus.DefaultGatherer,
		promhttp.HandlerOpts{
			// Opt into OpenMetrics to support exemplars.
			EnableOpenMetrics: true,
			Timeout:           timeout,
		},
	))

	server := &http.Server{
		Addr:    listen,
		Handler: mux,
	}
	go func(ctx context.Context, server *http.Server) {
		<-ctx.Done()

		if err := server.Shutdown(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Warn("failure shutting down prometheus server", zap.Error(err))
		}
	}(ctx, server)

	logger.Warn("prometheus server", zap.Error(server.ListenAndServe()))
}

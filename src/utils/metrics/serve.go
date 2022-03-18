//go:build !encrypted
// +build !encrypted

package metrics

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func serveMetrics(ctx context.Context) {
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
		Addr:    "0.0.0.0:9090",
		Handler: mux,
	}
	go func(ctx context.Context, server *http.Server) {
		<-ctx.Done()

		if err := server.Shutdown(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Println("failure shutting down prometheus server:", err)
		}
	}(ctx, server)

	if err := server.ListenAndServe(); err != nil {
		log.Fatal()
	}
}

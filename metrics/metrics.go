package metrics

import (
	"log"
)

type Metrics struct {
}

func (m Metrics) TrackDuration(addr string, s int64) {
	log.Printf("[METRIC] %s duration %d s", addr, s)
}

func (m Metrics) TrackStart(instance string, s int64) {
	log.Printf("[METRIC] %s started at %d", instance, s)
}

func (m Metrics) TrackFinish(instance string, s int64) {
	log.Printf("[METRIC] %s finished at %d s", instance, s)
}

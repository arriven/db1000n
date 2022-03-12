package utils

import (
	v1 "github.com/mjpitz/go-ga/client/v1"
	"github.com/mjpitz/go-ga/client/v1/gatypes"
)

var (
	previousTraffic int64
	client          = v1.NewClient("UA-222030361-1", "customUserAgent")
)

// ReportStatistics sends basic usage events to google analytics
func ReportStatistics(traffic int64, clientID string) error {
	delta := traffic - previousTraffic
	previousTraffic = traffic

	return trackEvent(delta, clientID)
}

func trackEvent(traffic int64, clientID string) error {
	const kb = 1024

	return client.SendPost(&gatypes.Payload{
		HitType:                           "event",
		NonInteractionHit:                 true,
		DisableAdvertisingPersonalization: true,
		Users:                             gatypes.Users{ClientID: clientID},
		Event: gatypes.Event{
			EventCategory: "statistics",
			EventAction:   "heartbeat",
			EventValue:    traffic / kb,
		},
	})
}

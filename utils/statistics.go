package utils

import (
	v1 "github.com/mjpitz/go-ga/client/v1"
	"github.com/mjpitz/go-ga/client/v1/gatypes"
)

var previousTraffic int64 = 0
var client = v1.NewClient("UA-222030361-1", "customUserAgent")

func ReportStatistics(traffic int64, clientId string) error {
	delta := traffic - previousTraffic
	previousTraffic = traffic
	return trackEvent(delta, clientId)
}

func trackEvent(traffic int64, clientId string) error {
	users := gatypes.Users{ClientID: clientId}
	ping := &gatypes.Payload{
		HitType:                           "event",
		NonInteractionHit:                 true,
		DisableAdvertisingPersonalization: true,
		Users:                             users,
		Event: gatypes.Event{
			EventCategory: "statistics",
			EventAction:   "heartbeat",
			EventValue:    traffic / 1024,
		},
	}

	error := client.SendPost(ping)
	return error
}
